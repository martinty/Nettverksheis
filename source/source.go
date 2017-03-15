package source

import (
	"fmt"
)

type ElevInfo struct {
	ID               string
	CurrentFloor     int
	CurrentDirection int
	State            string
	LocalOrders      [4]string
	ExternalOrders   [4][2]string
	NewOrder         Order
	DeleteOrder      Order
	DeleteAddOrder   Order
}

type Order struct {
	Command string
	Floor   int
	Button  int
	ElevID  string
}

func CheckForError(err error) bool {
	if err != nil {
		fmt.Println("Error:", err)
		return true
	}
	return false
}
