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

var Connection net.Conn
var SendingChannel = make(chan []byte, 100)
var SendingMutex = dataTypes.MessageData{}
var ReceivingChannel = make(chan string, 100)

func ReceivingThread() {
	fmt.Println("Starting Receiving thread")

	buffer := make([]byte, 1024)
receivingLoop:
	for connectedToServer {
		fmt.Println("Waiting to read connection...")
		mLen, err := Connection.Read(buffer)
		select {
		case channelMessage := <-ReceivingChannel:
			if channelMessage == "logout" {
				fmt.Println("Client logout has been triggered")
				CloseClient()
				break receivingLoop
			}
		default:
			break
		}
		if err != nil {
			fmt.Println("[MultiDiva] Error receiving score from " + cfg.ServerAddress + ":" + cfg.Port + ".")
			if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") || strings.Contains(err.Error(), "EOF") {
				fmt.Println("[MultiDiva] Unexpected server closure.")
				CloseClient()
				break receivingLoop
			}
			if cfg.Debug {
				fmt.Println("[MultiDiva] Error details: ", err.Error())
			}
			break
		}
		serverMessage := buffer[:mLen]
		if cfg.Debug {
			fmt.Println("[MultiDiva] Received: ", string(serverMessage))
		}

		var dat map[string]interface{}

		clientMessageString := string(serverMessage)
		var instructions []string
		instructions = append(instructions, clientMessageString)

		// Sometimes, if the client sends multiple messages too fast, the server reads multiple instructions as one.
		// This cause the json unmarshal to panic, and crash the server. Solution, split the message by closing braces.
		// This isn't the best solution, but it works so
		if strings.Count(clientMessageString, "}") > 1 {
			instructions = strings.Split(clientMessageString, "}")
			instructions = instructions[:len(instructions)-1]
			for i := range instructions {
				instructions[i] += "}"
			}
		}

		for i := range instructions {

			if err := json.Unmarshal([]byte(instructions[i]), &dat); err != nil {
				panic(err)
			}
			instruction := dat["Instruction"].(string)

			fmt.Println("INSTRUCTION: " + instruction)

			switch instruction {
			case "serverClosing":
				CloseClient()
				break receivingLoop
			case "roomNotFound":
				fmt.Println("Room not found") // TODO Show in UI
			case "invalidLogin":
				fmt.Println("Unknown error")
			case "versionMismatch":
				MajorServerVersion, _ := strconv.Atoi(dat["MajorServerVersion"].(string))
				if MajorServerVersion > MajorClientVersion { // TODO Show in UI
					fmt.Println("Please update client")
				} else {
					fmt.Println("Please update server")
				}
			case "note":
				fmt.Println("NOTED")
				fmt.Println(dat)
			default:
				// SendingMutex.Store(serverMessage)
				fmt.Println("Unknown command")
				fmt.Println(dat)
			}
		}
	}

	fmt.Println("Exiting receiving thread")
}

func SendingThread() {
	fmt.Println("Starting sending thread")

	var lastData []byte
	for connectedToServer {
		select {
		case incomingData := <-SendingChannel: // If the channel has been written to, send that. Otherwise, fallback to the mutex
			fmt.Println("ChannelData")
			fmt.Println("Sending: '" + string(incomingData) + "'")

			if _, err := Connection.Write(incomingData); err != nil {
				fmt.Println("[MultiDiva] Error sending data to", cfg.ServerAddress+":"+cfg.Port+", data/score not sent.")
				if cfg.Debug {
					fmt.Println("[MultiDiva]  Error details: ", err)
				}
			}

			if bytes.Equal(incomingData, []byte("{\"Instruction\":\"clientLogout\"}")) {
				connectedToServer = false
			}

		default:
			incomingData := SendingMutex.Get()
			if !bytes.Equal(incomingData, lastData) {
				lastData = incomingData
				fmt.Println("Sending: '" + string(incomingData) + "'")

				if _, err := Connection.Write(incomingData); err != nil {
					fmt.Println("[MultiDiva] Error sending data to", cfg.ServerAddress+":"+cfg.Port+", data/score not sent.")
					if cfg.Debug {
						fmt.Println("[MultiDiva]  Error details: ", err)
					}
				}
			}
		}
	}

	fmt.Println("Exiting sending thread")
}

func Connect(serverAddress string, serverPort string) bool {
	//establish connection
	var err error
	if Connection, err = net.Dial("tcp", serverAddress+":"+serverPort); err != nil {
		fmt.Println("[MultiDiva] Error connecting to MultiDiva server'" + cfg.ServerAddress + ":" + cfg.Port + "', MultiDiva is not active.")
		if cfg.Debug {
			fmt.Println("[MultiDiva] Error details:", err.Error())
		}
		return false
	} else {
		// m :=

		myData, _ := json.Marshal(map[string]string{
			"Instruction":        "login",
			"MajorClientVersion": strconv.Itoa(MajorClientVersion),
			"MinorClientVersion": strconv.Itoa(MinorClientVersion),
			"Username":           cfg.Username,
		})
		SendingChannel <- myData
		return true
	}
}

func CloseClient() {
	SendingChannel <- []byte("{\"Instruction\":\"clientLogout\"}")
	//connectedToServer = false
}
