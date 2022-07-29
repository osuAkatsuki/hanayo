package settings

var (
	Config struct {
		// Essential configuration that must be always checked for every environment.
		ListenTo      string `description:"ip:port from which to take requests."`
		Unix          bool   `description:"Whether ListenTo is an unix socket."`
		DSN           string `description:"MySQL server DSN"`
		RedisEnable   bool
		AvatarURL     string
		BaseURL       string
		API           string
		BanchoAPI     string
		CheesegullAPI string
		APISecret     string
		Offline       bool `description:"If this is true, files will be served from the local server instead of the CDN."`

		MainRippleFolder string `description:"Folder where all the non-go projects are contained, such as old-frontend, lets, ci-system. Used for changelog."`
		AvatarsFolder    string `description:"location folder of avatars, used for placing the avatars from the avatar change page."`
		EnableS3         bool   `description:"Whether to use S3 for Avatars"`
		S3Bucket         string

		CookieSecret string

		RedisMaxConnections int
		RedisNetwork        string
		RedisAddress        string
		RedisPassword       string

		DiscordServer string

		BaseAPIPublic string

		Production int `description:"This is a fake configuration value. All of the following from now on should only really be set in a production environment."`

		MailgunDomain        string
		MailgunPrivateAPIKey string
		MailgunPublicAPIKey  string
		MailgunFrom          string

		RecaptchaSite    string
		RecaptchaPrivate string

		DiscordOAuthID     string
		DiscordOAuthSecret string
		DonorBotURL        string
		DonorBotSecret     string

		CoinbaseAPIKey    string
		CoinbaseAPISecret string

		SentryDSN string

		IP_API string
	}
)
