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
	Update()
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
				timer()
				queue.ClearOrdersAtFloor(currentFloor)
				elevatorState = Idle
			}
		}
		break
	case Idle:
		if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection) {
			timer()
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
			break
		}
	default:
		break
	}
}

func timer() {
	Update()
	DoorOpenTimer := time.NewTimer(time.Second * 3)
	driver.ElevatorSetDoorOpenLamp(true)
	<-DoorOpenTimer.C
	driver.ElevatorSetDoorOpenLamp(false)
}

func Update() {
	queue.UpdateElevatorDirection(currentDirection)
	queue.UpdateElevatorFloor(currentFloor)
}
