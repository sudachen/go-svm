package svm

import "C"

// #include "./svm.h"
//
// extern svm_byte_array* svm_trampoline(svm_env_t *env, svm_byte_array *args, svm_byte_array *results);
//
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

//export svm_trampoline
func svm_trampoline(env *C.svm_env_t, args *C.svm_byte_array, results *C.svm_byte_array) *C.svm_byte_array {
	goArgs := Values{}
	if err := goArgs.Decode(svmByteArrayCloneToBytes(*args)); err != nil {
		panic(err)
	}

	hostEnv := (*functionEnvironment)(env.host_env)
	f := hostFunctionStore.get(hostEnv.hostFunctionStoreIndex)
	if f == nil {
		panic(fmt.Sprintf("go-svm: host import function not found; index: %v", hostEnv.hostFunctionStoreIndex))
	}

	goResults, err := f(goArgs)
	if err != nil {
		err := []byte(err.Error())
		cErr := bytesAliasToSvmByteArray(err)

		// Re-allocate the error on SVM side, so that it
		// would be able do de-allocate it once processing is done.
		cSvmErr := cSvmWasmErrorEreate(cErr)
		runtime.KeepAlive(err)

		return cSvmErr
	}

	rawResults := Values(goResults).Encode()
	*results = bytesCloneToSvmByteArray(rawResults)

	return nil
}

func cSvmWasmErrorEreate(err cSvmByteArray) *cSvmByteArray {
	return (*cSvmByteArray)(C.svm_wasm_error_create(err))
}

func cSvmImportFuncNew(
	imports Imports,
	namespace string,
	name string,
	env unsafe.Pointer,
	params ValueTypes,
	returns ValueTypes,
) error {
	cImports := imports._inner
	cNamespace := bytesCloneToSvmByteArray([]byte(namespace))
	cImportName := bytesCloneToSvmByteArray([]byte(name))
	cParams := bytesCloneToSvmByteArray(params.Encode())
	cReturns := bytesCloneToSvmByteArray(returns.Encode())
	cErr := cSvmByteArray{}

	defer func() {
		cNamespace.Free()
		cImportName.Free()
		cParams.Free()
		cReturns.Free()
		cErr.SvmFree()
	}()

	if res := C.svm_import_func_new(
		cImports,
		cNamespace,
		cImportName,
		(C.svm_func_callback_t)(C.svm_trampoline),
		env,
		cParams,
		cReturns,
		&cErr,
	); res != cSvmSuccess {
		return cErr.svmError()
	}

	return nil
}

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

func (hf *hostFunctions) get(index uint) hostFunction {
	return hf.functions[index]
}

func (hf *hostFunctions) set(function hostFunction) uint {
	hf.Lock()
	defer hf.Unlock()

	index := uint(len(hf.functions))
	hf.functions[index] = function

	return index
}
