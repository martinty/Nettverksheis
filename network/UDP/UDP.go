package UDP

import (
	"../../source"
	"encoding/json"
	//"fmt"
	//"math/rand"
	"net"
	"strings"
	"time"
)

func GetLocalIP() string {
	addr, err := net.InterfaceAddrs()
	source.CheckForError(err)
	localIP := strings.Split(addr[1].String(), "/")
	return localIP[0]
}

func GetLocalID(IP string) string {
	localID := strings.Split(IP, ".")
	return localID[3]
}

func Receiving(port string, newMsgChanRecive chan source.ElevInfo) {

	adress, err := net.ResolveUDPAddr("udp", port)
	source.CheckForError(err)

	connection, err := net.ListenUDP("udp", adress)
	source.CheckForError(err)

	recieveBuffer := make([]byte, 1024)

	for {
		n, _, err := connection.ReadFromUDP(recieveBuffer)
		source.CheckForError(err)

		var newMsg source.ElevInfo

		err = json.Unmarshal(recieveBuffer[:n], &newMsg)
		source.CheckForError(err)

		newMsgChanRecive <- newMsg
	}
}

func Transmitting(port string, msg source.ElevInfo, newMsgChanTransmit chan source.ElevInfo) {

	//adress, err := net.ResolveUDPAddr("udp", "129.241.187.150"+port) //Skole 3
	//adress, err := net.ResolveUDPAddr("udp", "129.241.187.154"+port) //Skole 7
	//adress, err := net.ResolveUDPAddr("udp", "129.241.187.145"+port) //Skole 17
	//adress, err := net.ResolveUDPAddr("udp", "129.241.187.149"+port) //Skole 2
	//adress, err := net.ResolveUDPAddr("udp", "129.241.187.140"+port) //Skole 1
	adress, err := net.ResolveUDPAddr("udp", "129.241.187.255"+port) //Skole broadcast
	//adress, err := net.ResolveUDPAddr("udp", "10.22.66.140"+port) //Skole laptop
	//adress, err := net.ResolveUDPAddr("udp", "192.168.1.3"+port) //Hjemme
	//adress, err := net.ResolveUDPAddr("udp", "192.168.1.255"+port) //Hjemme broadcast
	source.CheckForError(err)

	connection, err := net.DialUDP("udp", nil, adress)
	source.CheckForError(err)

	for {
		select {
		case msg = <-newMsgChanTransmit:
			buf, _ := json.Marshal(msg)
			_, err = connection.Write(buf)
			source.CheckForError(err)
		default:
			buf, _ := json.Marshal(msg)
			_, err = connection.Write(buf)
			source.CheckForError(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
