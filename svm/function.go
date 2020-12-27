package svm

import "C"

type hostFunction func([]Value) ([]Value, error)

type functionEnvironment struct {
	function hostFunction
}

type Function struct {
	environment *functionEnvironment
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
