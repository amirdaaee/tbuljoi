package settings

type configType struct {
	PhoneNumber   string `env:"PHONE_NUMBER,required"`
	AppID         int    `env:"APP_ID,required"`
	AppHash       string `env:"APP_HASH,required"`
	SessionFile   string `env:"SESSION_FILE,required"`
	ArchiveChatID int64  `env:"ARCHIVE_CHAT_ID"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"WARNING"`
	MaxFloodWait  int    `env:"MAX_FLOOD_WAIT" envDefault:"30"`
	DeviceName    string `env:"DEVICE_NAME" envDefault:"tbuljoi"`
}
