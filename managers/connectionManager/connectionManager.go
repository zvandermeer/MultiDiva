package connectionManager

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
	"unsafe"

	dataTypes "github.com/ovandermeer/MultiDiva/dataTypes"
)

const (
	SERVER_TYPE    = "tcp"
	CLIENT_VERSION = "0.1.0"
)

var connection net.Conn
var cfg dataTypes.ConfigData
var listening bool
var sending bool
var serverConnected bool
var InterruptSignal chan (os.Signal)
var oldScore int

func Connect(config_in dataTypes.ConfigData) {
	cfg = config_in
	//establish connection
	var err error
	if connection, err = net.Dial(SERVER_TYPE, cfg.Server_address+":"+cfg.Port); err != nil {
		fmt.Print("[MultiDiva] Error connecting to MultiDiva server'" + cfg.Server_address + ":" + cfg.Port + "', MultiDiva is not active.")
		if cfg.Debug {
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
					fmt.Print("[MultiDiva] Error sending score to", cfg.Server_address+":"+cfg.Port+", score not sent.")
					if cfg.Debug {
						fmt.Println(" Error details:", err)
					} else {
						fmt.Print("\n")
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
				fmt.Println("[MultiDiva] Error receiving score from " + cfg.Server_address + ":" + cfg.Port + ".")
				if cfg.Debug {
					fmt.Println(" Error details:", err.Error())
				} else {
					fmt.Print("\n")
				}
			}
			serverMessage := string(buffer[:mLen])
			if cfg.Debug {
				fmt.Println("[MultiDiva] Received: ", serverMessage)
			}

			if serverMessage == "/closePipe" {
				if serverConnected {
					serverConnected = false
					CloseClient()
				}
			}
			listening = false
		}
	}
}

func CloseClient() {
	_, err := connection.Write([]byte("/clientLogout"))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
	}
	time.Sleep(10 * time.Millisecond)
	connection.Close()
	fmt.Println("\nGoodbye!")
}
