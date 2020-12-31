package common

import (
	"encoding/hex"
)

const AddressLen = 20

type Address [AddressLen]byte

func (addr Address) String() string {
	return hex.EncodeToString(addr[:])
}

func BytesToAddress(b []byte) Address {
	var addr Address
	if len(b) <= AddressLen {
		copy(addr[:], b)
	} else {
		copy(addr[:], b[:AddressLen])
	}

	return addr
}
