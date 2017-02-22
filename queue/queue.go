package queue

import (
	"../source"
	"fmt"
)

func Init(queueMsg source.ElevInfo) {
	queueMsg.LocalOrders[1] = 1
	fmt.Println(queueMsg.LocalOrders)

}
