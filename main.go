package main

import (
	"./FSM"
	"./driver"
	"./network/UDP"
	"./queue"
	"./source"
	"os"
	"os/signal"
	"syscall"
)

func runFSM(deleteOrderChan chan source.Order, onlineStatus chan bool) {
	var floorNumber int = -1
	for {
		FSM.UpdateElevator()
		floorNumber = driver.ElevatorGetFloorSensorSignal()
		FSM.CheckIfElevatorIsStuck(floorNumber)

		if floorNumber != -1 {
			driver.ElevatorSetFloorIndicator(floorNumber)
			FSM.ElevatorHasArrivedAtFloor(floorNumber, deleteOrderChan, onlineStatus)
			FSM.CheckFloorAndSetElevetorDirection(deleteOrderChan, onlineStatus)
		}
	}
}

func handleProgramKill() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	driver.ElevatorSetMotorDirection(2)
	os.Exit(1)
}

func main() {

	var msg source.ElevInfo
	port := ":22017"
	msg.ID = UDP.GetLocalID()
	driver.InitializeElevator()
	FSM.ElevatorStartUp()
	queue.Init(msg.ID)

	newMsgChanRecive := make(chan source.ElevInfo, 1)
	newMsgChanTransmit := make(chan source.ElevInfo, 1)
	newOrderChan := make(chan source.Order, 10)
	deleteOrderChan := make(chan source.Order, 10)
	deleteAddOrderChan := make(chan source.Order, 1)
	onlineStatus := make(chan bool, 1)

	go UDP.Receiving(port, newMsgChanRecive)
	go UDP.TransmittingBroadcast(port, msg, newMsgChanTransmit, onlineStatus)
	go queue.UpdateOrdersAfterReceiving(newMsgChanRecive, deleteAddOrderChan)
	go queue.UpdateElevatorInfoBeforeTransmitting(newMsgChanTransmit, newOrderChan, deleteOrderChan, deleteAddOrderChan, msg)
	go queue.CheckForOrdersAndDistribute(newOrderChan)
	go runFSM(deleteOrderChan, onlineStatus)

	for {
		handleProgramKill()
	}
}
