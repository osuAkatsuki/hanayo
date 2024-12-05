package settings

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func getEnv(key string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		panic("Missing environment variable: " + key)
	}
	return val
}

func strToInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func strToBool(s string) bool {
	val, _ := strconv.ParseBool(s)
	return val
}

type Settings struct {
	APP_PORT          int
	APP_COOKIE_SECRET string
	APP_HANAYO_KEY    string

	APP_DEFAULT_LEADERBOARD_SIZE_SETTING int

	APP_ENV string

	APP_BASE_URL                    string
	PUBLIC_AVATARS_SERVICE_BASE_URL string
	INTERNAL_AKATSUKI_API_BASE_URL  string
	PUBLIC_AKATSUKI_API_BASE_URL    string
	PUBLIC_BANCHO_SERVICE_BASE_URL  string

	INTERNAL_BEATMAPS_SERVICE_BASE_URL string
	PUBLIC_BEATMAPS_SERVICE_BASE_URL   string

	DISCORD_SERVER_URL string
	DISCORD_CLIENT_ID  string

	DB_SCHEME string
	DB_HOST   string
	DB_PORT   int
	DB_USER   string
	DB_PASS   string
	DB_NAME   string

	REDIS_NETWORK_TYPE string
	REDIS_HOST         string
	REDIS_PORT         int
	REDIS_PASS         string
	REDIS_DB           int
	REDIS_USE_SSL      bool

	AWS_REGION            string
	AWS_ACCESS_KEY_ID     string
	AWS_SECRET_ACCESS_KEY string
	AWS_ENDPOINT_URL      string
	AWS_BUCKET_NAME       string

	MAILGUN_DOMAIN     string
	MAILGUN_API_KEY    string
	MAILGUN_PUBLIC_KEY string
	MAILGUN_FROM       string

	RECAPTCHA_SITE_KEY   string
	RECAPTCHA_SECRET_KEY string

	IP_LOOKUP_URL string

	PAYPAL_EMAIL_ADDRESS string

	AMPLITUDE_API_KEY string
}

var settings = Settings{}

func LoadSettings() Settings {
	godotenv.Load()

	settings.APP_PORT = strToInt(getEnv("APP_PORT"))
	settings.APP_COOKIE_SECRET = getEnv("APP_COOKIE_SECRET")
	settings.APP_HANAYO_KEY = getEnv("APP_HANAYO_KEY")

	settings.APP_DEFAULT_LEADERBOARD_SIZE_SETTING = strToInt(getEnv("APP_DEFAULT_LEADERBOARD_SIZE_SETTING"))

	settings.APP_ENV = getEnv("APP_ENV")

	// NOTE: These are used from the *client side*, and must be
	//       set to urls available to the public internet.
	settings.APP_BASE_URL = getEnv("APP_BASE_URL")
	settings.PUBLIC_AVATARS_SERVICE_BASE_URL = getEnv("PUBLIC_AVATARS_SERVICE_BASE_URL")
	settings.INTERNAL_AKATSUKI_API_BASE_URL = getEnv("INTERNAL_AKATSUKI_API_BASE_URL")
	settings.PUBLIC_AKATSUKI_API_BASE_URL = getEnv("PUBLIC_AKATSUKI_API_BASE_URL")
	settings.PUBLIC_BANCHO_SERVICE_BASE_URL = getEnv("PUBLIC_BANCHO_SERVICE_BASE_URL")

	settings.INTERNAL_BEATMAPS_SERVICE_BASE_URL = getEnv("INTERNAL_BEATMAPS_SERVICE_BASE_URL")
	settings.PUBLIC_BEATMAPS_SERVICE_BASE_URL = getEnv("PUBLIC_BEATMAPS_SERVICE_BASE_URL")

	settings.DISCORD_SERVER_URL = getEnv("DISCORD_SERVER_URL")
	settings.DISCORD_CLIENT_ID = getEnv("DISCORD_CLIENT_ID")

	settings.DB_SCHEME = getEnv("DB_SCHEME")
	settings.DB_HOST = getEnv("DB_HOST")
	settings.DB_PORT = strToInt(getEnv("DB_PORT"))
	settings.DB_USER = getEnv("DB_USER")
	settings.DB_PASS = getEnv("DB_PASS")
	settings.DB_NAME = getEnv("DB_NAME")

	settings.REDIS_NETWORK_TYPE = getEnv("REDIS_NETWORK_TYPE")
	settings.REDIS_HOST = getEnv("REDIS_HOST")
	settings.REDIS_PORT = strToInt(getEnv("REDIS_PORT"))
	settings.REDIS_PASS = getEnv("REDIS_PASS")
	settings.REDIS_DB = strToInt(getEnv("REDIS_DB"))
	settings.REDIS_USE_SSL = strToBool(getEnv("REDIS_USE_SSL"))

	settings.AWS_REGION = getEnv("AWS_REGION")
	settings.AWS_ACCESS_KEY_ID = getEnv("AWS_ACCESS_KEY_ID")
	settings.AWS_SECRET_ACCESS_KEY = getEnv("AWS_SECRET_ACCESS_KEY")
	settings.AWS_ENDPOINT_URL = getEnv("AWS_ENDPOINT_URL")
	settings.AWS_BUCKET_NAME = getEnv("AWS_BUCKET_NAME")

	settings.MAILGUN_DOMAIN = getEnv("MAILGUN_DOMAIN")
	settings.MAILGUN_API_KEY = getEnv("MAILGUN_API_KEY")
	settings.MAILGUN_PUBLIC_KEY = getEnv("MAILGUN_PUBLIC_KEY")
	settings.MAILGUN_FROM = getEnv("MAILGUN_FROM")

	settings.RECAPTCHA_SITE_KEY = getEnv("RECAPTCHA_SITE_KEY")
	settings.RECAPTCHA_SECRET_KEY = getEnv("RECAPTCHA_SECRET_KEY")

	settings.IP_LOOKUP_URL = getEnv("IP_LOOKUP_URL")

	settings.PAYPAL_EMAIL_ADDRESS = getEnv("PAYPAL_EMAIL_ADDRESS")

	settings.AMPLITUDE_API_KEY = getEnv("AMPLITUDE_API_KEY")

	return settings
}

func GetSettings() Settings {
	return settings
}
