package connectionManager

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"unsafe"

	"github.com/ovandermeer/MultiDiva/internal/dataTypes"
)

const (
	SERVER_TYPE    = "tcp"
	CLIENT_VERSION = "0.1.0"
)

var connection net.Conn
var myConfig dataTypes.ConfigData
var listening bool
var sending bool
var serverConnected bool
var InterruptSignal chan os.Signal
var oldScore int

func Connect(config dataTypes.ConfigData) {
	myConfig = config
	//establish connection
	var err error
	if connection, err = net.Dial(SERVER_TYPE, myConfig.ServerAddress+":"+myConfig.Port); err != nil {
		fmt.Print("[MultiDiva] Error connecting to MultiDiva server'" + myConfig.ServerAddress + ":" + myConfig.Port + "', MultiDiva is not active.")
		if myConfig.Debug {
			fmt.Println(" Error details:", err.Error())
		} else {
			fmt.Print("\n")
		}
		serverConnected = false
	} else {
		InterruptSignal = make(chan os.Signal, 1)
		signal.Notify(InterruptSignal, os.Interrupt)
		go func() {
			// update this to hook in with diva quitting
			for range InterruptSignal {
				serverConnected = false
				CloseClient()
			}
		}()
		serverConnected = true
	}
	sending = false
	listening = false
}

func SendScore() {
	if serverConnected {
		if !sending {
			sending = true
			location := uintptr(0x1412EF56C)
			p := unsafe.Pointer(location)
			score := *((*int)(p))
			if score != oldScore {
				oldScore = score
				score = score - 4294967296
				scoreString := strconv.Itoa(score)
				fmt.Println("Score: " + scoreString)
				if _, err := connection.Write([]byte(scoreString)); err != nil {
					fmt.Println("[MultiDiva] Error sending score to", myConfig.ServerAddress+":"+myConfig.Port+", score not sent.")
					if myConfig.Debug {
						fmt.Println("[MultiDiva] Error details: ", err)
					}
				}
			}
			sending = false
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
					fmt.Println("[MultiDiva] Error details:", err.Error())
				} else {
					fmt.Print("\n")
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
