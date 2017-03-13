package FSM

import (
	"../driver"
	"../queue"
	"../source"
	"fmt"
	"time"
)

const (
	Idle     = "Idle    "
	Running  = "Running "
	DoorOpen = "DoorOpen"
	Stuck    = "Stuck   "
)

var arrivedFloor int = -1
var currentFloor int = -1
var elevatorState string = ""
var currentDirection int = -1

var stuckTimer = time.Now().Add(1 * time.Hour)

func ElevatorStartUp() {
	fmt.Println("Elevator init")
	driver.ElevatorSetDoorOpenLamp(false)
	for driver.ElevatorGetFloorSensorSignal() == -1 {
		driver.ElevatorSetMotorDirection(driver.MotorDirectionDown)
	}
	driver.ElevatorSetMotorDirection(driver.MotorDirectionStop)
	currentFloor = driver.ElevatorGetFloorSensorSignal()
	driver.ElevatorSetFloorIndicator(currentFloor)
	currentDirection = driver.MotorDirectionDown
	elevatorState = Idle
	UpdateElevator()
	fmt.Println("Elevator ready")
}

func ElevatorHasArrivedAtFloor(floorNumber int, deleteOrderChan chan source.Order) {
	var deleteOrder source.Order
	switch elevatorState {
	case Running:
		arrivedFloor = floorNumber
		if queue.CheckIfOrderTableIsEmpty() {
			driver.ElevatorSetMotorDirection(driver.MotorDirectionStop)
			elevatorState = Idle
			UpdateElevator()
		} else if arrivedFloor != currentFloor {
			currentFloor = arrivedFloor
			if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection) {
				driver.ElevatorSetMotorDirection(driver.MotorDirectionStop)
				elevatorState = DoorOpen
				UpdateElevator()
				deleteOrder.Command = "delete"
				deleteOrder.Floor = currentFloor
				deleteOrderChan <- deleteOrder
				doorTimer()
			}
		}
		break
	default:
		break
	}
}

func CheckFloorAndSetElevetorDirection(deleteOrderChan chan source.Order) {
	var deleteOrder source.Order
	switch elevatorState {
	case DoorOpen:
		if queue.CheckIfOrderTableIsEmpty() {
			elevatorState = Idle
			UpdateElevator()
			break
		} else if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection) {
			deleteOrder.Command = "delete"
			deleteOrder.Floor = currentFloor
			deleteOrderChan <- deleteOrder
			UpdateElevator()
			doorTimer()
			break
		} else {
			currentDirection = queue.GetMotorDirection(currentFloor, currentDirection)
			driver.ElevatorSetMotorDirection(currentDirection)
			deleteOrder.Command = "none"
			deleteOrderChan <- deleteOrder
			elevatorState = Running
			UpdateElevator()
			StartStuckTimer()
			break
		}
	case Idle:
		if queue.CheckIfOrderTableIsEmpty() {
			break
		} else if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection) {
			deleteOrder.Command = "delete"
			deleteOrder.Floor = currentFloor
			deleteOrderChan <- deleteOrder
			elevatorState = DoorOpen
			UpdateElevator()
			doorTimer()
			break
		} else {
			currentDirection = queue.GetMotorDirection(currentFloor, currentDirection)
			driver.ElevatorSetMotorDirection(currentDirection)
			deleteOrder.Command = "none"
			deleteOrderChan <- deleteOrder
			elevatorState = Running
			UpdateElevator()
			StartStuckTimer()
			break
		}
	default:
		break
	}
}

func CheckIfElevatorIsStuck(floorNumber int) {
	switch elevatorState {
	case Running:
		if floorNumber != -1 && floorNumber != currentFloor {
			StartStuckTimer()
		}
		CheckStuckTimer()
		break
	case Stuck:
		if floorNumber != -1 && floorNumber != currentFloor {
			currentFloor = floorNumber
			elevatorState = Running
			UpdateElevator()
			StartStuckTimer()
		}
		break
	default:
		break
	}
}

func doorTimer() {
	DoorOpenTimer := time.NewTimer(time.Second * 3)
	driver.ElevatorSetDoorOpenLamp(true)
	<-DoorOpenTimer.C
	driver.ElevatorSetDoorOpenLamp(false)
}

func StartStuckTimer() {
	currentTime := time.Now()
	stuckTimer = currentTime.Add(5 * time.Second)
}

func CheckStuckTimer() {
	currentTime := time.Now()
	if currentTime.After(stuckTimer) {
		elevatorState = Stuck
	}
}

func UpdateElevator() {
	queue.UpdateElevatorDirection(currentDirection)
	queue.UpdateElevatorFloor(currentFloor)
	queue.UpdateElevatorState(elevatorState)
}
