package configManager

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var ConfigLocation string = "./MultiDiva-Config.yml"

func LoadConfig() (cfg ConfigData) {
	if _, err := os.Stat(ConfigLocation); os.IsNotExist(err) {
		writeConfig(NewConfigData())
	}

	cfg = readConfig()

	if cfg.ConfigVersion > CurrentConfigVersion {
		cfg.ConfigVersion = CurrentConfigVersion
		writeConfig(cfg)
	}

	return
}

func readConfig() (myConfig ConfigData) {
	myConfig = NewConfigData()

	dat, err := os.ReadFile(ConfigLocation)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(dat, &myConfig)

	fmt.Println(myConfig)

	return
}

func writeConfig(data ConfigData) {
	yamlOutput, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	f, err := os.Create(ConfigLocation)

	_, err = f.Write(yamlOutput)
	if err != nil {
		return
	}
}
