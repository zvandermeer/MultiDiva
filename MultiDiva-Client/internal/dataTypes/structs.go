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

type DivaScore struct {
	TotalScore     int32
	Unknown1       int32
	Unknown2       int32
	Unknown3       int32
	Unknown4       int32
	Unknown5       int32
	Unknown6       int32
	Unknown7       int32
	Unknown8       int32
	Combo          int32
	PreAdjustCool  int32
	PreAdjustFine  int32
	PreAdjustSafe  int32
	PreAdjustSad   int32
	PreAdjustWorst int32
	Cool           int32
	Fine           int32
	Safe           int32
	Sad            int32
	Worst          int32
}
