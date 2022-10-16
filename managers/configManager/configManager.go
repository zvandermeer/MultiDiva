package configManager

import (
	"fmt"
	"os"

	dataTypes "MultiDiva-Core/dataTypes"

	"github.com/spf13/viper"
)

func LoadConfig() (cfg dataTypes.ConfigData) {
	config_path := "MultiDiva-Config"

	viper.SetConfigName(config_path) // config file name without extension
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	config_path = "./" + config_path + ".yml"

	viper.SetDefault("config_version", 1)
	viper.SetDefault("debug", false)
	viper.SetDefault("default_timeout", 5000)
	viper.SetDefault("server_address", "localhost")
	viper.SetDefault("server_port", "9988")
	viper.SetDefault("username", "User")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("[MultiDiva] MultiDiva config does not exist. Attempting to create one..")
			if _, err := os.Create(config_path); err != nil {
				fmt.Println("[MultiDiva] Error creating MultiDiva config:", err.Error(), "\n Continuing with default settings.")
			} else {
				viper.WriteConfigAs(config_path)
				fmt.Println("[MultiDiva] New config created successfully!")
			}
		} else {
			fmt.Println("[MultiDiva] Error reading MultiDiva config:", err.Error(), "\n Continuing with default settings.")
		}
	}

	cfg.Config_Version = viper.GetInt("config_version")
	cfg.Debug = viper.GetBool("debug")
	cfg.Default_timeout = viper.GetInt("default_timeout")
	cfg.Server_address = viper.GetString("server_address")
	cfg.Port = viper.GetString("server_port")
	cfg.Username = viper.GetString("username")

	if cfg.Config_Version < 1 {
		fmt.Println("[MultiDiva] Config file is outdated. Attempting to update it...")
		viper.Set("config_version", 1)
		viper.Set("debug", false)
		if err := viper.WriteConfig(); err != nil {
			fmt.Println("[MultiDiva] Error when updating MultiDiva config file:", err.Error())
		} else {
			fmt.Println("[MultiDiva] Config has been updated!")
		}
	}
	fmt.Printf("[MultiDiva] Server address: %v Server port: %v Username: '%v' Config version: %v\n", cfg.Server_address, cfg.Port, cfg.Username, cfg.Config_Version)
	return
}
