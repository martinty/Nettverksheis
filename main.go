package main

import (
	"./driver"
	"./network/UDP"
	"./queue"
	"./source"
	"./FSM"
	"fmt"
	"time"
)

func printMsg(newMsgChanRecive chan source.ElevInfo) {

	var recievedMsg source.ElevInfo
	for {
		recievedMsg = <-newMsgChanRecive
		queue.WriteToFile(recievedMsg)
		fmt.Println(recievedMsg.ID, "--", recievedMsg.ExternalOrders, "--", recievedMsg.LocalOrders)
	}
}

func changeTransmit(newMsgChanTransmit chan source.ElevInfo, msg source.ElevInfo) {

	for {
		msg = driver.ElevatorGetButtonSignal(msg)
		select {
		case newMsgChanTransmit <- msg:
			continue
		default:
			continue
		}
	}
}

func testUDPNetwork() {
	var msg source.ElevInfo

	newMsgChanRecive := make(chan source.ElevInfo, 1)
	newMsgChanTransmit := make(chan source.ElevInfo, 1)

	port := ":20023"

	if queue.CreateFile() {
		msg.ID = "ElevInfo"
		msg.Words = "Hello from school"
		msg.ElevIP = UDP.GetLocalIP()
		msg.NewOrder.ID = "NewOrder"
		queue.WriteToFile(msg)
	} else {
		msg = queue.ReadFromFile()
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
			driver.ElevatorSetMotorDirection(1)
		} else if currentFloor == 0 {
			driver.ElevatorSetMotorDirection(0)
		}
	}
}

func testQueue() {
	for{
		queue.UpdateOrders()
	}
}

func testFSM(){
	FMS.ElevatorStartUp()
	var floorNumber int = -1
	for{
		floorNumber = driver.ElevatorGetFloorSensorSignal()
		if floorNumber != -1{
			driver.ElevatorSetFloorIndicator(floorNumber)
			FMS.ElevatorHasArrivedAtFloor(floorNumber)
			FMS.SetElevetorDirection()
		}
	}
}

func main() {
	driver.InitializeElevator()
	queue.DeleteFile()
	go testUDPNetwork()
	//go testDriver()
	go testQueue()
	go testFSM()
	for {
	}
}
