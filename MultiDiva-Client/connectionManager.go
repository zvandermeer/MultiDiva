package main

/*
	#include <string.h>
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
)

var Connection net.Conn
var SendingChannel = make(chan []byte, 100)
var SendingMutex = MessageData{}
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
				break receivingLoop
			}
		default:
			break
		}
		if err != nil {
			fmt.Println("[MultiDiva] Error receiving score from " + cfg.ServerAddress + ":" + cfg.Port + ".")
			if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") || strings.Contains(err.Error(), "EOF") {
				fmt.Println("[MultiDiva] Unexpected server closure.")
				setUIString(serverStatus, "Unexpected server closure", 256)
				setUIString(serverStatusTooltip, "", 256)
				setUIString(pushNotification, "[NOTICE] Server connection closed unexpectedly!", 256)
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
				setUIString(serverStatus, "Server was shut down.", 265)
				setUIString(serverStatusTooltip, "", 256)
				setUIString(pushNotification, "[NOTICE] Server is being shut down!", 256)
				CloseClient()
				break receivingLoop
			case "loginSuccess":
				//setUIString(serverStatus, "Connected to server successfully!", 256)
				setUIString(serverStatusTooltip, "", 256)
				setUIString(serverStatus, "Connected to server successfully!", 256)
			case "invalidLogin":
				setUIString(serverStatus, "An unknown login error occurred. Please try to login again.", 256)
				setUIString(serverStatusTooltip, "", 256)
				fmt.Println("Unknown login error")
			case "versionMismatch":
				MajorServerVersion, _ := strconv.Atoi(dat["MajorServerVersion"].(string))
				MinorServerVersion, _ := strconv.Atoi(dat["MinorServerVersion"].(string))
				if MajorServerVersion > MajorClientVersion {
					setUIString(serverStatus, "Outdated client. (Hover for more details)", 256)
					myTooltip := "Server version: v" + strconv.Itoa(MajorServerVersion) + "." + strconv.Itoa(MinorServerVersion) + "\n"
					myTooltip += "Client version: v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion) + "\n"
					myTooltip += "Please update client to a compatible version! (v" + strconv.Itoa(MajorServerVersion) + ".x)"
					setUIString(serverStatusTooltip, myTooltip, 256)
					fmt.Println("Please update client")
					connectedToServer = false
				} else {
					setUIString(serverStatus, "Outdated server. (Hover for more details)", 256)
					myTooltip := "Server version: v" + strconv.Itoa(MajorServerVersion) + "." + strconv.Itoa(MinorServerVersion) + "\n"
					myTooltip += "Client version: v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion) + "\n"
					myTooltip += "Please downgrade client to a compatible version, (v" + strconv.Itoa(MajorServerVersion) + ".x) or contact the server admin to update the server!"
					setUIString(serverStatusTooltip, myTooltip, 256)
					fmt.Println("Please update server")
					connectedToServer = false
				}
			case "note":
				score, _ := strconv.Atoi(dat["Score"].(string))
				combo, _ := strconv.Atoi(dat["Combo"].(string))
				ranking, _ := strconv.Atoi(dat["ranking"].(string))

				UINoteData[ranking].connectedPlayer = true

				UINoteData[ranking].fullScore = C.int(score)
				UINoteData[ranking].combo = C.int(combo)

				scoreSlice := splitInt(score)

				for i := 0; i < 7; i++ {
					if i < len(scoreSlice) {
						UINoteData[ranking].slicedScore[6-i] = C.int(scoreSlice[i])
					} else {
						UINoteData[ranking].slicedScore[6-i] = 0
					}
				}

				gradeInt, _ := strconv.Atoi(dat["Grade"].(string))

				UINoteData[ranking].grade = uint32(gradeInt)

				fmt.Println(UINoteData)
			case "roomConnectionUpdate":
				roomName := dat["RoomName"].(string)
				switch dat["Status"].(string) {
				case "roomNotFound":
					setUIString(roomStatus, "Room with name \""+roomName+"\" not found!", 256)
				case "roomAlreadyExists":
					setUIString(roomStatus, "Cannot create room, room with name \""+roomName+"\" already exists!", 256)
				case "connectedToRoom":
					connectedToRoom = true
					setUIString(roomStatus, "Connected to room "+roomName+" successfully!", 256)
				case "connectedAsLeader":
					connectedToRoom = true
					setUIString(roomStatus, "Connected to room successfully! You are now the leader of room \""+roomName+"\"", 256)
				case "disconnectedFromRoom":
					connectedToRoom = false
					setUIString(roomStatus, "Disconnected from room \""+roomName+"\"!", 256)
				case "kickedFromRoom":
					connectedToRoom = false
					setUIString(pushNotification, "You've been kicked from the room!", 256)
				}
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
				err := Connection.Close()
				if err != nil {
					panic(err)
				}
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
	fmt.Println("Client logging out...")
	SendingChannel <- []byte("{\"Instruction\":\"clientLogout\"}")
	connectedToServer = false
}
