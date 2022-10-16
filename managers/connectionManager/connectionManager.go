package connectionManager

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
	"unsafe"

	dataTypes "MultiDiva-Core/dataTypes"
)

const (
	SERVER_TYPE    = "tcp"
	CLIENT_VERSION = "0.1.0"
)

var clientQuit bool
var connection net.Conn
var cfg dataTypes.ConfigData
var listening bool
var sending bool

func Connect(cfg dataTypes.ConfigData) bool {
	//establish connection
	var err error
	if connection, err = net.Dial(SERVER_TYPE, cfg.Server_address+":"+cfg.Port); err != nil {
		fmt.Println("[MultiDiva] Error connecting to MultiDiva server'"+cfg.Server_address+":"+cfg.Port+"', aborting. Error details:", err.Error())
	} else {
		InterruptSignal := make(chan os.Signal, 1)
		signal.Notify(InterruptSignal, os.Interrupt)
		go func() {
			for range InterruptSignal {
				CloseClient()
			}
		}()
		return true
	}
	return false
}

func SendScore() {
	if !sending {
		sending = true
		location := uintptr(0x1412EF56C)
		p := unsafe.Pointer(location)
		score := *((*int)(p))
		score = score - 4294967296
		scoreString := strconv.Itoa(score)
		fmt.Println("Score: " + scoreString)
		if _, err := connection.Write([]byte(scoreString)); err != nil {
			fmt.Println("[MultiDiva] Error sending score to", cfg.Server_address+":"+cfg.Port+", score not sent. Error details:", err)
		}
		sending = false
	}
}

func ReceiveScore() {
	//exit := false

	//for !exit {
	//fmt.Println("Test")
	if !listening {
		listening = true
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("[MultiDiva] Error receiving score from :", err.Error())
			os.Exit(1)
		}
		serverMessage := string(buffer[:mLen])
		fmt.Println("[MultiDiva] Received: ", serverMessage)

		listening = false

		// if serverMessage == "/closePipe" {
		// 	exit = true
		// 	if !clientQuit {
		// 		CloseClient()
		// 	}
		// 	break
		// }
	}

	//}
}

func CloseClient() {
	clientQuit = true
	_, err := connection.Write([]byte("/clientLogout"))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
	}
	time.Sleep(10 * time.Millisecond)
	connection.Close()
	fmt.Println("\nGoodbye!")
	os.Exit(0)
}
