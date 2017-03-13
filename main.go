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
	//"time"
)

func runFSM(deleteOrderChan chan source.Order) {
	var floorNumber int = -1
	for {
		FSM.UpdateElevator()
		floorNumber = driver.ElevatorGetFloorSensorSignal()
		FSM.CheckIfElevatorIsStuck(floorNumber)

		if floorNumber != -1 {
			driver.ElevatorSetFloorIndicator(floorNumber)
			FSM.ElevatorHasArrivedAtFloor(floorNumber, deleteOrderChan)
			FSM.CheckFloorAndSetElevetorDirection(deleteOrderChan)
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
	newOrderChan := make(chan source.Order, 10)       // <- Kan tune det taller her
	deleteOrderChan := make(chan source.Order, 10)    // <- Kan tune det taller her
	deleteAddOrderChan := make(chan source.Order, 10) // <- Kan tune det taller her

	port := ":20013"

	if queue.CreateFile() {
		msg.ID = UDP.GetLocalID(UDP.GetLocalIP())
		queue.WriteToFile(msg)
	} else {
		msg = queue.ReadFromFile()
	}

	driver.InitializeElevator()
	FSM.ElevatorStartUp()
	queue.Init()

	go UDP.Receiving(port, newMsgChanRecive)
	go UDP.Transmitting(port, msg, newMsgChanTransmit)
	go queue.UpdateOrdersAfterReceiving(newMsgChanRecive, deleteAddOrderChan)
	go queue.UpdateElevatorInfoBeforeTransmitting(newMsgChanTransmit, newOrderChan, deleteOrderChan, deleteAddOrderChan, msg)
	go queue.CheckForOrdersAndDistribute(newOrderChan)
	go runFSM(deleteOrderChan)

	for {
		handleKill()
	}
}
