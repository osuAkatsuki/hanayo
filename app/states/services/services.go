package services

import (
	"github.com/amplitude/analytics-go/amplitude"
	"github.com/jmoiron/sqlx"
	csrf "github.com/osuAkatsuki/hanayo/internal/csrf"
	otp "github.com/osuAkatsuki/otp-service-client-go/client"
	"github.com/thehowl/qsql"
	"gopkg.in/mailgun/mailgun-go.v1"
	"gopkg.in/redis.v5"
)

var (
	ConfigMap map[string]interface{}
	DB        *sqlx.DB
	QB        *qsql.DB
	MG        mailgun.Mailgun
	RD        *redis.Client
	Amplitude amplitude.Client
	CSRF      csrf.CSRF
	OTP       otp.OtpClient
)
