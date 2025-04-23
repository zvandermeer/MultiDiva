package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

const CurrentConfigVersion int = 1

type ConfigData struct {
	ConfigVersion  int    `yaml:"config_version"`
	ServerAddress  string `yaml:"server_address"`
	Port           string `yaml:"server_port"`
	SongLimitPatch bool   `yaml:"song_limit_patch_enabled"`
	Username       string `yaml:"username"`
	LogLevel     int    `yaml:"log_level"`
	LogToFile    bool   `yaml:"log_to_file"`
	LogFilepath  string `yaml:"log_file_path"`
}

func NewConfigData() (config ConfigData) {
	config.ConfigVersion = CurrentConfigVersion
	config.ServerAddress = "localhost"
	config.Port = "9988"
	config.Username = "User"
	config.LogLevel = 0
	config.LogToFile = false
	config.LogFilepath = ""
	return
}

var ConfigLocation string = "./MultiDiva-Config.yml"

func LoadConfig() (cfg ConfigData) {
	if _, err := os.Stat(ConfigLocation); os.IsNotExist(err) {
		writeConfig(NewConfigData())
	}

	cfg = readConfig()

	if cfg.ConfigVersion < CurrentConfigVersion {
		cfg.ConfigVersion = CurrentConfigVersion
		writeConfig(cfg)
	}

	divalog = NewDivaLog(cfg.LogLevel, cfg.LogFilepath)

	divalog.Log(cfg, 2)

	return
}

func readConfig() (myConfig ConfigData) {
	myConfig = NewConfigData()

	dat, err := os.ReadFile(ConfigLocation)
	if err != nil {
		divalog.Log("Error reading config: " + err.Error(), 0)
	}

	err = yaml.Unmarshal(dat, &myConfig)
	if err != nil {
		divalog.Log("[MultiDiva] Error reading config: " + err.Error(), 0)
	}

	return
}

func writeConfig(data ConfigData) {
	yamlOutput, err := yaml.Marshal(data)
	if err != nil {
		divalog.Log("[MultiDiva] Error saving config: " + err.Error(), 0)
	}

	f, err := os.Create(ConfigLocation)
	if err != nil {
		divalog.Log("[MultiDiva] Error saving config: " + err.Error(), 0)
	}

	_, err = f.Write(yamlOutput)
	if err != nil {
		divalog.Log("[MultiDiva] Error reading config: " + err.Error(), 0)
	}
}
