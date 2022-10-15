package main

import (
	"fmt"
	// "strconv"

	// "unsafe"

	configManager "MultiDiva-Core/managers/configManager"
	connectionManager "MultiDiva-Core/managers/connectionManager"
)

import "C"

//export Init
func Init() {
	fmt.Println("[MultiDiva] Welcome to MultiDiva v0.0.2!")
	cfg := configManager.LoadConfig()
	connectionManager.Connect(cfg)
	fmt.Println("Closing!")
	//connectionManager.ReceiveScore()
}

//export OnFrame
func OnFrame() {
	connectionManager.ReceiveScore()
}

// //export StartReceiverThread
// func StartReceiverThread() {
// 	go connectionManager.ReceiveScore()
// }

// //export SendScore
// func SendScore(score int) {
// 	string_score := strconv.Itoa(score)
// 	connectionManager.SendScore(string_score)
// }

// //export CloseClient
// func CloseClient() {
// 	connectionManager.CloseClient()
// }

// func main() {
// 	if MultiDivaInit() {
// 		StartReceiverThread()
// 	}
// }

func main() {
	Init()
	for {
		OnFrame()
	}
}

// TODO: Have goroutine sitting, waiting on channel. SendScore simply passes number value in, and sends to that channel. Both channel and connection remain part or
// connectionManager.go

// TODO: Revisit transitioning from viper to yaml library