package FSM

import (
	"../driver"
	"../queue"
	"../source"
	"fmt"
	"time"
)

var arrivedFloor int = -1
var currentFloor int = -1
var elevatorState string = ""
var currentDirection int = -1
var stuckTimer = time.Now().Add(1 * time.Hour)

const (
	idle     = "idle    "
	running  = "running "
	doorOpen = "doorOpen"
	stuck    = "stuck   "
)

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
	elevatorState = idle
	UpdateElevator()
	fmt.Println("Elevator ready")
}

func ElevatorHasArrivedAtFloor(floorNumber int, deleteOrderChan chan source.Order, onlineStatus chan bool) {
	var deleteOrder source.Order
	var elevatorOnline bool
	switch elevatorState {
	case running:
		arrivedFloor = floorNumber
		if queue.CheckIfOrderTableIsEmpty() {
			driver.ElevatorSetMotorDirection(driver.MotorDirectionStop)
			elevatorState = idle
			UpdateElevator()
		} else if arrivedFloor != currentFloor {
			currentFloor = arrivedFloor
			if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection) {
				driver.ElevatorSetMotorDirection(driver.MotorDirectionStop)
				elevatorState = doorOpen
				UpdateElevator()
				elevatorOnline = <-onlineStatus
				if elevatorOnline {
					deleteOrder.Command = "delete"
					deleteOrder.Floor = currentFloor
					deleteOrderChan <- deleteOrder
				} else {
					queue.ClearOrdersAtFloor(currentFloor, "offline")
				}
				doorTimer()
			}
		} else if currentFloor == driver.NumFloors-1 {
			driver.ElevatorSetMotorDirection(driver.MotorDirectionDown)
			currentDirection = driver.MotorDirectionDown
		} else if currentFloor == 0 {
			driver.ElevatorSetMotorDirection(driver.MotorDirectionUp)
			currentDirection = driver.MotorDirectionUp
		}
		break
	default:
		break
	}
}

func CheckFloorAndSetElevetorDirection(deleteOrderChan chan source.Order, onlineStatus chan bool) {
	var deleteOrder source.Order
	var elevatorOnline bool
	switch elevatorState {
	case doorOpen:
		if queue.CheckIfOrderTableIsEmpty() {
			elevatorState = idle
			UpdateElevator()
			break
		} else if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection) {
			elevatorOnline = <-onlineStatus
			if elevatorOnline {
				deleteOrder.Command = "delete"
				deleteOrder.Floor = currentFloor
				deleteOrderChan <- deleteOrder
			} else {
				queue.ClearOrdersAtFloor(currentFloor, "offline")
			}
			UpdateElevator()
			doorTimer()
			break
		} else {
			currentDirection = queue.GetMotorDirection(currentFloor, currentDirection)
			driver.ElevatorSetMotorDirection(currentDirection)
			elevatorOnline = <-onlineStatus
			if elevatorOnline {
				deleteOrder.Command = "none"
				deleteOrder.Floor = currentFloor
				deleteOrderChan <- deleteOrder
			} else {
				queue.ClearOrdersAtFloor(currentFloor, "offline")
			}
			elevatorState = running
			UpdateElevator()
			startstuckTimer()
			break
		}
	case idle:
		if queue.CheckIfOrderTableIsEmpty() {
			break
		} else if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection) {
			elevatorOnline = <-onlineStatus
			if elevatorOnline {
				deleteOrder.Command = "delete"
				deleteOrder.Floor = currentFloor
				deleteOrderChan <- deleteOrder
			} else {
				queue.ClearOrdersAtFloor(currentFloor, "offline")
			}
			elevatorState = doorOpen
			UpdateElevator()
			doorTimer()
			break
		} else {
			currentDirection = queue.GetMotorDirection(currentFloor, currentDirection)
			driver.ElevatorSetMotorDirection(currentDirection)
			elevatorOnline = <-onlineStatus
			if elevatorOnline {
				deleteOrder.Command = "none"
				deleteOrder.Floor = currentFloor
				deleteOrderChan <- deleteOrder
			} else {
				queue.ClearOrdersAtFloor(currentFloor, "offline")
			}
			elevatorState = running
			UpdateElevator()
			startstuckTimer()
			break
		}
	default:
		break
	}
}

func CheckIfElevatorIsStuck(floorNumber int) {
	switch elevatorState {
	case running:
		if floorNumber != -1 && floorNumber != currentFloor {
			startstuckTimer()
		}
		checkstuckTimer()
		break
	case stuck:
		if floorNumber != -1 && floorNumber != currentFloor {
			currentFloor = floorNumber
			elevatorState = running
			UpdateElevator()
			startstuckTimer()
		}
		break
	default:
		break
	}
}

func UpdateElevator() {
	queue.UpdateElevatorDirection(currentDirection)
	queue.UpdateElevatorFloor(currentFloor)
	queue.UpdateElevatorState(elevatorState)
}

func doorTimer() {
	doorOpenTimer := time.NewTimer(time.Second * 3)
	driver.ElevatorSetDoorOpenLamp(true)
	<-doorOpenTimer.C
	driver.ElevatorSetDoorOpenLamp(false)
}

func startstuckTimer() {
	currentTime := time.Now()
	stuckTimer = currentTime.Add(5 * time.Second)
}

func checkstuckTimer() {
	currentTime := time.Now()
	if currentTime.After(stuckTimer) {
		elevatorState = stuck
	}
}
