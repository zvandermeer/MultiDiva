package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
)

var connection net.Conn
var myConfig *dataTypes.ConfigData

func ReceivingThread(receivingChannel *dataTypes.MessageData, sendingChannel *dataTypes.MessageData, connectedToServer *bool) {
	buffer := make([]byte, 1024)
receivingLoop:
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

		var dat map[string]interface{}

		if err := json.Unmarshal(serverMessage, &dat); err != nil {
			panic(err)
		}

		instruction := dat["Instruction"].(string)

		fmt.Println("INSTRUCTION: " + instruction)

		switch instruction {
		case "serverClosing":
			CloseClient(sendingChannel)
			*connectedToServer = false
			break receivingLoop
		case "roomNotFound":
			fmt.Println("Room not found")
		case "invalidLogin":
			fmt.Println("Unknown error")
		case "versionMismatch":
			MajorServerVersion, _ := strconv.Atoi(dat["MajorServerVersion"].(string))
			if MajorServerVersion > MajorClientVersion {
				fmt.Println("Please update client")
			} else {
				fmt.Println("Please update server")
			}
		default:
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
			fmt.Println("Sending: '" + string(incomingData) + "'")
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
		// m :=

		myData, _ := json.Marshal(map[string]string{
			"Instruction":        "login",
			"MajorClientVersion": strconv.Itoa(MajorClientVersion),
			"MinorClientVersion": strconv.Itoa(MinorClientVersion),
			"Username":           config.Username,
		})
		sendingChannel.Store(myData)
		return true
	}
}

func CloseClient(sendingChannel *dataTypes.MessageData) {
	sendingChannel.Store([]byte("{\"Instruction\":\"clientLogout\"}"))
}
