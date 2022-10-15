package dataTypes

type ConfigData struct {
	Config_Version int `yaml:"config_version"`
	Server_address string `yaml:"server_address"`
	Port string `yaml:"server_port"`
	Username string `yaml:"username"`
}