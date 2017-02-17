package main

import (
	//"./driver"
	"./network/UDP"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

func print_msg(newMsgChanRecive chan UDP.Elev_info) {

	var recievedMsg UDP.Elev_info

	for {
		recievedMsg = <-newMsgChanRecive
		fmt.Println(recievedMsg.ID, "--", recievedMsg.Words, "--", recievedMsg.NewOrder.ID, "--", recievedMsg.Recipe)
	}

}

func changeTransmit(newMsgChanTransmit chan UDP.Elev_info, msg UDP.Elev_info) {

	for {
		fmt.Scan(&msg.Words)
		newMsgChanTransmit <- msg
		writeToFile(msg)
	}
}

func createFile() bool {
	test, err := os.Open("file.txt")
	test.Close()
	if err != nil {
		file, err := os.Create("file.txt")
		file.Close()
		if err != nil {
			log.Fatal("Cannot create file", err)
		}
		return true
	}
	return false
}

func writeToFile(msg UDP.Elev_info) {
	_ = os.Remove("file.txt")
	file, _ := os.Create("file.txt")
	file.Close()
	file, err := os.OpenFile("file.txt", os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Cannot write to file", err)
	}

	buf, _ := json.Marshal(msg)
	_, err = file.Write(buf)

	file.Close()
}

func readFromFile() UDP.Elev_info {
	file, err := os.Open("file.txt")
	if err != nil {
		log.Fatal("Cannot read from file", err)
	}

	data := make([]byte, 1024)
	count, err := file.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	var msgFromFile UDP.Elev_info

	err = json.Unmarshal(data[:count], &msgFromFile)
	if err != nil {
		fmt.Println("Error:", err)
	}

	file.Close()
	return msgFromFile
}

func delete_file() {
	_ = os.Remove("file.txt")
}

func test_UDP_network() {
	var msg UDP.Elev_info

	newMsgChanRecive := make(chan UDP.Elev_info, 1)
	newMsgChanTransmit := make(chan UDP.Elev_info, 1)

	port := ":20003"

	if createFile() {
		msg.ID = "ElevInfo"
		msg.Words = "Hello from school"
		msg.ElevIP = UDP.GetLocalIP()
		msg.NewOrder.ID = "NewOrder"
		writeToFile(msg)
	} else {
		msg = readFromFile()
	}

	go UDP.Receiving(port, newMsgChanRecive)
	go UDP.Transmitter(port, msg, newMsgChanTransmit)
	go print_msg(newMsgChanRecive)
	go changeTransmit(newMsgChanTransmit, msg)

	for {
		time.Sleep(1 * time.Second)
	}
}
/*
func test_driver() {
	driver.InitializeElevator()
	driver.ElevatorSetMotorDirection(1)

	var currentFloor int

	for {
		currentFloor = driver.ElevatorGetFloorSensorSignal()
		if currentFloor != -1 {
			driver.ElevatorSetFloorIndicator(currentFloor)
		}
		if currentFloor == 3 {
			driver.ElevatorSetMotorDirection(-1)
		} else if currentFloor == 0 {
			driver.ElevatorSetMotorDirection(1)
		}
		if driver.ElevatorGetButtonSignal(2, 3) {
			driver.ElevatorSetMotorDirection(0)
		}
		if driver.ElevatorGetButtonSignal(2, 2) {
			driver.ElevatorSetMotorDirection(1)
		}
	}

}
*/
func main() {
	//delete_file()
	test_UDP_network()
	//test_driver()
}
