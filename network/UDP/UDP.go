package UDP

import (
	"../../source"
	"encoding/json"
	"net"
	"strings"
	"time"
)

func GetLocalID() string {
	addr, err := net.InterfaceAddrs()
	source.CheckForError(err)
	IP := strings.Split(addr[1].String(), "/")
	localIP := IP[0]
	localID := strings.Split(localIP, ".")
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

func TransmittingBroadcast(port string, msg source.ElevInfo, newMsgChanTransmit chan source.ElevInfo, onlineStatus chan bool) {
	adress, err := net.ResolveUDPAddr("udp", "129.241.187.255"+port)
	source.CheckForError(err)
	connection, err := net.DialUDP("udp", nil, adress)
	source.CheckForError(err)
	for {
		select {
		case msg = <-newMsgChanTransmit:
			buf, _ := json.Marshal(msg)
			_, err = connection.Write(buf)
			if source.CheckForError(err) {
				select {
				case onlineStatus <- false:
				default:
				}
			} else {
				select {
				case onlineStatus <- true:
				default:
				}
			}
		default:
			buf, _ := json.Marshal(msg)
			_, err = connection.Write(buf)
			if source.CheckForError(err) {
				select {
				case onlineStatus <- false:
				default:
				}
			} else {
				select {
				case onlineStatus <- true:
				default:
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}
