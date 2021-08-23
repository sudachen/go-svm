package svm

/*
#cgo LDFLAGS: -lsvm_runtime_ffi
#include "svm.h"
#include "memory.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type byteArray = C.svm_byte_array

type ByteArray struct {
	byteArray
}

func (ba *ByteArray) FromBytes(bs []byte) error {
	if ba.byteArray.capacity < C.uint(len(bs)) {
		return fmt.Errorf("bytearray is too small, required %v bytes but just %v is available", len(bs), ba.byteArray.capacity)
	}
	C.memcpy(unsafe.Pointer(ba.byteArray.bytes),unsafe.Pointer(&bs[0]),C.ulong(len(bs)))
	ba.byteArray.length = C.uint(len(bs))
	return nil
}

func (ba *ByteArray) Bytes() []byte {
	return C.GoBytes(unsafe.Pointer(ba.byteArray.bytes), C.int(ba.byteArray.length))
}

func (ba *ByteArray) Destroy() {
	if ba.byteArray.capacity != 0 {
		C.svm_byte_array_destroy(ba.byteArray)
		ba.byteArray.capacity = 0
		ba.byteArray.length = 0
	}
}

func (ba *ByteArray) Close() error {
	ba.Destroy()
	return nil
}

