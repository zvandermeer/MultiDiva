package dataTypes

var CurrentConfigVersion int = 1

type ConfigData struct {
	ConfigVersion int    `yaml:"config_version"`
	Debug         bool   `yaml:"debug"`
	ServerAddress string `yaml:"server_address"`
	Port          string `yaml:"server_port"`
	Username      string `yaml:"username"`
}

func NewConfigData() (config ConfigData) {
	config.ConfigVersion = CurrentConfigVersion
	config.Debug = false
	config.ServerAddress = "localhost"
	config.Port = "9988"
	config.Username = "User"
	return
}
