package source

import (
	"fmt"
)

type ElevInfo struct {
	ID             string
	Words          string
	ElevIP         string
	CurrentFloor   int
	Direction      int
	NumberOrders   int
	LocalOrders    [4]int
	ExternalOrders [4][2]int
	Recipe         bool
	NewOrder       Order
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