package FSM

import (
	"../driver"
	//"../source"
	"../queue"
	"fmt"
	"time"
)

const (
	Idle = iota
	Running
	Stuck
)

var arrivedFloor int = -1
var currentFloor int = -1
var elevatorState int = -1
var currentDirection int = -1

var stuckTimer = time.Now().Add(100 * time.Second)

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
	queue.Init()
	UpdateElevator()
	fmt.Println("Elevator ready")
}

func ElevatorHasArrivedAtFloor(floorNumber int) {
	switch elevatorState {
	case Running:
		arrivedFloor = floorNumber
		if arrivedFloor != currentFloor {
			currentFloor = arrivedFloor
			if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection) {
				driver.ElevatorSetMotorDirection(driver.MotorDirectionStop)
				elevatorState = Idle
				UpdateElevator()
				doorTimer()
				queue.ClearOrdersAtFloor(currentFloor)
			}
		}
		break
	case Idle:
		if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection) {
			doorTimer()
			queue.ClearOrdersAtFloor(currentFloor)
		}
		break
	default:
		break
	}
}

func SetElevetorDirection() {
	switch elevatorState {
	case Idle:
		if queue.CheckIfOrderTableIsEmpty() {
			break
		} else {
			currentDirection = queue.GetMotorDirection(currentFloor, currentDirection)
			driver.ElevatorSetMotorDirection(currentDirection)
			elevatorState = Running
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
			elevatorState = Running
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
