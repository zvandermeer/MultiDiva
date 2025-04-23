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

type Client struct {
	Connected             bool
	Connection            net.Conn
	IncomingMessageBuffer chan map[string]interface{}
	OutgoingMessageBuffer chan []byte
}

func NewClient(serverAddress string, serverPort string) *Client {
	c := Client{}
	var err error

	c.Connection, err = net.Dial("tcp", serverAddress+":"+serverPort)
	if err != nil {
		fmt.Println("[MultiDiva] Error connecting to MultiDiva server'" + cfg.ServerAddress + ":" + cfg.Port + "', MultiDiva is not active.")
		if cfg.Debug {
			fmt.Println("[MultiDiva] Error details:", err.Error())
		}
		return nil
	}

	c.IncomingMessageBuffer = make(chan map[string]interface{}, 20)
	c.OutgoingMessageBuffer = make(chan []byte, 20)

	myData := map[string]string{
		"Instruction":        "login",
		"MajorClientVersion": strconv.Itoa(MajorClientVersion),
		"MinorClientVersion": strconv.Itoa(MinorClientVersion),
		"Username":           cfg.Username,
	}
	
	c.sendJsonMessage(myData)

	go c.listen()
	go c.write()
	go c.processPackets()

	return &c
}

func (c Client) listen() {
	fmt.Println("Starting listening thread")

	buffer := make([]byte, 1024)

	for {
		messageLen, err := c.Connection.Read(buffer)

		if err != nil {
			fmt.Println("[MultiDiva] Error receiving score from " + cfg.ServerAddress + ":" + cfg.Port + ".")
			if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") || strings.Contains(err.Error(), "EOF") {
				fmt.Println("[MultiDiva] Unexpected server closure.")
				setCStr(&ConnectionMenu.serverStatus, "Unexpected server closure")
				setCStr(&ConnectionMenu.serverStatusTooltip, "")
				setCStr(&ConnectionMenu.pushNotification, "[NOTICE] Server connection closed unexpectedly!")
				c.close()
			}
			if cfg.Debug {
				fmt.Println("[MultiDiva] Error details: ", err.Error())
			}
			break
		}

		serverMessageBytes := buffer[:messageLen]
		serverMessageStr := string(serverMessageBytes)

		if strings.Count(serverMessageStr, "\n") > 1 {
			messages := strings.Split(serverMessageStr, "\n")

			for _, message := range messages {
				c.processJSON([]byte(message))
			}
		} else {
			c.processJSON(serverMessageBytes)
		}
	}
}

func (c Client) processJSON(jsonMessage []byte) {
	var data map[string]interface{}

	if err := json.Unmarshal(jsonMessage, &data); err != nil {
		panic(err)
	} else {
		c.IncomingMessageBuffer <- data
	}
}

func (c Client) processPackets() {

processingLoop:
	for {
		data := <-c.IncomingMessageBuffer

		instruction := data["Instruction"].(string)

		fmt.Println("INSTRUCTION: " + instruction)

		switch instruction {

		case "serverClosing":
			setCStr(&ConnectionMenu.serverStatus, "Server was shut down.")
			setCStr(&ConnectionMenu.serverStatusTooltip, "")
			setCStr(&ConnectionMenu.pushNotification, "[NOTICE] Server is being shut down!")
			c.close()
			break processingLoop

		case "clientLogout":
			break processingLoop

		case "loginSuccess":
			setCStr(&ConnectionMenu.serverStatusTooltip, "")
			setCStr(&ConnectionMenu.serverStatus, "Connected to server successfully!")

		case "invalidLogin":
			setCStr(&ConnectionMenu.serverStatus, "An unknown login error occurred. Please try to login again.")
			setCStr(&ConnectionMenu.serverStatusTooltip, "")
			fmt.Println("Unknown login error")

		case "versionMismatch":
			MajorServerVersion, _ := strconv.Atoi(data["MajorServerVersion"].(string))
			MinorServerVersion, _ := strconv.Atoi(data["MinorServerVersion"].(string))

			var outdated string

			if MajorServerVersion > MajorClientVersion {
				outdated = "Client"
			} else {
				outdated = "Server"
			}

			setCStr(&ConnectionMenu.serverStatus, "Outdated "+outdated+". (Hover for more details)")

			myTooltip := "Server version: v" + strconv.Itoa(MajorServerVersion) + "." + strconv.Itoa(MinorServerVersion) + "\n"
			myTooltip += "Client version: v" + strconv.Itoa(MajorClientVersion) + "." + strconv.Itoa(MinorClientVersion) + "\n"
			myTooltip += "Please update " + outdated + " to a compatible version (v" + strconv.Itoa(MajorServerVersion) + ".x)"

			setCStr(&ConnectionMenu.serverStatusTooltip, myTooltip)

			c.close()

		case "note":
			score, _ := strconv.Atoi(data["Score"].(string))
			combo, _ := strconv.Atoi(data["Combo"].(string))
			ranking, _ := strconv.Atoi(data["ranking"].(string))

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

			gradeInt, _ := strconv.Atoi(data["Grade"].(string))

			InGameMenu.scores[ranking].grade = uint32(gradeInt)

			fmt.Println(InGameMenu.scores)

		case "roomConnectionUpdate":
			roomName := data["RoomName"].(string)
			switch data["Status"].(string) {
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
			fmt.Println("Unknown command")
			fmt.Println(data)
		}
	}
}

func (c Client) write() {
	fmt.Println("Starting sending thread")

	for {
		incomingData := <-c.OutgoingMessageBuffer

		if _, err := c.Connection.Write(incomingData); err != nil {
			fmt.Println("[MultiDiva] Error sending data to", cfg.ServerAddress+":"+cfg.Port+", data/score not sent.")
			if cfg.Debug {
				fmt.Println("[MultiDiva]  Error details: ", err)
			}
		}

		if bytes.Equal(incomingData, []byte("{\"Instruction\":\"clientLogout\"}")) {
			err := c.Connection.Close()
			if err != nil {
				panic(err)
			}
			break
		}
	}

	fmt.Println("Exiting sending thread")
}

func (c Client) close() {
	fmt.Println("Client logging out...")
	c.sendSimpleInstruction("clientLogout")

	m := map[string]interface{}{
		"Instruction": "clientLogout",
	}

	c.IncomingMessageBuffer <- m

	c.Connected = false
	ConnectionMenu.connectedToServer = false
}

func (c Client) sendSimpleInstruction(instruction string) {
	m := map[string]string{
		"Instruction": instruction,
	}

	c.sendJsonMessage(m)
}

func (c Client) sendJsonMessage(message map[string]string) {
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
	}

	c.OutgoingMessageBuffer <- data
}
