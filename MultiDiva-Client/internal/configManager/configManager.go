package configManager

import (
	"fmt"
	"os"

	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
	"gopkg.in/yaml.v3"
)

var ConfigLocation string = "./MultiDiva-Config.yml"

func LoadConfig() (cfg dataTypes.ConfigData) {
	if _, err := os.Stat(ConfigLocation); os.IsNotExist(err) {
		writeConfig(dataTypes.NewConfigData())
	}

	cfg = readConfig()

	if cfg.ConfigVersion < dataTypes.CurrentConfigVersion {
		cfg.ConfigVersion = dataTypes.CurrentConfigVersion
		writeConfig(cfg)
	}

	return
}

func readConfig() (myConfig dataTypes.ConfigData) {
	myConfig = dataTypes.NewConfigData()

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

func writeConfig(data dataTypes.ConfigData) {
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
