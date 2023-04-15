package connectionManager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
)

const (
	CLIENT_VERSION = "0.1.0"
)

var connection net.Conn
var myConfig *dataTypes.ConfigData

func ReceivingThread(receivingChannel *dataTypes.MessageData, sendingChannel *dataTypes.MessageData, connectedToServer *bool) {
	buffer := make([]byte, 1024)
	for *connectedToServer {
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("[MultiDiva] Error receiving score from " + myConfig.ServerAddress + ":" + myConfig.Port + ".")
			if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") || strings.Contains(err.Error(), "EOF") {
				fmt.Println("[MultiDiva] Unexpected server closure.")
			}
			if myConfig.Debug {
				fmt.Println("[MultiDiva] Error details: ", err.Error())
			}
			break
		}
		serverMessage := buffer[:mLen]
		if myConfig.Debug {
			fmt.Println("[MultiDiva] Received: ", string(serverMessage))
		}

		if string(serverMessage) == "{\"Instruction\":\"serverClosing\"}" {
			CloseClient(sendingChannel)
			*connectedToServer = false
			break
		} else {
			receivingChannel.Store(serverMessage)
		}
	}
}

func SendingThread(sendingChannel *dataTypes.MessageData, connectedToServer *bool) {
	var lastData []byte
	for *connectedToServer {
		incomingData := sendingChannel.Get()
		if !bytes.Equal(incomingData, lastData) {
			lastData = incomingData
			if _, err := connection.Write(incomingData); err != nil {
				fmt.Println("[MultiDiva] Error sending data to", myConfig.ServerAddress+":"+myConfig.Port+", data/score not sent.")
				if myConfig.Debug {
					fmt.Println("[MultiDiva]  Error details: ", err)
				}
			}
		}
	}
}

func Connect(config *dataTypes.ConfigData, sendingChannel *dataTypes.MessageData) bool {
	myConfig = config
	//establish connection
	var err error
	if connection, err = net.Dial("tcp", myConfig.ServerAddress+":"+myConfig.Port); err != nil {
		fmt.Println("[MultiDiva] Error connecting to MultiDiva server'" + myConfig.ServerAddress + ":" + myConfig.Port + "', MultiDiva is not active.")
		if myConfig.Debug {
			fmt.Println("[MultiDiva] Error details:", err.Error())
		}
		return false
	} else {
		myData, _ := json.Marshal(dataTypes.Handshake{
			Instruction:   "handshake",
			ClientVersion: CLIENT_VERSION,
			Username:      config.Username,
		})
		sendingChannel.Store(myData)
		return true
	}
}

func CloseClient(sendingChannel *dataTypes.MessageData) {
	sendingChannel.Store([]byte("{\"Instruction\":\"clientLogout\"}"))
}
