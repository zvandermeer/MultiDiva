package main

import "C"
import (
	"fmt"

	"github.com/ovandermeer/MultiDiva/internal/configManager"
	"github.com/ovandermeer/MultiDiva/internal/connectionManager"
	"github.com/ovandermeer/MultiDiva/internal/scoreManager"
)

var s scoreManager.ScoreData

//export MultiDivaInit
func MultiDivaInit() {
	fmt.Println("[MultiDiva] Welcome to MultiDiva v0.0.1!")
	cfg := configManager.LoadConfig()
	connectionManager.Connect(cfg)
	fmt.Println("Closing!")
}

//export MainLoop
func MainLoop() {
	go connectionManager.ReceiveFromServer()
	go scoreManager.GetFrameScore(&s)
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

//export OnScoreTrigger
func OnScoreTrigger() {
	go scoreManager.GetFinalScore()
}

//export OnNoteHit
func OnNoteHit() {
	go fmt.Println("Note Hit!")
	go scoreManager.NoteHit(&s)
}

// use for debugging without diva running
func main() {
	MultiDivaInit()
	for {
		MainLoop()
	}
}
