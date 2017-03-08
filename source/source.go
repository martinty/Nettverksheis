package source

import (
	"fmt"
)

const Elevators int = 3

type ElevInfo struct {
	IP               string
	ID               string
	CurrentFloor     int
	CurrentDirection int
	State            int
	LocalOrders      [4]int
	ExternalOrders   [4][2]int
	Recipe           bool
	//NewOrder       Order
}

type Order struct {
	ID       string
	Words    string
	Position [2]int
	ElevIP   string
}

func CheckForError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}
