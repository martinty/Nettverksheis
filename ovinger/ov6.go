package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const FILENAME = "counterSave.gob"
const PORT_HAX_FILE = "port.gob"

func main() {

	fmt.Println("I am backup")

	localIP := "129.241.187.150"
	localPort := read_port_from_file()
	sendToPort := write_next_port_to_file(localPort)
	localAddr := localIP + ":" + localPort
	sendToAddr := localIP + ":" + sendToPort

	fmt.Println("this is sendToAddr: ", sendToAddr)
	udpAddr, err := net.ResolveUDPAddr("udp4", localAddr)

	if err != nil {
		panic(err)
	}

	connection, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Listening on port ", localPort)
	defer connection.Close()

	masterChan := make(chan bool)
	go detect_precense(connection, masterChan)

	<-masterChan
	fmt.Println("I am master")
	spawn_backup()
	counterChan := make(chan int)
	go read_from_file(counterChan)
	time.Sleep(time.Second)
	go spam_precense(sendToAddr)
	go counter_and_write_to_file(counterChan)
	time.Sleep(time.Second * 10000)
}

func counter_and_write_to_file(counterChan chan int) {
	counter := <-counterChan
	for {
		counter++

		dataFile, err := os.Create(FILENAME)

		if err != nil {
			fmt.Println("Error while writing next port to file")
			panic(err)
		}

		dataEncoder := gob.NewEncoder(dataFile)
		dataEncoder.Encode(counter)
		dataFile.Close()
		fmt.Println(counter)
		time.Sleep(time.Millisecond * 200)
	}
}

func read_from_file(counterChan chan int) {

	var data int
	var counter int

	if _, err := os.Stat(FILENAME); os.IsNotExist(err) {
		counter = 0
	} else {
		dataFile, err := os.Open(FILENAME)
		dataDecoder := gob.NewDecoder(dataFile)
		err = dataDecoder.Decode(&data)

		if err != nil {
			fmt.Println("error in reading counter")
			panic(err)
		}
		counter = data
	}
	counterChan <- counter
}

func spawn_backup() {
	command := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run ov6.go")
	err := command.Run()
	if err != nil {
		fmt.Println("You messed up in spawn_backup")
		panic(err)
	}
	fmt.Println("backup should be spawned")
}

func spam_precense(remoteAddr string) {
	udpRemote, _ := net.ResolveUDPAddr("udp", remoteAddr)

	connection, err := net.DialUDP("udp", nil, udpRemote)
	if err != nil {
		fmt.Println("You messed up in spam presence")
		panic(err)
	}
	defer connection.Close()
	for {
		_, err := connection.Write([]byte("I am the master"))
		if err != nil {
			fmt.Println("You messed up in spam_precense")
			panic(err)
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func detect_precense(connection *net.UDPConn, masterChan chan bool) {
	buffer := make([]byte, 2048)
	for {
		t := time.Now()
		connection.SetReadDeadline(t.Add(3 * time.Second))
		_, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("UDP timeout: ", err)
			masterChan <- true
			break
		}
	}
}

func write_next_port_to_file(portNum string) string {
	portNumToFile, _ := strconv.Atoi(portNum)
	portNumToFile++
	dataFile, err := os.Create(PORT_HAX_FILE)

	if err != nil {
		fmt.Println("Error while writing next port to file")
		panic(err)
	}

	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(portNumToFile)
	dataFile.Close()

	return strconv.Itoa(portNumToFile)
}

func read_port_from_file() string {
	var data int

	if _, err := os.Stat(PORT_HAX_FILE); os.IsNotExist(err) {
		return "20058"
	}
	dataFile, err := os.Open(PORT_HAX_FILE)
	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&data)

	if err != nil {
		fmt.Println("error in reading port")
		panic(err)
	}
	portToRead := strconv.Itoa(data)
	return portToRead
}
