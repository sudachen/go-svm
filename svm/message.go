package svm

/*
#include "svm.h"
*/
import "C"
import (
	"encoding/binary"
	"errors"
)

var errMessageTooLong = errors.New("message too long")

type Message struct {
	ByteArray
}

func NewMessage(length int) *Message {
	m := &Message{}
	m.byteArray = C.svm_message_alloc(C.uint(length))
	return m
}

func NewMessageFromBytes(bs []byte) *Message {
	m := NewMessage(len(bs))
	m.FromBytes(bs)
	return m
}

type CallMessage struct {
	Version uint16
	Target [AddressLength]byte
	FuncName string
	VerifyData []byte
	CallData []byte
}

func (cm CallMessage) Encode() (*Message, error) {
	p := 0
	fn := []byte(cm.FuncName)
	bs := make([]byte,2+AddressLength+1+len(fn)+1+len(cm.VerifyData)+1+len(cm.CallData))
	appendBytes := func(b []byte) error {
		ln := len(b)
		if ln > 255 { return errMessageTooLong }
		bs[p] = uint8(ln)
		copy(bs[p+1:p+1+ln],b)
		p += 1+ln
		return nil
	}
	binary.BigEndian.PutUint16(bs[p:p+2],cm.Version)
	p += 2
	copy(bs[p:p+AddressLength],cm.Target[:])
	p += AddressLength
	if err := appendBytes(fn); err != nil { return nil, err }
	if err := appendBytes(cm.VerifyData); err != nil { return nil, err }
	if err := appendBytes(cm.CallData); err != nil { return nil, err }
	return NewMessageFromBytes(bs), nil
}

type SpawnMessage struct {
	Version uint16
	Template [AddressLength]byte
	FuncName string
	Ctor string
	CallData []byte
}

func (cm SpawnMessage) Encode() (*Message, error) {
	p := 0
	fn := []byte(cm.FuncName)
	ct := []byte(cm.Ctor)
	bs := make([]byte,2+AddressLength+1+len(fn)+1+len(ct)+1+len(cm.CallData))
	appendBytes := func(b []byte) error {
		ln := len(b)
		if ln > 255 { return errMessageTooLong }
		bs[p] = uint8(ln)
		copy(bs[p+1:p+1+ln],b)
		p += 1+ln
		return nil
	}
	binary.BigEndian.PutUint16(bs[p:p+2],cm.Version)
	p += 2
	copy(bs[p:p+AddressLength],cm.Template[:])
	p += AddressLength
	if err := appendBytes(fn); err != nil { return nil, err }
	if err := appendBytes(ct); err != nil { return nil, err }
	if err := appendBytes(cm.CallData); err != nil { return nil, err }
	return NewMessageFromBytes(bs), nil
}

type DeployHeaderSection struct {
	Name, Desc string
	CodeVerison uint32
}

type DeployCodeSection struct {
	// Kind == Wasm
	SvmVersion uint32
	Code []byte
	// Flags uint64 == ExeceFlags (0x01)
	// GasMode == Fixed
}