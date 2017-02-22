package main

import (
	"./driver"
	"./network/UDP"
	"./queue"
	"./source"
	"encoding/json"
	"fmt"
	//"log"
	"os"
	"time"
)

func printMsg(newMsgChanRecive chan source.ElevInfo) {

	var recievedMsg source.ElevInfo
	for {
		recievedMsg = <-newMsgChanRecive
		writeToFile(recievedMsg)
		fmt.Println(recievedMsg.ID, "--", recievedMsg.Words, "--", recievedMsg.NewOrder.ID, "--", recievedMsg.Recipe)
	}
}

func changeTransmit(newMsgChanTransmit chan source.ElevInfo, msg source.ElevInfo) {

	for {
		fmt.Scan(&msg.Words)
		newMsgChanTransmit <- msg
	}
}

func createFile() bool {
	test, err := os.Open("file.txt")
	test.Close()
	if err != nil {
		file, err := os.Create("file.txt")
		file.Close()
		source.CheckForError(err)
		return true
	}
	return false
}

func writeToFile(msg source.ElevInfo) {
	_ = os.Remove("file.txt")
	file, _ := os.Create("file.txt")
	file.Close()
	file, err := os.OpenFile("file.txt", os.O_WRONLY, 0666)
	source.CheckForError(err)

	buf, _ := json.Marshal(msg)
	_, err = file.Write(buf)
	source.CheckForError(err)

	file.Close()
}

func readFromFile() source.ElevInfo {
	file, err := os.Open("file.txt")
	source.CheckForError(err)

	data := make([]byte, 1024)
	count, err := file.Read(data)
	source.CheckForError(err)

	var msgFromFile source.ElevInfo

	err = json.Unmarshal(data[:count], &msgFromFile)
	source.CheckForError(err)

	file.Close()
	return msgFromFile
}

func deleteFile() {
	_ = os.Remove("file.txt")
}

func testUDPNetwork() {
	var msg source.ElevInfo

	newMsgChanRecive := make(chan source.ElevInfo, 1)
	newMsgChanTransmit := make(chan source.ElevInfo, 1)

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
	go UDP.Transmitting(port, msg, newMsgChanTransmit)
	go printMsg(newMsgChanRecive)
	go changeTransmit(newMsgChanTransmit, msg)

	for {
		time.Sleep(1 * time.Second)
	}
}

func testDriver() {
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

func testQueue() {
	var QueueMsg source.ElevInfo
	queue.Init(QueueMsg)
}

func main() {
	//deleteFile()
	go testUDPNetwork()
	//go testDriver()
	//testQueue()
	for {
	}

}
