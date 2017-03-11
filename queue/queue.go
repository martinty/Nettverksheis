package queue

import (
	"../driver"
	"../source"
	"encoding/json"
	"fmt"
	"os"
	//"time"
)

var orderTable [4][3]string
var externalOrders [4][2]string
var elevatorDirection int = -1
var elevatorFloor int = -1
var elevatorState int = -1
var elevatorID string

var infoTable map[string]source.ElevInfo

const (
	Idle = iota
	Running
	Stuck
)

const cost = 10

func Init() {
	infoTable = make(map[string]source.ElevInfo)
	var info source.ElevInfo = ReadFromFile()
	elevatorID = info.ID
	for floor := 0; floor < driver.NumFloors; floor++ {
		for button := 0; button < driver.NumButtons; button++ {
			if button == driver.ButtonTypeCommand {
				orderTable[floor][driver.ButtonTypeCommand] = info.LocalOrders[floor]
			} else {
				orderTable[floor][button] = info.ExternalOrders[floor][button]
			}
		}
	}
}

func Cost(floor int, button int) string {
	minCost := cost * 2 * driver.NumFloors
	minCostID := ""

	for key := range infoTable {

		tempCost := 0
		dir := 0
		buttonDirection := 0
		elevator := infoTable[key]
		floorNumber := elevator.CurrentFloor

		if elevator.CurrentDirection == driver.MotorDirectionUp {
			dir = 1
		} else {
			dir = -1
		}

		if button == driver.ButtonTypeUp {
			buttonDirection = 1
		} else {
			buttonDirection = -1
		}

		if elevator.State == Stuck {
			continue
		} else if elevator.State == Idle && floorNumber == floor {
			tempCost = 0
		} else {
			if floorNumber == driver.NumFloors-1 {
				dir = -1
			} else if floorNumber == 0 {
				dir = 1
			}
			for {
				if floorNumber == floor && dir == buttonDirection {
					break
				} else if floorNumber == driver.NumFloors-1 {
					if floorNumber == floor {
						break
					}
					dir = -1
				} else if floorNumber == 0 {
					if floorNumber == floor {
						break
					}
					dir = 1
				}
				floorNumber += dir
				tempCost += cost
			}
		}
		if tempCost < minCost {
			minCost = tempCost
			minCostID = key
		}
	}
	return minCostID
}

func UpdateOrders() {
	for floor := 0; floor < driver.NumFloors; floor++ {
		for button := 0; button < driver.NumButtons; button++ {
			if driver.ElevatorCheckButtonSignal(button, floor) {
				if orderTable[floor][button] == "" {
					if button == driver.ButtonTypeCommand {
						orderTable[floor][button] = elevatorID
					} else {
						ID := Cost(floor, button)
						externalOrders[floor][button] = ID
					}
				}
			}
		}
	}
}

func UpdateTables(msg source.ElevInfo) {
	infoTable[msg.ID] = msg
	fmt.Println(infoTable)
	for floor := 0; floor < driver.NumFloors; floor++ {
		for button := 0; button < driver.NumButtons-1; button++ {
			if orderTable[floor][button] == "" {
				orderTable[floor][button] = msg.ExternalOrders[floor][button]
			}
		}
	}
}

func UpdateElevatorInfo(msg source.ElevInfo) source.ElevInfo {
	for floor := 0; floor < driver.NumFloors; floor++ {
		for button := 0; button < driver.NumButtons; button++ {
			if button == driver.ButtonTypeCommand {
				msg.LocalOrders[floor] = orderTable[floor][button]
			} else {
				msg.ExternalOrders[floor][button] = externalOrders[floor][button]
			}
		}
	}
	msg.CurrentDirection = elevatorDirection
	msg.CurrentFloor = elevatorFloor
	msg.State = elevatorState
	return msg
}

func UpdateElevatorFloor(floor int) {
	elevatorFloor = floor
}

func UpdateElevatorDirection(direction int) {
	elevatorDirection = direction
}

func UpdateElevatorState(state int) {
	elevatorState = state
}

func DeleteOrder(button int, floor int) {
	orderTable[floor][button] = ""
}

func UpdateButtonLight() {
	for floor := 0; floor < driver.NumFloors; floor++ {
		for button := 0; button < driver.NumButtons; button++ {
			if orderTable[floor][button] != "" {
				driver.ElevatorSetButtonLamp(button, floor, true)
			} else {
				driver.ElevatorSetButtonLamp(button, floor, false)
			}
		}
	}
}

func GetMotorDirection(currentFloor int, currentDirection int) int {
	if currentFloor == driver.NumFloors {
		return driver.MotorDirectionDown
	} else if currentFloor == 0 {
		return driver.MotorDirectionUp
	} else if currentDirection == driver.MotorDirectionDown {
		for floor := 0; floor < currentFloor; floor++ {
			for button := 0; button < driver.NumButtons; button++ {
				if orderTable[floor][button] == elevatorID {
					return driver.MotorDirectionDown
				}
			}
		}
		return driver.MotorDirectionUp
	} else if currentDirection == driver.MotorDirectionUp {
		for floor := currentFloor + 1; floor < driver.NumFloors; floor++ {
			for button := 0; button < driver.NumButtons; button++ {
				if orderTable[floor][button] == elevatorID {
					return driver.MotorDirectionUp
				}
			}
		}
		return driver.MotorDirectionDown
	} else {
		return driver.MotorDirectionStop
	}
}

func ShouldElevatorStopAtFloor(currentFloor int, currentDirection int) bool {
	if orderTable[currentFloor][driver.ButtonTypeCommand] == elevatorID {
		return true
	} else if orderTable[currentFloor][currentDirection] == elevatorID {
		return true
	} else if currentFloor == driver.NumFloors && orderTable[currentFloor][driver.ButtonTypeDown] == elevatorID {
		return true
	} else if currentFloor == 0 && orderTable[currentFloor][driver.ButtonTypeUp] == elevatorID {
		return true
	} else if currentDirection == driver.MotorDirectionUp {
		for floor := currentFloor + 1; floor < driver.NumFloors; floor++ {
			for button := 0; button < driver.NumButtons; button++ {
				if orderTable[floor][button] == elevatorID {
					return false
				}
			}
		}
		if orderTable[currentFloor][driver.ButtonTypeDown] == elevatorID {
			return true
		}
	} else if currentDirection == driver.MotorDirectionDown {
		for floor := 0; floor < currentFloor; floor++ {
			for button := 0; button < driver.NumButtons; button++ {
				if orderTable[floor][button] == elevatorID {
					return false
				}
			}
		}
		if orderTable[currentFloor][driver.ButtonTypeUp] == elevatorID {
			return true
		}
	}
	return false
}

func ClearOrdersAtFloor(currentFloor int) {
	for button := 0; button < driver.NumButtons; button++ {
		orderTable[currentFloor][button] = ""
		if button != driver.ButtonTypeCommand {
			externalOrders[currentFloor][button] = ""
		}
	}
}

func CheckIfOrderTableIsEmpty() bool {
	for floor := 0; floor < driver.NumFloors; floor++ {
		for button := 0; button < driver.NumButtons; button++ {
			if orderTable[floor][button] == elevatorID {
				return false
			}
		}
	}
	return true
}

func CreateFile() bool {
	test, err := os.Open("file.txt")
	test.Close()
	if err != nil {
		file, err := os.Create("file.txt")
		file.Close()
		source.CheckForError(err)
		return true
	}
	return false
}

func WriteToFile(msg source.ElevInfo) {
	_ = os.Remove("file.txt")
	file, _ := os.Create("file.txt")
	file.Close()
	file, err := os.OpenFile("file.txt", os.O_WRONLY, 0666)
	source.CheckForError(err)

	buf, _ := json.Marshal(msg)
	_, err = file.Write(buf)
	source.CheckForError(err)

	file.Close()
}

func ReadFromFile() source.ElevInfo {
	file, err := os.Open("file.txt")
	source.CheckForError(err)

	data := make([]byte, 1024)
	count, err := file.Read(data)
	source.CheckForError(err)

	var msgFromFile source.ElevInfo

	err = json.Unmarshal(data[:count], &msgFromFile)
	source.CheckForError(err)

	file.Close()
	return msgFromFile
}

func DeleteFile() {
	_ = os.Remove("file.txt")
}
