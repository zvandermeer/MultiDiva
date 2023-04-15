package main

import "C"
import (
	"fmt"

	"github.com/ovandermeer/MultiDiva/internal/configManager"
	"github.com/ovandermeer/MultiDiva/internal/connectionManager"
	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
	"github.com/ovandermeer/MultiDiva/internal/scoreManager"
)

var cfg dataTypes.ConfigData

var sendingData = dataTypes.MessageData{}
var receivingData = dataTypes.MessageData{}

var connectedToServer bool

//export MultiDivaInit
func MultiDivaInit() {
	fmt.Println("[MultiDiva] Welcome to MultiDiva v0.0.1!")
	cfg = configManager.LoadConfig()

	connectedToServer = connectionManager.Connect(&cfg, &sendingData)
	if connectedToServer {
		go connectionManager.SendingThread(&sendingData, &connectedToServer)
		go connectionManager.ReceivingThread(&receivingData, &sendingData, &connectedToServer)
	}
}

//export MainLoop
func MainLoop() {
	if connectedToServer {
		go scoreManager.GetFrameScore(&sendingData)
	}
}

//export SongUpdate
func SongUpdate(songID C.int, isPractice bool) {
	go fmt.Println("Received")
	go fmt.Println(songID)
	go fmt.Println(isPractice)
}

//export MultiDivaDispose
func MultiDivaDispose() {
	connectionManager.CloseClient(&sendingData)
}

//export OnScoreTrigger
func OnScoreTrigger() {
	go scoreManager.GetFinalScore(&cfg, &sendingData)
}

// use for debugging without diva running
func main() {
	MultiDivaInit()
	for {

	}
}
