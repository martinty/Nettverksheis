package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
	"encoding/json"
	"strconv"
	"os/signal"
	"syscall"
)

type info struct {
	Counting int
	Port     int
	NewStart bool
}

var state string

func main() {
	fmt.Println("I'm the Backup")
	var data info
	const heartBeat = 1

	if createFile(){
		data.Counting = 0
		data.Port = 1000
		data.NewStart = false
		writeToFile(data)
	} else{
		data = readFromFile()
		if data.NewStart {
			data.Counting = 0
			data.Port = 1000
			data.NewStart = false
			writeToFile(data)
		}
	}

	state = "backup"
	port := ":"+strconv.Itoa(data.Port)
	newMsgChanReceive := make(chan int, 1)

	go handleProgramKill(data)
	go receiving(port, newMsgChanReceive)

	if checkIfMasterIsDead(newMsgChanReceive){
		data = readFromFile()
		data.Port += 1
		writeToFile(data)
		port = ":"+strconv.Itoa(data.Port)
		go transmittingBroadcast(port, heartBeat)
		fmt.Println("I'm now the Master")
		state = "master"
		spawnNewTerminal()
	}

	for {
		time.Sleep(time.Second)
		data = readFromFile()
		fmt.Printf("%+v\n",data)
		data.Counting += 1
		writeToFile(data)
	}
}

func checkIfMasterIsDead(newMsgChanReceive chan int) bool {
	offlineTimer := time.Now().Unix()
	for{
		select {
		case <-newMsgChanReceive:
			offlineTimer = time.Now().Unix()
		default:
			if time.Now().Unix()-offlineTimer > 1 {
				return true
			}
		}
	}
}

func receiving(port string, newMsgChanReceive chan int) {
	address, _ := net.ResolveUDPAddr("udp", port)
	connection, _ := net.ListenUDP("udp", address)
	receiveBuffer := make([]byte, 1024)
	for {
		n, _, _ := connection.ReadFromUDP(receiveBuffer)
		var newMsg int
		_ = json.Unmarshal(receiveBuffer[:n], &newMsg)
		newMsgChanReceive <- newMsg
	}
}

func transmittingBroadcast(port string, heartBeat int) {
	address, _ := net.ResolveUDPAddr("udp", "192.168.10.146"+port) //.255 in the end to BraodcastAll
	connection, _ := net.DialUDP("udp", nil, address)
	for {
		buf, _ := json.Marshal(heartBeat)
		_, _ = connection.Write(buf)
		time.Sleep(time.Millisecond*100)
	}
}

func spawnNewTerminal()  {
	cmd := exec.Command("cmd", "/C  start go run phoenix.go")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func handleProgramKill(data info) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	if state == "backup"{
		data.NewStart = true
		writeToFile(data)
	}
	os.Exit(1)
}

func createFile() bool {
	test, err := os.Open("backup.txt")
	test.Close()
	if err != nil {
		file, _ := os.Create("backup.txt")
		file.Close()
		return true
	}
	return false
}

func writeToFile(data info) {
	_ = os.Remove("backup.txt")
	file, _ := os.Create("backup.txt")
	file.Close()
	file,_ = os.OpenFile("backup.txt", os.O_WRONLY, 0666)
	buf, _ := json.Marshal(data)
	_, _ = file.Write(buf)
	file.Close()
}

func readFromFile() info {
	file, _ := os.Open("backup.txt")
	data := make([]byte, 1024)
	count, _ := file.Read(data)
	var msgFromFile info
	_ = json.Unmarshal(data[:count], &msgFromFile)
	file.Close()
	return msgFromFile
}

func deleteFile()  {
	_ = os.Remove("backup.txt")
}