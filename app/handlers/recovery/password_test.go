package recovery

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	gs "github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/middleware"
	"github.com/osuAkatsuki/hanayo/app/models"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slog"
	"gopkg.in/mailgun/mailgun-go.v1"
)

const testResetToken = "0123456789abcdef0123456789abcdef0123456789abcdef01"

func TestMain(m *testing.M) {
	if err := os.Chdir("../../.."); err != nil {
		panic(err)
	}
	gin.SetMode(gin.TestMode)
	tu.LoadTemplates("")
	os.Exit(m.Run())
}

func recoveryTestRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock, *bytes.Buffer) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	oldDB, oldLogger, oldAuditLogger := services.DB, log.Logger, slog.Default()
	services.DB = sqlx.NewDb(db, "sqlmock")
	logs := new(bytes.Buffer)
	log.Logger = zerolog.New(logs)
	slog.SetDefault(slog.New(slog.NewJSONHandler(logs, nil)))
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		db.Close()
		services.DB, log.Logger = oldDB, oldLogger
		slog.SetDefault(oldAuditLogger)
	})

	r := gin.New()
	r.Use(middleware.StructuredLogger())
	r.Use(gs.Sessions("test", gs.NewCookieStore([]byte("password-recovery-test-cookie-key"))))
	r.Use(func(c *gin.Context) {
		c.Set("context", models.Context{})
		c.Set("session", gs.Default(c))
	})
	r.POST("/pwreset", PasswordResetPageHandler)
	r.GET("/pwreset/continue", PasswordResetContinuePageHandler)
	r.POST("/pwreset/continue", PasswordResetContinueSubmitHandler)
	return r, mock, logs
}

func resetRequest(method, path string) *http.Request {
	form := url.Values{"username": {"example"}, "k": {testResetToken}, "password": {"a-test-password-8dqP"}}
	if method == http.MethodGet {
		path += "?k=" + testResetToken
	}
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "192.0.2.10:12345"
	return req
}

func expectResetLookup(mock sqlmock.Sqlmock, userID int, privileges common.UserPrivileges) {
	mock.ExpectQuery("SELECT password_recovery.id, users.id, users.username, users.privileges, password_recovery.created_at").
		WithArgs(testResetToken).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "username", "privileges", "created_at"}).
			AddRow(1, userID, "example", privileges, time.Now()))
}

func assertProtectedAudit(t *testing.T, logs *bytes.Buffer, stage string, userID int) {
	t.Helper()
	for _, field := range []string{
		`"stage":"` + stage + `"`, fmt.Sprintf(`"user_id":%d`, userID), `"client_ip":"192.0.2.10"`,
	} {
		if !strings.Contains(logs.String(), field) {
			t.Errorf("protected-account log missing %s: %s", field, logs.String())
		}
	}
	if stage != "request" && !strings.Contains(logs.String(), `"reset_id":1`) {
		t.Error("token use was not tied to the matched reset record")
	}
	if strings.Contains(logs.String(), testResetToken) {
		t.Error("recovery exposed the token in logs")
	}
}

func assertRecoveryRedirect(t *testing.T, response *httptest.ResponseRecorder, location string) {
	t.Helper()
	if response.Code != http.StatusFound || response.Header().Get("Location") != location {
		t.Fatalf("expected normal recovery redirect to %s, got %d: %s", location, response.Code, response.Body.String())
	}
}

func TestProtectedAccountsSeeNormalFormButCannotChangePassword(t *testing.T) {
	accounts := []struct {
		name       string
		id         int
		privileges common.UserPrivileges
	}{
		{"panel access without chat moderation", 1234, common.UserPrivilegeNormal | common.AdminPrivilegeAccessRAP},
		{"moderator", 1234, common.UserPrivilegeNormal | common.AdminPrivilegeChatMod},
		{"tournament staff", 1234, common.UserPrivilegeNormal | common.UserPrivilegeTournamentStaff},
		{"unrecognized privilege", 1234, common.UserPrivilegeNormal | 1<<40},
		{"Aika without staff flags", 999, common.UserPrivilegePublic},
	}
	for _, account := range accounts {
		for _, stage := range []string{"view", "redeem"} {
			t.Run(account.name+"/"+stage, func(t *testing.T) {
				router, mock, logs := recoveryTestRouter(t)
				expectResetLookup(mock, account.id, account.privileges)
				method := http.MethodGet
				if stage == "redeem" {
					method = http.MethodPost
					mock.ExpectBegin()
					mock.ExpectQuery(`SELECT privileges FROM users WHERE id = \? FOR UPDATE`).WithArgs(account.id).
						WillReturnRows(sqlmock.NewRows([]string{"privileges"}).AddRow(account.privileges))
					mock.ExpectExec("UPDATE password_recovery").WithArgs("revoked", nil, 1, sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(0, 1))
					// Committing revocation is allowed; any password write fails the test.
					mock.ExpectCommit()
				}
				response := httptest.NewRecorder()
				router.ServeHTTP(response, resetRequest(method, "/pwreset/continue"))
				if stage == "view" {
					if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `name="password"`) ||
						!strings.Contains(response.Body.String(), testResetToken) {
						t.Fatalf("expected normal reset form, got %d: %s", response.Code, response.Body.String())
					}
				} else {
					assertRecoveryRedirect(t, response, "/login")
				}
				assertProtectedAudit(t, logs, stage, account.id)
			})
		}
	}
}

type capturedResetToken struct{ value string }

func (token *capturedResetToken) Match(value driver.Value) bool {
	token.value, _ = value.(string)
	return len(token.value) == 50
}

func TestProtectedAccountsReceiveRealResetEmails(t *testing.T) {
	for _, account := range []struct {
		name       string
		id         int
		privileges common.UserPrivileges
	}{
		{"staff", 1234, common.UserPrivilegeNormal | common.AdminPrivilegeManageUsers},
		{"locked Aika", 999, common.UserPrivilegePublic},
	} {
		t.Run(account.name, func(t *testing.T) {
			router, mock, logs := recoveryTestRouter(t)
			messages := make(chan string, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseMultipartForm(1 << 20); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				messages <- r.FormValue("html")
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"message":"Queued. Thank you.","id":"test-message"}`)
			}))
			t.Cleanup(server.Close)
			oldMailgun := services.MG
			services.MG = mailgun.NewMailgun("example.invalid", "test-key", "")
			services.MG.SetAPIBase(server.URL)
			t.Cleanup(func() { services.MG = oldMailgun })
			mock.ExpectQuery("SELECT id, username, email, privileges FROM users").WithArgs("example").
				WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "privileges"}).
					AddRow(account.id, "example", "example@example.invalid", account.privileges))
			mock.ExpectBegin()
			mock.ExpectExec("UPDATE password_recovery").WithArgs(account.id, sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(0, 0))
			token := new(capturedResetToken)
			mock.ExpectExec("INSERT INTO password_recovery").WithArgs(account.id, token, sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()
			response := httptest.NewRecorder()
			router.ServeHTTP(response, resetRequest(http.MethodPost, "/pwreset"))
			assertRecoveryRedirect(t, response, "/")
			select {
			case message := <-messages:
				if !strings.Contains(message, "/pwreset/continue?k="+token.value) {
					t.Error("email did not contain the token inserted in the database")
				}
			default:
				t.Error("no reset email was sent")
			}
			assertProtectedAudit(t, logs, "request", account.id)
			if strings.Contains(logs.String(), token.value) {
				t.Error("issued token leaked into logs")
			}
		})
	}
}

func TestPromotionBeforePasswordWriteSilentlyBlocksRecovery(t *testing.T) {
	router, mock, logs := recoveryTestRouter(t)
	expectResetLookup(mock, 1234, common.UserPrivilegeNormal)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT privileges FROM users WHERE id = \? FOR UPDATE`).WithArgs(1234).
		WillReturnRows(sqlmock.NewRows([]string{"privileges"}).
			AddRow(common.UserPrivilegeNormal | common.AdminPrivilegeManageUsers))
	mock.ExpectExec("UPDATE password_recovery").WithArgs("revoked", nil, 1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, resetRequest(http.MethodPost, "/pwreset/continue"))
	assertRecoveryRedirect(t, response, "/login")
	assertProtectedAudit(t, logs, "redeem", 1234)
}

func TestInvalidTokensDoNotLogSuccessfulPossession(t *testing.T) {
	for _, invalid := range []string{"unknown or already revoked", "expired"} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			t.Run(invalid+"/"+method, func(t *testing.T) {
				router, mock, logs := recoveryTestRouter(t)
				query := mock.ExpectQuery("SELECT password_recovery.id, users.id, users.username, users.privileges, password_recovery.created_at").
					WithArgs(testResetToken)
				if invalid == "expired" {
					query.WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "username", "privileges", "created_at"}).
						AddRow(1, 999, "example", common.UserPrivilegePublic, time.Now().Add(-time.Hour)))
				} else {
					query.WillReturnError(sql.ErrNoRows)
				}
				response := httptest.NewRecorder()
				router.ServeHTTP(response, resetRequest(method, "/pwreset/continue"))
				if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "That key could not be found") {
					t.Fatalf("expected invalid-token response, got %d: %s", response.Code, response.Body.String())
				}
				if strings.Contains(logs.String(), `"reset_id"`) || strings.Contains(logs.String(), testResetToken) {
					t.Error("invalid token logged as successful possession or leaked into logs")
				}
			})
		}
	}
}

func TestOrdinaryPlayerCanRedeemPasswordReset(t *testing.T) {
	router, mock, logs := recoveryTestRouter(t)
	privileges := common.UserPrivilegePublic | common.UserPrivilegeNormal |
		common.UserPrivilegeDonor | common.UserPrivilegePremium
	expectResetLookup(mock, 1234, privileges)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT privileges FROM users WHERE id = \? FOR UPDATE`).WithArgs(1234).
		WillReturnRows(sqlmock.NewRows([]string{"privileges"}).AddRow(privileges))
	mock.ExpectExec("UPDATE password_recovery").WithArgs("used", sqlmock.AnyArg(), 1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET password_md5 = \? WHERE id = \?`).WithArgs(sqlmock.AnyArg(), 1234).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, resetRequest(http.MethodPost, "/pwreset/continue"))
	if response.Code != http.StatusFound || response.Header().Get("Location") != "/login" {
		t.Fatalf("expected successful recovery redirect, got %d: %s", response.Code, response.Body.String())
	}
	if strings.Contains(logs.String(), testResetToken) {
		t.Error("recovery token leaked into logs")
	}
}

func TestPermissionRecheckFailureCannotChangePassword(t *testing.T) {
	router, mock, _ := recoveryTestRouter(t)
	expectResetLookup(mock, 1234, common.UserPrivilegeNormal)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT privileges FROM users WHERE id = \? FOR UPDATE`).WithArgs(1234).
		WillReturnError(errors.New("database unavailable"))
	mock.ExpectRollback()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, resetRequest(http.MethodPost, "/pwreset/continue"))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected failure without changing the password, got %d", response.Code)
	}
}
