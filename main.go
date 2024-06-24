package main

// about using johnniedoe/contrib/gzip:
// johnniedoe's fork fixes a critical issue for which .String resulted in
// an ERR_DECODING_FAILED. This is an actual pull request on the contrib
// repo, but apparently, gin is dead.

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"golang.org/x/exp/slog"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/fatih/structs"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/johnniedoe/contrib/gzip"
	beatmapsHandlers "github.com/osuAkatsuki/hanayo/app/handlers/beatmaps"
	clansHandlers "github.com/osuAkatsuki/hanayo/app/handlers/clans"
	clanCreationHandlers "github.com/osuAkatsuki/hanayo/app/handlers/clans/create"
	errorHandlers "github.com/osuAkatsuki/hanayo/app/handlers/errors"
	ircHandlers "github.com/osuAkatsuki/hanayo/app/handlers/irc"
	miscHandlers "github.com/osuAkatsuki/hanayo/app/handlers/misc"
	multiplayerHandlers "github.com/osuAkatsuki/hanayo/app/handlers/multiplayer"
	profilesHandlers "github.com/osuAkatsuki/hanayo/app/handlers/profiles"
	profileEditHandlers "github.com/osuAkatsuki/hanayo/app/handlers/profiles/settings"
	accountRecoveryHandlers "github.com/osuAkatsuki/hanayo/app/handlers/recovery"
	loginHandlers "github.com/osuAkatsuki/hanayo/app/handlers/sessions/login"
	logoutHandlers "github.com/osuAkatsuki/hanayo/app/handlers/sessions/logout"
	registerHandlers "github.com/osuAkatsuki/hanayo/app/handlers/sessions/register"
	logging "github.com/osuAkatsuki/hanayo/app/logging"
	middleware "github.com/osuAkatsuki/hanayo/app/middleware"
	"github.com/osuAkatsuki/hanayo/app/middleware/pagemappings"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	sessionsmanager "github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
	"github.com/osuAkatsuki/hanayo/app/version"
	"github.com/osuAkatsuki/hanayo/internal/btcconversions"
	"github.com/osuAkatsuki/hanayo/internal/csrf/cieca"
	"github.com/thehowl/qsql"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
	"gopkg.in/mailgun/mailgun-go.v1"
	"gopkg.in/redis.v5"
)

var startTime = time.Now()

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	slog.Info("The server is starting up", "version", version.Version)

	settings := settingsState.LoadSettings()

	services.ConfigMap = structs.Map(settings)

	// initialise db
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		settings.DB_USER,
		settings.DB_PASS,
		settings.DB_HOST,
		settings.DB_PORT,
		settings.DB_NAME,
	)

	db, err := sqlx.Open(settings.DB_SCHEME, dsn)
	if err != nil {
		panic(err)
	}
	services.DB = db

	qb := qsql.New(db.DB)
	if err != nil {
		panic(err)
	}
	services.QB = qb

	// if config.EnableS3 {
	// 	sess = session.Must(session.NewSessionWithOptions(session.Options{
	// 		SharedConfigState: session.SharedConfigEnable,
	// 	}))
	// }

	// initialise mailgun
	mg := mailgun.NewMailgun(
		settings.MAILGUN_DOMAIN,
		settings.MAILGUN_API_KEY,
		settings.MAILGUN_PUBLIC_KEY,
	)
	services.MG = mg

	// initialise CSRF service
	services.CSRF = cieca.NewCSRF()

	if gin.Mode() == gin.DebugMode {
		slog.Info("Development environment detected. Starting fsnotify on template folder...")
		err := tu.Reloader()
		if err != nil {
			slog.Error("Failed to start template reload watcher", "error", err.Error())
		}
	}

	var tlsConfig *tls.Config
	if settings.REDIS_USE_SSL {
		tlsConfig = &tls.Config{
			ServerName: "*.c.db.ondigitalocean.com",
		}
	}

	// initialise redis
	rd := redis.NewClient(&redis.Options{
		Addr:      fmt.Sprintf("%s:%d", settings.REDIS_HOST, settings.REDIS_PORT),
		Password:  settings.REDIS_PASS,
		DB:        settings.REDIS_DB,
		TLSConfig: tlsConfig,
	})
	services.RD = rd

	amplitudeConfig := amplitude.NewConfig(settings.AMPLITUDE_API_KEY)
	amplitudeConfig.MinIDLength = 4 // our user ids start from 1000
	amplitudeConfig.Logger = (*logging.AmplitudeLogger)(logger)
	services.Amplitude = amplitude.NewClient(amplitudeConfig)

	// even if it's not release, we say that it's release
	// so that gin doesn't spam
	gin.SetMode(gin.ReleaseMode)

	gobRegisters := []interface{}{
		[]msg.Message{},
		msg.ErrorMessage{},
		msg.InfoMessage{},
		msg.NeutralMessage{},
		msg.WarningMessage{},
		msg.SuccessMessage{},
	}
	for _, el := range gobRegisters {
		gob.Register(el)
	}

	slog.Info("Importing templates")
	tu.LoadTemplates("")

	slog.Info("Setting up rate limiter")
	middleware.SetUpLimiter()

	slog.Info("Intialisation", "time", time.Since(startTime))

	r := generateEngine()
	slog.Info("Listening on port", "port", settings.APP_PORT)

	err = r.Run(fmt.Sprintf(":%d", settings.APP_PORT))
	if err != nil {
		slog.Error("Failed to start server", "error", err.Error())
		panic(err)
	}
}

func generateEngine() *gin.Engine {
	slog.Info("Starting session system")
	settings := settingsState.GetSettings()
	var store sessions.Store
	var err error
	if settings.REDIS_MAX_CONNECTIONS != 0 {
		store, err = sessions.NewRedisStore(
			settings.REDIS_MAX_CONNECTIONS,
			settings.REDIS_NETWORK_TYPE,
			fmt.Sprintf("%s:%d", settings.REDIS_HOST, settings.REDIS_PORT),
			settings.REDIS_PASS,
			[]byte(settings.APP_COOKIE_SECRET),
		)
	} else {
		store = sessions.NewCookieStore([]byte(settings.APP_COOKIE_SECRET))
	}

	if err != nil {
		slog.Error("Failed to crreate redis store", "error", err.Error())
		panic(err)
	}

	// Initalize Gin
	gin.SetMode(gin.ReleaseMode)
	// Instantiate a new engine
	r := gin.New()

	r.Use(
		// Use our custom logger
		middleware.StructuredLogger(),
		// Still use the built-in recovery middleware that is called with default
		gin.Recovery(),
		gzip.Gzip(gzip.DefaultCompression),
		pagemappings.CheckRedirect,
		sessions.Sessions("session", store),
		sessionsmanager.SessionInitializer(),
		middleware.RateLimiter(false),
		gintrace.Middleware("hanayo"))

	r.Static("/static", "web/static")
	r.StaticFile("/favicon.ico", "web/static/favicon.ico")

	r.POST("/login", loginHandlers.LoginSubmitHandler)
	r.GET("/logout", logoutHandlers.LogoutSubmitHandler)

	r.GET("/", miscHandlers.HomepagePageHandler)

	r.GET("/register", registerHandlers.RegisterPageHandler)
	r.POST("/register", registerHandlers.RegisterSubmitHandler)
	r.GET("/register/verify", registerHandlers.VerifyAccountPageHandler)
	r.GET("/register/welcome", registerHandlers.WelcomePageHandler)

	r.GET("/clans/create", clanCreationHandlers.ClanCreatePageHandler)
	r.POST("/clans/create", clanCreationHandlers.ClanCreateSubmitHandler)

	r.GET("/u/:user", profilesHandlers.UserProfilePageHandler)
	r.GET("/rx/u/:user", func(c *gin.Context) { // redirect for old links.
		c.Redirect(301, fmt.Sprintf("/u/%s?rx=1", c.Param("user")))
	})

	r.GET("/c/:cid", clansHandlers.ClanPageHandler)
	r.GET("/b/:bid", beatmapsHandlers.BeatmapPageHandler)

	r.GET("/matches/:mid", multiplayerHandlers.MultiplayerHistoryHandler)

	// TODO: maybe change this long names?
	r.POST(
		"/pwreset",
		accountRecoveryHandlers.PasswordResetPageHandler,
	)
	r.GET(
		"/pwreset/continue",
		accountRecoveryHandlers.PasswordResetContinuePageHandler,
	)
	r.POST(
		"/pwreset/continue",
		accountRecoveryHandlers.PasswordResetContinueSubmitHandler,
	)

	r.POST("/irc/generate", ircHandlers.IrcGenTokenSubmitHandler)

	r.GET("/settings/password", profileEditHandlers.ChangePasswordPageHandler)
	r.POST("/settings/password", profileEditHandlers.ChangePasswordSubmitHandler)
	r.POST("/settings/userpage/parse", profileEditHandlers.ParseBBCodeSubmitHandler)
	r.POST("/settings/avatar", profileEditHandlers.AvatarSubmitHandler)
	// r.POST("/settings/flag", profileEditHandlers.FlagChangeSubmitHandler)
	r.POST("/settings/username", profileEditHandlers.NameChangeSubmitHandler)
	//r.GET("/settings/discord/finish", profileEditHandlers.discordFinish)
	r.POST(
		"/settings/profbackground/:type",
		profileEditHandlers.ProfileBackgroundSubmitHandler,
	)

	r.GET("/donate/rates", btcconversions.GetRates)

	r.GET("/about", miscHandlers.AboutPageHandler)

	tu.LoadSimplePages(r)

	r.NoRoute(errorHandlers.NotFoundHandler)

	return r
}

const alwaysRespondText = `Ooops! Looks like something went really wrong while trying to process your request.
Perhaps report this to a Akatsuki developer?
Retrying doing again what you were trying to do might work, too.`
