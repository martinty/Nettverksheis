package FMS

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
)

var arrivedFloor int = -1
var currentFloor int = -1
var elevatorState int = -1
var currentDirection int = -1

func ElevatorStartUp() {
	for driver.ElevatorGetFloorSensorSignal() != 0 {
		driver.ElevatorSetMotorDirection(driver.MotorDirectionDown)
	}
	currentFloor = 0
	driver.ElevatorSetMotorDirection(driver.MotorDirectionStop)
	driver.ElevatorSetFloorIndicator(currentFloor)
	currentDirection = driver.MotorDirectionDown
	elevatorState = Idle
}

func ElevatorHasArrivedAtFloor(floorNumber int){
	switch(elevatorState){
	case Running:
		arrivedFloor = floorNumber
		if arrivedFloor != currentFloor{
			currentFloor = arrivedFloor
			if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection){
				driver.ElevatorSetMotorDirection(driver.MotorDirectionStop)
				timer()
				queue.ClearOrdersAtFloor(currentFloor)
				elevatorState = Idle
			}
		}
		break
	case Idle:
		if queue.ShouldElevatorStopAtFloor(currentFloor, currentDirection){
			timer()
			queue.ClearOrdersAtFloor(currentFloor)
		}
		break
	default:
		break
	}
}

func SetElevetorDirection() {
	switch(elevatorState){
		case Idle:
			if queue.CheckIfOrderTableIsEmpty(){
				break
			} else{
				currentDirection = queue.GetMotorDirection(currentFloor, currentDirection)
				driver.ElevatorSetMotorDirection(currentDirection)
				elevatorState = Running
				break				
			}
		default:
			break
	}
}

func timer(){
	DoorOpenTimer := time.NewTimer(time.Second * 3)
	fmt.Println("Door open")
	<-DoorOpenTimer.C
	fmt.Println("Door closed")
}
