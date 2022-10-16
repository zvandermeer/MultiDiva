package dataTypes

type ConfigData struct {
	Config_Version  int    `yaml:"config_version"`
	Debug           bool   `yaml:"debug"`
	Default_timeout int    `yaml:"default_timeout"`
	Server_address  string `yaml:"server_address"`
	Port            string `yaml:"server_port"`
	Username        string `yaml:"username"`
}
