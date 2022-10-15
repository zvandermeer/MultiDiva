package connectionManager

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	dataTypes "MultiDiva-Core/dataTypes"
)

const (
	SERVER_TYPE    = "tcp"
	CLIENT_VERSION = "0.1.0"
)

var clientQuit bool
var connection net.Conn
var cfg dataTypes.ConfigData

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

func SendScore(score string) {
	if _, err := connection.Write([]byte(score)); err != nil {
		fmt.Println("[MultiDiva] Error sending score to", cfg.Server_address+":"+cfg.Port+", score not sent. Error details:", err)
	}
}

func ReceiveScore() {
	//exit := false

	//for !exit {
		//fmt.Println("Test")
		myChannel := make(chan string)
		var serverMessage string
		go listener(myChannel)
		select{
		case <-time.After(7 * time.Millisecond):
			//fmt.Println("fail")
		case serverMessage = <-myChannel:
			fmt.Println("[MultiDiva] Received: ", serverMessage)
		}
		// if serverMessage == "/closePipe" {
		// 	exit = true
		// 	if !clientQuit {
		// 		CloseClient()
		// 	}
		// 	break
		// }

	//}
}

func listener(myChannel chan string){
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("[MultiDiva] Error receiving score from :", err.Error())
			os.Exit(1)
		}
	myChannel <- string(buffer[:mLen])
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
