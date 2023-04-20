package main

import "C"
import (
	"fmt"
	"strconv"

	"github.com/ovandermeer/MultiDiva/internal/configManager"
	"github.com/ovandermeer/MultiDiva/internal/connectionManager"
	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
	"github.com/ovandermeer/MultiDiva/internal/scoreManager"
)

const (
	MajorClientVersion = 0
	MinorClientVersion = 1
)

var cfg dataTypes.ConfigData

var sendingData = dataTypes.MessageData{}
var receivingData = dataTypes.MessageData{}

var connectedToServer bool

//export MultiDivaInit
func MultiDivaInit() {
	fmt.Println("[MultiDiva] Welcome to MultiDiva v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion))
	cfg = configManager.LoadConfig()

	connectedToServer = connectionManager.Connect(&cfg, &sendingData, MajorClientVersion, MinorClientVersion)
	if connectedToServer {
		go connectionManager.SendingThread(&sendingData, &connectedToServer)
		go connectionManager.ReceivingThread(&receivingData, &sendingData, &connectedToServer, MajorClientVersion)
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
