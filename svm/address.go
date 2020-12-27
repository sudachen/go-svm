package svm

import "go-svm/common"

//const AddressLen = 20
//
//type Address [AddressLen]byte
//
//func BytesToAddress(b []byte) Address {
//	var addr Address
//	if len(b) <= AddressLen {
//		copy(addr[:], b)
//	} else {
//		copy(addr[:], b[:AddressLen])
//	}
//
//	return addr
//}

func svmByteArrayCloneToAddress(ba cSvmByteArray) Address {
	b := svmByteArrayCloneToBytes(ba)
	return common.BytesToAddress(b)
}
