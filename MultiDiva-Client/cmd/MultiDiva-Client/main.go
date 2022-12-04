package main

import "C"
import (
	"fmt"

	"github.com/ovandermeer/MultiDiva/internal/configManager"
	"github.com/ovandermeer/MultiDiva/internal/connectionManager"
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
func SongUpdate(songID C.int, isPractice bool) {
	go fmt.Println("Received")
	go fmt.Println(songID)
	go fmt.Println(isPractice)
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