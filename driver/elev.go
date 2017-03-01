package driver

import (
    "../source"
    "fmt"
)

const motorSpeed = 2800
const NumFloors int = 4
const NumButtons int = 3
const on bool = true
const off bool = false

var ButtonType = map[string]int{
    "Button call up":        0,
    "Button call down":      1,
    "Button internal panel": 2,
}

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

    var initSuccess bool = IOInitializeElevator()

    if !initSuccess {

        return false
    }

    for floor := 0; floor < NumFloors; floor++ {
        for button := 0; button < NumButtons; button++ {

            if (button == ButtonType["Button call down"]) && (floor != 0) {
                ElevatorSetButtonLamp(button, floor, off)
            }

            if (button == ButtonType["Button call up"]) && (floor != 3) {
                ElevatorSetButtonLamp(button, floor, off)
            }
            if button == ButtonType["Button internal panel"] {
                ElevatorSetButtonLamp(button, floor, off)
            }
        }
    }

    ElevatorSetDoorOpenLamp(off)
    return true
}

func ElevatorSetMotorDirection(motorDirection int) {

    if motorDirection > 0 { //Set direction up if positive number
        IOClearBit(motorDir)
        IOWriteAnalog(motor, motorSpeed)
    } else if motorDirection < 0 { //Set direction down if negative number
        IOSetBit(motorDir)
        IOWriteAnalog(motor, motorSpeed)
    } else if motorDirection == 0 {
        IOWriteAnalog(motor, 0) //if not stop elevator
    } else {
        fmt.Println("Unable to set motor direction")
    }
}

func ElevatorSetButtonLamp(setButtonType int, floor int, on bool) {

    if (floor < 0) || (floor > NumFloors) {
        fmt.Println("Invalid floor to set buttonlamp")
    }

    if (setButtonType < 0) || (setButtonType > NumButtons) {
        fmt.Println("Invalid button type")
    }
    if (floor == 0) && (setButtonType == ButtonType["Button call down"]) {
        fmt.Println("Invalid button type to set button lamp")
    }
    if (floor == NumFloors-1) && (setButtonType == ButtonType["Button call up"]) {
        fmt.Println("Invalid button type to set button lamp")
    }

    if on {
        IOSetBit(lampChannelsMatrix[floor][setButtonType])
    } else {
        IOClearBit(lampChannelsMatrix[floor][setButtonType])
    }
}

func ElevatorSetFloorIndicator(floor int) {
    if floor < 0 || floor > NumFloors {
        fmt.Println("Invalid floor to set floor indicator")
    }

    // Binary encoding. one light must always be on.
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

func ElevatorSetDoorOpenLamp(on bool) {
    if on {
        IOSetBit(lightDoorOpen)
    } else {
        IOClearBit(lightDoorOpen)
    }
}

func ElevatorGetButtonSignal(buttonUpdate source.ElevInfo) source.ElevInfo {
    for floor := 0; floor < NumFloors; floor++ {
        for button := 0; button < NumButtons; button++ {
            if IOReadBit(buttonChannelsMatrix[floor][button]) != 0 {
                if button == 2 {
                    buttonUpdate.LocalOrders[floor] = 1
                } else {
                    buttonUpdate.ExternalOrders[floor][button] = 1
                }
            }
        }
    }
    return buttonUpdate
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
