package main

import (
	"fmt"

	configManager "github.com/ovandermeer/MultiDiva/managers/configManager"
	connectionManager "github.com/ovandermeer/MultiDiva/managers/connectionManager"
)

import "C"

//export Init
func Init() {
	fmt.Println("[MultiDiva] Welcome to MultiDiva v0.0.2!")
	cfg := configManager.LoadConfig()
	connectionManager.Connect(cfg)
	fmt.Println("Closing!")
}

//export OnFrame
func OnFrame() {
	go connectionManager.ReceiveScore()
	go connectionManager.SendScore()
}

// use for debugging without diva running
func main() {
	Init()
	for {
		OnFrame()
	}
}

// TODO: Revisit transitioning from viper to yaml library
