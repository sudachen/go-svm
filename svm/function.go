package svm

import "C"
import (
	"fmt"
	"sync"
)

type functionEnvironment struct {
	hostFunctionStoreIndex uint
}

type Function struct {
	environment *functionEnvironment
}

type hostFunction func([]Value) ([]Value, error)

var hostFunctionStore = hostFunctions{
	functions: make(map[uint]hostFunction),
}

type hostFunctions struct {
	sync.RWMutex
	functions map[uint]hostFunction
}

func (self *hostFunctions) load(index uint) (hostFunction, error) {
	hostFunction, exists := self.functions[index]

	if exists && hostFunction != nil {
		return hostFunction, nil
	}

	return nil, fmt.Errorf("host function `%d` does not exist", index)
}

func (self *hostFunctions) store(function hostFunction) uint {
	self.Lock()
	// By default, the index is the size of the store.
	index := uint(len(self.functions))

	for nth, hostFunc := range self.functions {
		// Find the first empty slot in the store.
		if hostFunc == nil {
			// Use that empty slot for the index.
			index = nth
			break
		}
	}

	self.functions[index] = function
	self.Unlock()

	return index
}

func (self *hostFunctions) remove(index uint) {
	self.Lock()
	self.functions[index] = nil
	self.Unlock()
}

//func NewFunction(function hostFunction, ty FunctionType) *Function {
//
//}
//
//
//func NewFunction(name string, params ValueTypes, returns ValueTypes, impl hostFunction) (ImportsBuilder, error) {
//	env := functionEnvironment{
//		function: impl,
//	}
//
//	ib.imports[name] = ImportFunction{
//		namespace,
//		env,
//		params,
//		returns,
//	}
//
//	return ib, nil
//}

//
//func NewFunction(function hostFunction, ty FunctionType) *Function {
//	hostFunction := &hostFunction{
//		store:    store,
//		function: function,
//	}
//	environment := &FunctionEnvironment{
//		hostFunctionStoreIndex: hostFunctionStore.store(hostFunction),
//	}
//	pointer := C.wasm_func_new_with_env(
//		store.inner(),
//		ty.inner(),
//		(C.wasm_func_callback_t)(C.function_trampoline),
//		unsafe.Pointer(environment),
//		(C.wasm_func_callback_env_finalizer_t)(C.function_environment_finalizer),
//	)
//
//	runtime.KeepAlive(environment)
//
//	return newFunction(pointer, environment, nil)
//}
