package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unsafe"

	"github.com/ovandermeer/MultiDiva/internal/configManager"
	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
)

const (
	MajorClientVersion = 0
	MinorClientVersion = 1
)

var cfg dataTypes.ConfigData

var connectedToServer bool

//export MultiDivaInit
func MultiDivaInit() *bool {
	fmt.Println("[MultiDiva] Initializing MultiDiva v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion) + "...")
	cfg = configManager.LoadConfig()
	return &connectedToServer
}

//export LeaveServer
func LeaveServer() {
	CloseClient()
	ReceivingChannel <- "logout"
}

// TODO for some reason can't login again, always sends 'clientLogout' on first instruction

//export ConnectToServer
func ConnectToServer(serverAddress *C.char, serverPort *C.char) {
	for len(SendingChannel) > 0 { // Clear out the channel to make sure the "login" instruction is sent
		<-SendingChannel
	}
	connectedToServer = Connect(C.GoString(serverAddress), C.GoString(serverPort))
	fmt.Println("Past")
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
}

//export OnScoreTrigger
func OnScoreTrigger() {
	go GetFinalScore()
}

//export StringTest
func StringTest(ctitle *C.char, test bool) *C.char {
	// ctitle := C.CString(title)
	fmt.Println(C.GoString(ctitle))
	fmt.Println(strconv.FormatBool(test))
	// defer C.free(unsafe.Pointer(ctitle))

	return C.CString("Hello from go!")
}

// use for debugging without diva running
func main() {
	MultiDivaInit()

	myCString := C.CString("localhost")
	myCString2 := C.CString("9988")

	ConnectToServer(myCString, myCString2)

	C.free(unsafe.Pointer(myCString))
	C.free(unsafe.Pointer(myCString2))

	myCString3 := C.CString("")

	switch os.Args[1] {
	case "createRoom":
		myCString3 = C.CString(os.Args[2])
		CreateRoom(myCString3, false)
	case "joinRoom":
		myCString3 = C.CString(os.Args[2])
		JoinRoom(myCString3)
	}

	C.free(unsafe.Pointer(myCString3))

	select {}
}
