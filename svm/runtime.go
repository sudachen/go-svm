package svm

/*
#include "svm.h"
#include "memory.h"
*/
import "C"
import "unsafe"


type Runtime struct {
	svmRuntime unsafe.Pointer
}

func NewRuntime() (*Runtime, error) {
	rt := &Runtime{}
	err := Error{}
	if res := C.svm_memory_runtime_create(&rt.svmRuntime,err.ptr()); res != C.SVM_SUCCESS {
		defer err.Destroy()
		return nil, err.ToError("can't create new SVM runtime: ")
	}
	return rt, nil
}

func (rt *Runtime) Destroy() {
	if rt.svmRuntime != nil {
		C.svm_runtime_destroy(rt.svmRuntime)
		rt.svmRuntime = nil
	}
}

func (rt *Runtime) Close() error {
	rt.Destroy()
	return nil
}

func (rt *Runtime) ValidateCall(msg *Message) error {
	err := Error{}
	if res := C.svm_validate_call(rt.svmRuntime,msg.byteArray,err.ptr()); res != C.SVM_SUCCESS {
		defer err.Destroy()
		return err.ToError("failed to validate call message: ")
	}
	return nil
}

func (rt *Runtime) Call(envelope *Envelope, msg *Message, ctx *Context) (*Receipt, error) {
	var err Error
	rcpt := &ByteArray{}
	if res := C.svm_validate_call(rt.svmRuntime, msg.byteArray, err.ptr()); res != C.SVM_SUCCESS {
		defer err.Destroy()
		return nil, err.ToError("failed to validate call message: ")
	}
	if res := C.svm_call(&rcpt.byteArray, rt.svmRuntime, envelope.byteArray, msg.byteArray, ctx.byteArray, err.ptr()); res != C.SVM_SUCCESS {
		defer err.Destroy()
		return nil, err.ToError("failed to validate call message: ")
	}
	return nil, nil
}

func (rt *Runtime) ValidateSpawn(msg *Message) error {
	err := Error{}
	if res := C.svm_validate_call(rt.svmRuntime,msg.byteArray,err.ptr()); res != C.SVM_SUCCESS {
		defer err.Destroy()
		return err.ToError("failed to call contract: ")
	}
	return nil
}

func (rt *Runtime) Spawn(envelope *Envelope, msg *Message, ctx *Context) (*Receipt, error) {
	var err Error
	rcpt := &ByteArray{}
	if res := C.svm_validate_spawn(rt.svmRuntime, msg.byteArray, err.ptr()); res != C.SVM_SUCCESS {
		defer err.Destroy()
		return nil, err.ToError("failed to validate call message: ")
	}
	if res := C.svm_spawn(&rcpt.byteArray, rt.svmRuntime, envelope.byteArray, msg.byteArray, ctx.byteArray, err.ptr()); res != C.SVM_SUCCESS {
		defer err.Destroy()
		return nil, err.ToError("failed spawn contract: ")
	}
	return nil, nil
}
