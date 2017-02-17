package driver  // where "driver" is the folder that contains io.go, io.c, io.h, channels.go, channels.h and driver.go
/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
*/
import "C"



func IOInitializeElevator() bool{

	initstatus := int(C.io_init())
	return bool(initstatus != 0)
}

func IOSetBit(channel int){

	C.io_set_bit(C.int(channel))
}

func IOClearBit(channel int){
	C.io_clear_bit(C.int(channel))
}

func IOWriteAnalog(channel int, value int){

	C.io_write_analog(C.int(channel),C.int(value))
}

func IOReadBit(channel int) int{

	return int(C.io_read_bit(C.int(channel)))
}

func IOReadAnalog(channel int) int{

	return int(C.io_read_analog(C.int(channel)))
}
