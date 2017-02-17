package UDP

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

type Elev_info struct {
	ID             string
	Words          string
	ElevIP         string
	CurrentFloor   int
	Direction      int
	NumberOrders   int
	LocolOrders    [4]int
	ExternalOrders [4][2]int
	Recipe         bool
	NewOrder       New_order
}

type New_order struct {
	ID       string
	Words    string
	Position [2]int
	ElevIP   string
}

func CheckForError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func GetLocalIP() string {
	addr, _ := net.InterfaceAddrs()
	localIP := strings.Split(addr[1].String(), "/")
	return localIP[0]
}

func Receiving(port string, newMsgChanRecive chan Elev_info) {
	adress, err := net.ResolveUDPAddr("udp", port)
	CheckForError(err)

	connection, err := net.ListenUDP("udp", adress)
	CheckForError(err)

	recieveBuffer := make([]byte, 1024)

	for {
		n, _, err := connection.ReadFromUDP(recieveBuffer)
		CheckForError(err)

		ID := string(recieveBuffer[7:15])

		if ID == "ElevInfo" {

			var newMsg Elev_info

			err = json.Unmarshal(recieveBuffer[:n], &newMsg)
			CheckForError(err)

			newMsgChanRecive <- newMsg
		}
	}
}

func Transmitter(port string, msg Elev_info, newMsgChanTransmit chan Elev_info) {
	//adress, err := net.ResolveUDPAddr("udp", "129.241.187.150"+port) //Skole 3
	//adress, err := net.ResolveUDPAddr("udp", "129.241.187.149"+port) //Skole 2
	//adress, err := net.ResolveUDPAddr("udp", "129.241.187.140"+port) //Skole 1
	//adress, err := net.ResolveUDPAddr("udp", "129.241.187.255"+port) //Skole broadcast
	//adress, err := net.ResolveUDPAddr("udp", "10.22.66.140"+port) //Skole laptop
	//adress, err := net.ResolveUDPAddr("udp", "192.168.1.3"+port) //Hjemme
	adress, err := net.ResolveUDPAddr("udp", "192.168.1.255"+port) //Hjemme broadcast
	CheckForError(err)

	connection, err := net.DialUDP("udp", nil, adress)
	CheckForError(err)

	for {
		select {
		case msg = <-newMsgChanTransmit:
			buf, _ := json.Marshal(msg)
			_, err = connection.Write(buf)
			CheckForError(err)
			time.Sleep(1 * time.Second)
		default:
			buf, _ := json.Marshal(msg)
			_, err = connection.Write(buf)
			CheckForError(err)
			time.Sleep(1 * time.Second)
		}
	}
}
