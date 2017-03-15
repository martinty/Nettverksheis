package driver

import (
	"fmt"
)

const motorSpeed int = 2800
const NumFloors int = 4
const NumButtons int = 3
const ButtonTypeUp int = 0
const ButtonTypeDown int = 1
const ButtonTypeCommand int = 2
const MotorDirectionUp int = 0
const MotorDirectionDown int = 1
const MotorDirectionStop int = 2

var lampChannelsMatrix = [NumFloors][NumButtons]int{
	{lightUp1, lightDown1, lightCommand1},
	{lightUp2, lightDown2, lightCommand2},
	{lightUp3, lightDown3, lightCommand3},
	{lightUp4, lightDown4, lightCommand4},
}

var buttonChannelsMatrix = [NumFloors][NumButtons]int{
	{buttonUp1, buttonDown1, buttonCommand1},
	{buttonUp2, buttonDown2, buttonCommand2},
	{buttonUp3, buttonDown3, buttonCommand3},
	{buttonUp4, buttonDown4, buttonCommand4},
}

func InitializeElevator() bool {
	initStatus := IOInitializeElevator()
	if !initStatus {
		return false
	}
	for floor := 0; floor < NumFloors; floor++ {
		for button := 0; button < NumButtons; button++ {
			ElevatorSetButtonLamp(button, floor, false)
		}
	}
	ElevatorSetDoorOpenLamp(true)
	return true
}

func ElevatorSetMotorDirection(motorDirection int) {
	if motorDirection == 0 {
		IOClearBit(motorDir)
		IOWriteAnalog(motor, motorSpeed)
	} else if motorDirection == 1 {
		IOSetBit(motorDir)
		IOWriteAnalog(motor, motorSpeed)
	} else if motorDirection == 2 {
		IOWriteAnalog(motor, 0)
	} else {
		fmt.Println("Unable to set motor direction")
	}
}

func ElevatorSetButtonLamp(buttonType int, floor int, status bool) {
	if status {
		IOSetBit(lampChannelsMatrix[floor][buttonType])
	} else {
		IOClearBit(lampChannelsMatrix[floor][buttonType])
	}
}

func ElevatorSetFloorIndicator(floor int) {
	if floor < 0 || floor > NumFloors {
		fmt.Println("Invalid floor to set floor indicator")
	}
	if floor&0x02 != 0 {
		IOSetBit(lightFloorInd1)
	} else {
		IOClearBit(lightFloorInd1)
	}
	if floor&0x01 != 0 {
		IOSetBit(lightFloorInd2)
	} else {
		IOClearBit(lightFloorInd2)
	}
}

func ElevatorSetDoorOpenLamp(status bool) {
	if status {
		IOSetBit(lightDoorOpen)
	} else {
		IOClearBit(lightDoorOpen)
	}
}

func ElevatorCheckButtonSignal(button int, floor int) bool {
	if IOReadBit(buttonChannelsMatrix[floor][button]) != 0 {
		return true
	} else {
		return false
	}
}

func ElevatorGetFloorSensorSignal() int {

	if IOReadBit(sensorFloor1) != 0 {
		return 0
	} else if IOReadBit(sensorFloor2) != 0 {
		return 1
	} else if IOReadBit(sensorFloor3) != 0 {
		return 2
	} else if IOReadBit(sensorFloor4) != 0 {
		return 3
	} else {
		return -1
	}
}
