package queue

import (
	"../driver"
	"../source"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

var orderTable [4][3]string
var elevatorDirection int = -1
var elevatorFloor int = -1
var elevatorState string = ""
var elevatorID string = ""
var infoTable map[string]source.ElevInfo
var offlineTimer map[string]int64
var orderTimer map[string]int64
var mutex sync.Mutex

const (
	idle     = "idle    "
	running  = "running "
	doorOpen = "doorOpen"
	stuck    = "stuck   "
)
const cost = 10

func Init(ID string) {
	infoTable = make(map[string]source.ElevInfo)
	offlineTimer = make(map[string]int64)
	orderTimer = make(map[string]int64)
	createFile()
	var backupFile source.ElevInfo = readFromFile()
	elevatorID = ID
	for floor := 0; floor < driver.NumFloors; floor++ {
		for button := 0; button < driver.NumButtons; button++ {
			if button == driver.ButtonTypeCommand {
				orderTable[floor][driver.ButtonTypeCommand] = backupFile.LocalOrders[floor]
			} else {
				orderTable[floor][button] = backupFile.ExternalOrders[floor][button]
			}
		}
	}
	updateButtonLight()
}

func CheckForOrdersAndDistribute(newOrderChan chan source.Order) {
	var newOrder source.Order
	for {
		for floor := 0; floor < driver.NumFloors; floor++ {
			for button := 0; button < driver.NumButtons; button++ {
				if driver.ElevatorCheckButtonSignal(button, floor) {
					if orderTable[floor][button] == "" {
						if button == driver.ButtonTypeCommand {
							orderTable[floor][button] = elevatorID
							updateButtonLight()
						} else {
							ID := calculateCost(floor, button)
							newOrder.Command = "add"
							newOrder.Floor = floor
							newOrder.Button = button
							newOrder.ElevID = ID
							newOrderChan <- newOrder
						}
					}
				}
			}
		}
	}
}

func UpdateElevatorInfoBeforeTransmitting(newMsgChanTransmit chan source.ElevInfo, newOrderChan chan source.Order, deleteOrderChan chan source.Order, deleteAddOrderChan chan source.Order, msg source.ElevInfo) {
	for {
		for floor := 0; floor < driver.NumFloors; floor++ {
			for button := 0; button < driver.NumButtons; button++ {
				if button == driver.ButtonTypeCommand {
					msg.LocalOrders[floor] = orderTable[floor][button]
				} else {
					msg.ExternalOrders[floor][button] = orderTable[floor][button]
				}
			}
		}
		msg.CurrentDirection = elevatorDirection
		msg.CurrentFloor = elevatorFloor
		msg.State = elevatorState
		select {
		case msg.NewOrder = <-newOrderChan:
		default:
			msg.NewOrder.Command = "none"
		}
		select {
		case msg.DeleteOrder = <-deleteOrderChan:
		default:
			break
		}
		select {
		case msg.DeleteAddOrder = <-deleteAddOrderChan:
		default:
			msg.DeleteAddOrder.Command = "none"
		}
		writeToFile(msg)
		newMsgChanTransmit <- msg
	}
}

func UpdateOrdersAfterReceiving(newMsgChanRecive chan source.ElevInfo, deleteAddOrderChan chan source.Order) {
	var recievedMsg source.ElevInfo
	for {
		recievedMsg = <-newMsgChanRecive
		if recievedMsg.NewOrder.Command == "add" {
			if orderTable[recievedMsg.NewOrder.Floor][recievedMsg.NewOrder.Button] == "" {
				orderTable[recievedMsg.NewOrder.Floor][recievedMsg.NewOrder.Button] = recievedMsg.NewOrder.ElevID
				updateButtonLight()
			}
		}
		if recievedMsg.DeleteOrder.Command == "delete" {
			ClearOrdersAtFloor(recievedMsg.DeleteOrder.Floor, recievedMsg.ID)
			updateButtonLight()
		}
		if recievedMsg.DeleteAddOrder.Command == "delete+add" {
			orderTable[recievedMsg.DeleteAddOrder.Floor][recievedMsg.DeleteAddOrder.Button] = recievedMsg.DeleteAddOrder.ElevID
			updateButtonLight()
		}

		mutex.Lock()
		infoTable[recievedMsg.ID] = recievedMsg
		fmt.Println(infoTable[elevatorID])
		mutex.Unlock()

		offlineTimer[recievedMsg.ID] = time.Now().Unix()
		for key := range offlineTimer {
			if time.Now().Unix()-offlineTimer[key] >= 2 {
				mutex.Lock()
				delete(infoTable, key)
				mutex.Unlock()
				redistributeOrders(key, deleteAddOrderChan)
			}
		}
		if recievedMsg.State == stuck {
			redistributeOrders(recievedMsg.ID, deleteAddOrderChan)
		}
	}
}

func redistributeOrders(ID string, deleteAddOrderChan chan source.Order) {
	var deleteAddOrder source.Order
	for floor := 0; floor < driver.NumFloors; floor++ {
		for button := 0; button < driver.NumButtons; button++ {
			if button != driver.ButtonTypeCommand {
				if orderTable[floor][button] == ID {
					newID := calculateCost(floor, button)
					deleteAddOrder.Command = "delete+add"
					deleteAddOrder.Floor = floor
					deleteAddOrder.Button = button
					deleteAddOrder.ElevID = newID
					deleteAddOrderChan <- deleteAddOrder
				}
			}
		}
	}
}

func calculateCost(floor int, button int) string {
	minCost := cost * 2 * driver.NumFloors
	minCostID := ""
	mutex.Lock()
	for key := range infoTable {
		tempCost := 0
		dir := 0
		buttonDirection := 0
		elevator := infoTable[key]
		floorNumber := elevator.CurrentFloor

		if elevator.State == stuck {
			continue
		} else if elevator.LocalOrders[floor] == key || elevator.ExternalOrders[floor][0] == key || elevator.ExternalOrders[floor][1] == key {
			minCost = 0
			minCostID = key
			break
		} else if floorNumber == floor {
			if elevator.State == idle || elevator.State == doorOpen {
				minCost = 0
				minCostID = key
				break
			}
		}

		if button == driver.ButtonTypeUp {
			buttonDirection = 1
		} else {
			buttonDirection = -1
		}

		if floorNumber == driver.NumFloors-1 {
			dir = -1
		} else if floorNumber == 0 {
			dir = 1
		} else if elevator.CurrentDirection == driver.MotorDirectionUp {
			dir = 1
		} else {
			dir = -1
		}

		if elevator.State == idle {
			tempCost = (elevator.CurrentFloor - floor) * cost
			if tempCost < 0 {
				tempCost = -tempCost
			}
		} else {
			if elevator.State == running {
				floorNumber += dir
				tempCost += cost
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
	mutex.Unlock()
	return minCostID
}

func UpdateElevatorFloor(floor int) {
	elevatorFloor = floor
}

func UpdateElevatorDirection(direction int) {
	elevatorDirection = direction
}

func UpdateElevatorState(state string) {
	elevatorState = state
}

func updateButtonLight() {
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

func ClearOrdersAtFloor(currentFloor int, ID string) {
	orderTable[currentFloor][driver.ButtonTypeUp] = ""
	orderTable[currentFloor][driver.ButtonTypeDown] = ""
	if ID == elevatorID || ID == "offline" {
		orderTable[currentFloor][driver.ButtonTypeCommand] = ""
	}
	updateButtonLight()
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

func createFile() bool {
	test, err := os.Open("backup.txt")
	test.Close()
	if err != nil {
		file, err := os.Create("backup.txt")
		file.Close()
		source.CheckForError(err)
		return true
	}
	return false
}

func writeToFile(msg source.ElevInfo) {
	_ = os.Remove("backup.txt")
	file, _ := os.Create("backup.txt")
	file.Close()
	file, err := os.OpenFile("backup.txt", os.O_WRONLY, 0666)
	source.CheckForError(err)
	buf, _ := json.Marshal(msg)
	_, err = file.Write(buf)
	source.CheckForError(err)
	file.Close()
}

func readFromFile() source.ElevInfo {
	file, err := os.Open("backup.txt")
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
