package main

import (
	"./FSM"
	"./driver"
	"./network/UDP"
	"./queue"
	"./source"
	//"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func printMsg(newMsgChanRecive chan source.ElevInfo) {

	var recievedMsg source.ElevInfo
	for {
		recievedMsg = <-newMsgChanRecive
		queue.WriteToFile(recievedMsg)
		queue.UpdateTables(recievedMsg)
		//fmt.Println(recievedMsg)
	}
}

func updateTransmit(newMsgChanTransmit chan source.ElevInfo, msg source.ElevInfo) {

	for {
		msg = queue.UpdateElevatorInfo(msg)
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

	port := ":20007"

	if queue.CreateFile() {
		msg.IP = UDP.GetLocalIP()
		msg.ID = UDP.GetLocalID(msg.IP)
		queue.WriteToFile(msg)
	} else {
		msg = queue.ReadFromFile()
	}

	go UDP.Receiving(port, newMsgChanRecive)
	go UDP.Transmitting(port, msg, newMsgChanTransmit)
	go printMsg(newMsgChanRecive)
	go updateTransmit(newMsgChanTransmit, msg)

	for {
		time.Sleep(1 * time.Second)
	}
}

func testQueue() {
	for {
		queue.UpdateOrders()
		queue.UpdateButtonLight()
	}
}

func testFSM() {
	var floorNumber int = -1
	for {
		FSM.UpdateElevator()
		floorNumber = driver.ElevatorGetFloorSensorSignal()
		FSM.CheckIfElevatorIsStuck(floorNumber)

		if floorNumber != -1 {
			driver.ElevatorSetFloorIndicator(floorNumber)
			FSM.ElevatorHasArrivedAtFloor(floorNumber)
			FSM.SetElevetorDirection()
		}
	}
}

func handleKill() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	driver.ElevatorSetMotorDirection(2)
	os.Exit(1)
}

func main() {

	//queue.DeleteFile()

	var msg source.ElevInfo

	newMsgChanRecive := make(chan source.ElevInfo, 1)
	newMsgChanTransmit := make(chan source.ElevInfo, 1)

	port := ":20007"

	if queue.CreateFile() {
		msg.IP = UDP.GetLocalIP()
		msg.ID = UDP.GetLocalID(msg.IP)
		queue.WriteToFile(msg)
	} else {
		msg = queue.ReadFromFile()
	}

	driver.InitializeElevator()
	FSM.ElevatorStartUp()

	go UDP.Receiving(port, newMsgChanRecive)
	go UDP.Transmitting(port, msg, newMsgChanTransmit)
	go printMsg(newMsgChanRecive)
	go updateTransmit(newMsgChanTransmit, msg)
	go testQueue()
	go testFSM()

	for {
		handleKill()
	}
}
