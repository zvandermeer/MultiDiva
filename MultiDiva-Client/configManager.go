package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const CurrentConfigVersion int = 1

type ConfigData struct {
	ConfigVersion  int    `yaml:"config_version"`
	Debug          bool   `yaml:"debug"`
	ServerAddress  string `yaml:"server_address"`
	Port           string `yaml:"server_port"`
	SongLimitPatch bool   `yaml:"song_limit_patch_enabled"`
	Username       string `yaml:"username"`
}

func NewConfigData() (config ConfigData) {
	config.ConfigVersion = CurrentConfigVersion
	config.Debug = false
	config.ServerAddress = "localhost"
	config.Port = "9988"
	config.Username = "User"
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

	fmt.Println(cfg)

	return
}

func readConfig() (myConfig ConfigData) {
	myConfig = NewConfigData()

	dat, err := os.ReadFile(ConfigLocation)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(dat, &myConfig)
	if err != nil {
		fmt.Println(err)
	}

	return
}

func writeConfig(data ConfigData) {
	yamlOutput, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	f, err := os.Create(ConfigLocation)
	if err != nil {
		fmt.Println(err)
	}

	_, err = f.Write(yamlOutput)
	if err != nil {
		return
	}
}
