package driver

import (
    "fmt"
)

const MOTORSPEED = 2800
const NUMFLOORS int = 4
const NUMBUTTONS int = 3
const ON bool = true
const OFF bool = false

var ButtonType = map[string]int{
    "Button call up":        0,
    "Button call down":      1,
    "Button internal panel": 2,
}

var lampChannelsMatrix = [NUMFLOORS][NUMBUTTONS]int{
    {LIGHTUP1, LIGHTDOWN1, LIGHTCOMMAND1},
    {LIGHTUP2, LIGHTDOWN2, LIGHTCOMMAND2},
    {LIGHTUP3, LIGHTDOWN3, LIGHTCOMMAND3},
    {LIGHTUP4, LIGHTDOWN4, LIGHTCOMMAND4},
}

var buttonChannelsMatrix = [NUMFLOORS][NUMBUTTONS]int{
    {BUTTONUP1, BUTTONDOWN1, BUTTONCOMMAND1},
    {BUTTONUP2, BUTTONDOWN2, BUTTONCOMMAND2},
    {BUTTONUP3, BUTTONDOWN3, BUTTONCOMMAND3},
    {BUTTONUP4, BUTTONDOWN4, BUTTONCOMMAND4},
}

func ElevatorSetMotorDirection(motorDirection int) {

    if motorDirection > 0 { //Set direction up if positive number
        IOClearBit(MOTORDIR)
        IOWriteAnalog(MOTOR, MOTORSPEED)
    } else if motorDirection < 0 { //Set direction down if negative number
        IOSetBit(MOTORDIR)
        IOWriteAnalog(MOTOR, MOTORSPEED)
    } else if motorDirection == 0 {
        IOWriteAnalog(MOTOR, 0) //if not stop elevator
    } else {
        fmt.Println("Unable to set motor direction")
    }
}

func ElevatorSetButtonLamp(setButtonType int, floor int, on bool) {

    if (floor < 0) || (floor > NUMFLOORS) {
        fmt.Println("Invalid floor to set buttonlamp")
    }

    if (setButtonType < 0) || (setButtonType > NUMBUTTONS) {
        fmt.Println("Invalid button type")
    }
    if (floor == 0) && (setButtonType == ButtonType["Button call down"]) {
        fmt.Println("Invalid button type to set button lamp")
    }
    if (floor == NUMFLOORS-1) && (setButtonType == ButtonType["Button call up"]) {
        fmt.Println("Invalid button type to set button lamp")
    }

    if on {
        IOSetBit(lampChannelsMatrix[floor][setButtonType])
    } else {
        IOClearBit(lampChannelsMatrix[floor][setButtonType])
    }
}

func ElevatorSetFloorIndicator(floor int) {
    if floor < 0 || floor > NUMFLOORS {
        fmt.Println("Invalid floor to set floor indicator")
    }

    // Binary encoding. One light must always be on.
    if floor&0x02 != 0 {
        IOSetBit(LIGHTFLOORIND1)
    } else {
        IOClearBit(LIGHTFLOORIND1)
    }

    if floor&0x01 != 0 {
        IOSetBit(LIGHTFLOORIND2)
    } else {
        IOClearBit(LIGHTFLOORIND2)
    }
}

func ElevatorSetDoorOpenLamp(on bool) {
    if on {
        IOSetBit(LIGHTDOOROPEN)
    } else {
        IOClearBit(LIGHTDOOROPEN)
    }
}

func ElevatorGetButtonSignal(getButtonType int, floor int) bool {

    if (floor < 0) || (floor > NUMFLOORS) {
        fmt.Println("Invalid floor to get button signal")
    }
    if (getButtonType < 0) || (getButtonType > NUMBUTTONS) {
        fmt.Println("Invalid button type to get button signal")
    }
    if (floor == 0) && (getButtonType == ButtonType["Button call down"]) {
        fmt.Println("Invalid button type to get button signal")
    }
    if (floor == NUMFLOORS-1) && (getButtonType == ButtonType["Button call up"]) {
        fmt.Println("Invalid button type to get button signal")
    }

    if IOReadBit(buttonChannelsMatrix[floor][getButtonType]) != 0 {
        return true
    } else {
        return false
    }
}

func ElevatorGetFloorSensorSignal() int {

    if IOReadBit(SENSORFLOOR1) != 0 {
        return 0
    } else if IOReadBit(SENSORFLOOR2) != 0 {
        return 1
    } else if IOReadBit(SENSORFLOOR3) != 0 {
        return 2
    } else if IOReadBit(SENSORFLOOR4) != 0 {
        return 3
    } else {
        return -1
    }
}

func InitializeElevator() bool {

    var initSuccess bool = IOInitializeElevator()

    if !initSuccess {

        return false
    }

    for floor := 0; floor < NUMFLOORS; floor++ {
        for button := 0; button < NUMBUTTONS; button++ {

            if (button == ButtonType["Button call down"]) && (floor != 0) {
                ElevatorSetButtonLamp(button, floor, OFF)
            }

            if (button == ButtonType["Button call up"]) && (floor != 3) {
                ElevatorSetButtonLamp(button, floor, OFF)
            }
            if button == ButtonType["Button internal panel"] {
                ElevatorSetButtonLamp(button, floor, OFF)
            }
        }
    }

    ElevatorSetDoorOpenLamp(OFF)
    //ElevatorSetFloorIndicator(ElevatorGetFloorSensorSignal())
    return true
}
