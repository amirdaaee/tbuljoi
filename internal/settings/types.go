package settings

type configType struct {
	PhoneNumber string `env:"PHONE_NUMBER,required"`
	AppID       int    `env:"APP_ID,required"`
	AppHash     string `env:"APP_HASH,required"`
	SessionFile string `env:"SESSION_FILE,required"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"WARNING"`
}
