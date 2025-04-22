package main

/*
#include <string.h>
#include <stdbool.h>
#include <stdlib.h>
#include <stdint.h>

enum NoteGrade {
	Cool = 0,
	Good = 1,
	Safe = 2,
	Cool_Wrong = 3,
	Good_Wrong = 4
};

struct UIPlayerScore {
	bool connectedPlayer;
	int32_t fullScore;
	int32_t slicedScore[7];
	int32_t combo;
	enum NoteGrade grade;
};

struct InGameMenu {
	bool menuVisible;
	struct UIPlayerScore* scores[10];
};

struct EndgameMenu {
	bool menuVisible;
};

struct ConnectionMenu {
	bool menuVisible;
	bool connectedToServer;
	bool connectedToRoom;

	char* pushNotification;
	char* serverStatus;
	char* serverStatusTooltip;
	char* roomStatus;
	char* serverVersion;

	char* serverAddress;
	char* serverPort;
	char* roomName;
};
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unsafe"
)

const (
	MajorClientVersion = 0
	MinorClientVersion = 1
)

var cfg ConfigData

func setCStr(stringToSet **C.char, newString string) {
	newCString := C.CString(newString)
	oldCString := *stringToSet
	C.free(unsafe.Pointer(oldCString))
	*stringToSet = newCString
}

var ConnectionMenu *C.struct_ConnectionMenu
var InGameMenu *C.struct_InGameMenu
var EndgameMenu *C.struct_EndgameMenu

//export MultiDivaInit
func MultiDivaInit(_connectionMenu *C.struct_ConnectionMenu, _ingameMenu *C.struct_InGameMenu, _endgameMenu *C.struct_EndgameMenu) {
	fmt.Println("[MultiDiva] Initializing MultiDiva v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion) + "...")

	ConnectionMenu = _connectionMenu
	InGameMenu = _ingameMenu
	EndgameMenu = _endgameMenu

	versionString := strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion)

	setCStr(&ConnectionMenu.serverVersion, versionString)

	cfg = LoadConfig()
}

//export LeaveServer
func LeaveServer() {
	CloseClient()
	ReceivingChannel <- "logout"
	setCStr(&ConnectionMenu.serverStatus, "Disconnected from server successfully!")
	setCStr(&ConnectionMenu.serverStatusTooltip, "")
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

	ConnectionMenu.connectedToServer = C.bool(Connect(C.GoString(serverAddress), C.GoString(serverPort)))

	if ConnectionMenu.connectedToServer {
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
	if ConnectionMenu.connectedToServer {
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

//export SongRunning
func SongRunning() {
	fmt.Println("Pressing key")
}

// use for debugging without diva running
func main() {
	//var myData []NoteData
	//MultiDivaInit(C.CString(""), C.CString(""), C.CString(""), C.CString(""), C.CString(""), myData)

	ConnectToServer(C.CString("localhost"), C.CString("9988"))

	switch os.Args[1] {
	case "createRoom":
		CreateRoom(C.CString(os.Args[2]), false)
	case "joinRoom":
		JoinRoom(C.CString(os.Args[2]))
	}

	select {}
}
