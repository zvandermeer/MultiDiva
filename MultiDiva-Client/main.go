package main

import "C"
import (
	"fmt"
	"strconv"

	"github.com/ovandermeer/MultiDiva/internal/configManager"
	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
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
	fmt.Println("[MultiDiva] Initializing MultiDiva v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion) + "...")
	cfg = configManager.LoadConfig()

	connectedToServer = Connect(&cfg, &sendingData)
	if connectedToServer {
		go SendingThread(&sendingData, &connectedToServer)
		go ReceivingThread(&receivingData, &sendingData, &connectedToServer)
	}
}

//export MainLoop
func MainLoop() {
	if connectedToServer {
		go GetFrameScore(&sendingData)
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
	CloseClient(&sendingData)
}

//export OnScoreTrigger
func OnScoreTrigger() {
	go GetFinalScore(&cfg, &sendingData)
}

// use for debugging without diva running
func main() {
	MultiDivaInit()
	for {

	}
}
