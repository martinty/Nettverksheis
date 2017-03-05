package queue

import (
	"../driver"
	"../source"
	"encoding/json"
	"fmt"
	"os"
)

var orderTable[4][3] int

func Init() {
	fmt.Println("heis")
}

func UpdateOrders() {
	for floor := 0; floor < driver.NumFloors; floor++ {
        for button := 0; button < driver.NumButtons; button++ {
        	if driver.ElevatorCheckButtonSignal(button, floor){
        		if orderTable[floor][button] == 0{
        			orderTable[floor][button] = 1
        			fmt.Println(orderTable)
        		}
        	}
        }
    }
}

func DeleteOrder(button int, floor int) {
	orderTable[floor][button] = 0
}

func GetMotorDirection(currentFloor int, currentDirection int) int{
	if currentFloor == driver.NumFloors{
		return driver.MotorDirectionDown
	}else if currentFloor == 0{
		return driver.MotorDirectionUp
	} else if currentDirection == driver.MotorDirectionDown {
		for floor := 0; floor < currentFloor; floor++ {
        	for button := 0; button < driver.NumButtons; button++ {
        		if orderTable[floor][button] == 1{
        			return driver.MotorDirectionDown
        		}
        	}
        }
        return driver.MotorDirectionUp		
	} else if currentDirection == driver.MotorDirectionUp {
		for floor := currentFloor + 1; floor < driver.NumFloors; floor++ {
        	for button := 0; button < driver.NumButtons; button++ {
        		if orderTable[floor][button] == 1{
        			return driver.MotorDirectionUp
        		}
        	}
        }
        return driver.MotorDirectionDown				
	}else{
		return driver.MotorDirectionStop
	}
}

func ShouldElevatorStopAtFloor(currentFloor int, currentDirection int) bool {
	if orderTable[currentFloor][driver.ButtonTypeCommand] == 1{
		return true
	} else if orderTable[currentFloor][currentDirection] == 1 {
		return true
	} else if currentFloor == driver.NumFloors && orderTable[currentFloor][driver.ButtonTypeDown] == 1 {
		return true
	} else if currentFloor == 0 && orderTable[currentFloor][driver.ButtonTypeUp] == 1 {
		return true
	} else if currentDirection == driver.MotorDirectionUp{
		for floor := currentFloor + 1; floor < driver.NumFloors; floor++ {
        	for button := 0; button < driver.NumButtons; button++ {
        		if orderTable[floor][button] == 1{
        			return false
        		}
        	}
        }
        if orderTable[currentFloor][driver.ButtonTypeDown] == 1{
        	return true
        }
	} else if currentDirection == driver.MotorDirectionDown{
		for floor := 0; floor < currentFloor; floor++ {
        	for button := 0; button < driver.NumButtons; button++ {
        		if orderTable[floor][button] == 1{
        			return false
        		}
        	}
        }
        if orderTable[currentFloor][driver.ButtonTypeUp] == 1{
        	return true
        }
    }
    return false
}

func ClearOrdersAtFloor(currentFloor int) {
	for button := 0; button < driver.NumButtons; button++ {
		orderTable[currentFloor][button] = 0
	}
}

func CheckIfOrderTableIsEmpty()bool{
	for floor := 0; floor < driver.NumFloors; floor++ {
    	for button := 0; button < driver.NumButtons; button++ {
    		if orderTable[floor][button] != 0{
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
