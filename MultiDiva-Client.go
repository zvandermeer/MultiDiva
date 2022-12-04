package main

import "C"
import (
	"fmt"

	configManager "github.com/ovandermeer/MultiDiva/managers/configManager"
	connectionManager "github.com/ovandermeer/MultiDiva/managers/connectionManager"
)

//export MultiDivaInit
func MultiDivaInit() {
	fmt.Println("[MultiDiva] Welcome to MultiDiva v0.0.2!")
	cfg := configManager.LoadConfig()
	connectionManager.Connect(cfg)
	fmt.Println("Closing!")
}

//export MainLoop
func MainLoop() {
	go connectionManager.ReceiveScore()
	go connectionManager.SendScore()
}

//export SongUpdate
func SongUpdate(songID C.int, is_practice bool) {
	go fmt.Println("Received")
	go fmt.Println(songID)
	go fmt.Println(is_practice)
}

//export MultiDivaDispose
func MultiDivaDispose() {
	connectionManager.CloseClient()
}

// use for debugging without diva running
func main() {

	MultiDivaInit()
	for {
		MainLoop()
	}
}

// TODO: Revisit transitioning from viper to yaml library
