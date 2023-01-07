package connectionManager

import (
	"fmt"
	"net"
	"strings"

	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
)

const (
	SERVER_TYPE    = "tcp"
	CLIENT_VERSION = "0.1.0"
)

var connection net.Conn
var myConfig dataTypes.ConfigData
var listening bool
var serverConnected bool

func Connect(config dataTypes.ConfigData) {
	myConfig = config
	//establish connection
	var err error
	if connection, err = net.Dial(SERVER_TYPE, myConfig.ServerAddress+":"+myConfig.Port); err != nil {
		fmt.Println("[MultiDiva] Error connecting to MultiDiva server'" + myConfig.ServerAddress + ":" + myConfig.Port + "', MultiDiva is not active.")
		if myConfig.Debug {
			fmt.Println("[MultiDiva] Error details:", err.Error())
		}
		serverConnected = false
	} else {
		serverConnected = true
	}
	listening = false
}

func SendScore(scoreString string) {
	if serverConnected {
		if _, err := connection.Write([]byte(scoreString)); err != nil {
			fmt.Println("[MultiDiva] Error sending score to", myConfig.ServerAddress+":"+myConfig.Port+", score not sent.")
			if myConfig.Debug {
				fmt.Println("[MultiDiva]  Error details: ", err)
			}
		}
	}
}

func ReceiveScore() {
	if serverConnected {
		if !listening {
			listening = true
			buffer := make([]byte, 1024)
			mLen, err := connection.Read(buffer)
			if err != nil {
				fmt.Println("[MultiDiva] Error receiving score from " + myConfig.ServerAddress + ":" + myConfig.Port + ".")
				if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") || strings.Contains(err.Error(), "EOF") {
					fmt.Println("[MultiDiva] Unexpected server closure.")
					serverConnected = false
				}
				if myConfig.Debug {
					fmt.Println("[MultiDiva] Error details: ", err.Error())
				}
			}
			serverMessage := string(buffer[:mLen])
			if myConfig.Debug {
				fmt.Println("[MultiDiva] Received: ", serverMessage)
			}

			if serverMessage == "/closePipe" {
				CloseClient()
			}
			listening = false
		}
	}
}

func CloseClient() {
	serverConnected = false
	_, err := connection.Write([]byte("/clientLogout"))
	if err != nil {
		fmt.Println("[MultiDiva] Error writing:", err.Error())
	}
	fmt.Println("\nGoodbye!")
}
