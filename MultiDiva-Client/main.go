package main

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/ovandermeer/MultiDiva/internal/configManager"
	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
	"os"
	"strconv"
)

const (
	MajorClientVersion = 0
	MinorClientVersion = 1
)

var cfg dataTypes.ConfigData

var connectedToServer bool
var connectedToRoom bool
var pushNotification *C.char
var serverStatus *C.char
var serverStatusTooltip *C.char
var roomStatus *C.char

//export MultiDivaInit
func MultiDivaInit(_pushNotification *C.char, _serverStatus *C.char, _serverStatusTooltip *C.char, _roomStatus *C.char, serverVersion *C.char) (*bool, *bool) {
	C.strncpy((*C.char)(serverVersion), (*C.char)(C.CString(strconv.Itoa(MajorClientVersion)+"."+strconv.Itoa(MinorClientVersion))), 5)
	pushNotification = _pushNotification
	serverStatus = _serverStatus
	serverStatusTooltip = _serverStatusTooltip
	roomStatus = _roomStatus
	fmt.Println("[MultiDiva] Initializing MultiDiva v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion) + "...")
	cfg = configManager.LoadConfig()
	return &connectedToServer, &connectedToRoom
}

//export LeaveServer
func LeaveServer() {
	CloseClient()
	ReceivingChannel <- "logout"
	C.strncpy((*C.char)(serverStatus), (*C.char)(C.CString("Disconnected from server successfully!")), 256)
	C.strncpy((*C.char)(serverStatusTooltip), (*C.char)(C.CString("")), 256)
}

//export ConnectToServer
func ConnectToServer(serverAddress *C.char, serverPort *C.char) {
	// Clear out both channels before new connection
	for len(SendingChannel) > 0 {
		<-SendingChannel
	}
	for len(ReceivingChannel) > 0 {
		<-ReceivingChannel
	}

	connectedToServer = Connect(C.GoString(serverAddress), C.GoString(serverPort))

	if connectedToServer {
		fmt.Println("Connected!")
		go SendingThread()
		go ReceivingThread()
	}
}

//export CreateRoom
func CreateRoom(roomName *C.char, publicRoom bool) {
	m := map[string]string{
		"Instruction":       "createRoom",
		"roomName":          C.GoString(roomName),
		"passwordProtected": "false",
		"publicRoom":        strconv.FormatBool(publicRoom),
	}

	data, _ := json.Marshal(m)

	SendingChannel <- data
}

//export JoinRoom
func JoinRoom(roomName *C.char) {
	m := map[string]string{
		"Instruction": "joinRoom",
		"roomName":    C.GoString(roomName),
	}

	data, _ := json.Marshal(m)

	SendingChannel <- data
}

//export MainLoop
func MainLoop() {
	if connectedToServer {
		go GetFrameScore()
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
	CloseClient()
	ReceivingChannel <- "logout"
}

//export OnScoreTrigger
func OnScoreTrigger() {
	go GetFinalScore()
}

// use for debugging without diva running
func main() {
	MultiDivaInit(C.CString(""), C.CString(""), C.CString(""), C.CString(""), C.CString(""))

	ConnectToServer(C.CString("localhost"), C.CString("9988"))

	switch os.Args[1] {
	case "createRoom":
		CreateRoom(C.CString(os.Args[2]), false)
	case "joinRoom":
		JoinRoom(C.CString(os.Args[2]))
	}

	select {}
}
