package services

import (
	"github.com/jmoiron/sqlx"
	csrf "github.com/osuAkatsuki/hanayo/internal/csrf"
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
	CSRF      csrf.CSRF
)
