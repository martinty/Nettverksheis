package queue

import (
	"../driver"
	"../source"
	"encoding/json"
	"fmt"
	"os"
)

var msg source.ElevInfo

func Init(queueMsg source.ElevInfo) {
	queueMsg.LocalOrders[1] = 1
	fmt.Println(queueMsg.LocalOrders)
}

func UpdateMsg() {
	var buttonUpdate source.ElevInfo
	for {
		buttonUpdate = driver.ElevatorGetButtonSignal(buttonUpdate)
	}
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
