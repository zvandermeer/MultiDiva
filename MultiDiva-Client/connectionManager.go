package main

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
	for ConnectionMenu.connectedToServer {
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
				setCStr(&ConnectionMenu.serverStatus, "Unexpected server closure")
				setCStr(&ConnectionMenu.serverStatusTooltip, "")
				setCStr(&ConnectionMenu.pushNotification, "[NOTICE] Server connection closed unexpectedly!")
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
				setCStr(&ConnectionMenu.serverStatus, "Server was shut down.")
				setCStr(&ConnectionMenu.serverStatusTooltip, "")
				setCStr(&ConnectionMenu.pushNotification, "[NOTICE] Server is being shut down!")
				CloseClient()
				break receivingLoop
			case "loginSuccess":
				//copyGoStr(ConnectionMenu.serverStatus, "Connected to server successfully!", 256)
				setCStr(&ConnectionMenu.serverStatusTooltip, "")
				setCStr(&ConnectionMenu.serverStatus, "Connected to server successfully!")
			case "invalidLogin":
				setCStr(&ConnectionMenu.serverStatus, "An unknown login error occurred. Please try to login again.")
				setCStr(&ConnectionMenu.serverStatusTooltip, "")
				fmt.Println("Unknown login error")
			case "versionMismatch":
				MajorServerVersion, _ := strconv.Atoi(dat["MajorServerVersion"].(string))
				MinorServerVersion, _ := strconv.Atoi(dat["MinorServerVersion"].(string))
				if MajorServerVersion > MajorClientVersion {
					setCStr(&ConnectionMenu.serverStatus, "Outdated client. (Hover for more details)")
					myTooltip := "Server version: v" + strconv.Itoa(MajorServerVersion) + "." + strconv.Itoa(MinorServerVersion) + "\n"
					myTooltip += "Client version: v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion) + "\n"
					myTooltip += "Please update client to a compatible version! (v" + strconv.Itoa(MajorServerVersion) + ".x)"
					setCStr(&ConnectionMenu.serverStatusTooltip, myTooltip)
					fmt.Println("Please update client")
					ConnectionMenu.connectedToServer = false
				} else {
					setCStr(&ConnectionMenu.serverStatus, "Outdated server. (Hover for more details)")
					myTooltip := "Server version: v" + strconv.Itoa(MajorServerVersion) + "." + strconv.Itoa(MinorServerVersion) + "\n"
					myTooltip += "Client version: v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion) + "\n"
					myTooltip += "Please downgrade client to a compatible version, (v" + strconv.Itoa(MajorServerVersion) + ".x) or contact the server admin to update the server!"
					setCStr(&ConnectionMenu.serverStatusTooltip, myTooltip)
					fmt.Println("Please update server")
					ConnectionMenu.connectedToServer = false
				}
			case "note":
				score, _ := strconv.Atoi(dat["Score"].(string))
				combo, _ := strconv.Atoi(dat["Combo"].(string))
				ranking, _ := strconv.Atoi(dat["ranking"].(string))

				InGameMenu.scores[ranking].connectedPlayer = true

				InGameMenu.scores[ranking].fullScore = C.int(score)
				InGameMenu.scores[ranking].combo = C.int(combo)

				scoreSlice := splitInt(score)

				for i := 0; i < 7; i++ {
					if i < len(scoreSlice) {
						InGameMenu.scores[ranking].slicedScore[6-i] = C.int(scoreSlice[i])
					} else {
						InGameMenu.scores[ranking].slicedScore[6-i] = 0
					}
				}

				gradeInt, _ := strconv.Atoi(dat["Grade"].(string))

				InGameMenu.scores[ranking].grade = uint32(gradeInt)

				fmt.Println(InGameMenu.scores)
			case "roomConnectionUpdate":
				roomName := dat["RoomName"].(string)
				switch dat["Status"].(string) {
				case "roomNotFound":
					setCStr(&ConnectionMenu.roomStatus, "Room with name \""+roomName+"\" not found!")
				case "roomAlreadyExists":
					setCStr(&ConnectionMenu.roomStatus, "Cannot create room, room with name \""+roomName+"\" already exists!")
				case "connectedToRoom":
					ConnectionMenu.connectedToRoom = true
					setCStr(&ConnectionMenu.roomStatus, "Connected to room "+roomName+" successfully!")
				case "connectedAsLeader":
					ConnectionMenu.connectedToRoom = true
					setCStr(&ConnectionMenu.roomStatus, "Connected to room successfully! You are now the leader of room \""+roomName+"\"")
				case "disconnectedFromRoom":
					ConnectionMenu.connectedToRoom = false
					setCStr(&ConnectionMenu.roomStatus, "Disconnected from room \""+roomName+"\"!")
				case "kickedFromRoom":
					ConnectionMenu.connectedToRoom = false
					setCStr(&ConnectionMenu.pushNotification, "You've been kicked from the room!")
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
	for ConnectionMenu.connectedToServer {
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
	ConnectionMenu.connectedToServer = false
}
