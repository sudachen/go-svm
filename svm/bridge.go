package svm

// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR} -lsvm_runtime_c_api
// #include "./svm.h"
// #include <string.h>
//
// extern int svm_trampoline();
import "C"
import (
	"unsafe"
)

//unsafe extern "C" fn trampoline(
//env: *mut svm_env_t,
//params: *const svm_byte_array,
//results: *mut svm_byte_array,
//) -> *mut svm_byte_array {

//export svm_trampoline
func svm_trampoline() C.int {
	return 3
}

type cUchar = C.uchar
type cUint = C.uint
type cSvmByteArray = C.svm_byte_array
type cSvmResultT = C.svm_result_t

const cSvmSuccess = (C.svm_result_t)(C.SVM_SUCCESS)

func cSvmImportsAlloc(imports *unsafe.Pointer, count uint) cSvmResultT {
	return (cSvmResultT)(C.svm_imports_alloc(imports, C.uint(count)))
}

func cSvmImportFuncNew(
	imports Imports,
	namespace string,
	name string,
	importFunction ImportFunction,
) error {
	cImports := imports.p
	cNamespace := bytesCloneToSvmByteArray([]byte(namespace))
	cImportName := bytesCloneToSvmByteArray([]byte(name))
	cParams := bytesCloneToSvmByteArray(importFunction.params.Encode())
	cReturns := bytesCloneToSvmByteArray(importFunction.returns.Encode())
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
		unsafe.Pointer(&importFunction.env),
		cParams,
		cReturns,
		&cErr,
	); res != cSvmSuccess {
		return cErr.svmError()
	}

	return nil
}

func cSvmMemoryRuntimeCreate(runtime *unsafe.Pointer, kv, host, imports unsafe.Pointer) error {
	err := cSvmByteArray{}
	defer err.SvmFree()

	if res := C.svm_memory_runtime_create(
		runtime,
		kv,
		imports,
		&err,
	); res != cSvmSuccess {
		return err.svmError()
	}

	return nil
}

func cSvmMemoryKVCreate(p *unsafe.Pointer) cSvmResultT {
	return (cSvmResultT)(C.svm_memory_state_kv_create(p))
}

func cSvmEncodeAppTemplate(version int, name string, code []byte, dataLayout DataLayout) ([]byte, error) {
	appTemplate := cSvmByteArray{}
	cVersion := C.uint(version)
	cName := bytesCloneToSvmByteArray([]byte(name))
	cCode := bytesCloneToSvmByteArray(code)
	cDataLayout := bytesCloneToSvmByteArray(dataLayout.Encode())
	err := cSvmByteArray{}

	defer func() {
		appTemplate.SvmFree()
		cName.Free()
		cCode.Free()
		cDataLayout.Free()
		err.SvmFree()
	}()

	if res := C.svm_encode_app_template(
		&appTemplate,
		cVersion,
		cName,
		cCode,
		cDataLayout,
		&err,
	); res != cSvmSuccess {
		return nil, err.svmError()
	}

	return svmByteArrayCloneToBytes(appTemplate), nil
}

func cSvmValidateTemplate(runtime Runtime, appTemplate []byte) error {
	cRuntime := runtime.p
	cAppTemplate := bytesCloneToSvmByteArray(appTemplate)
	cErr := cSvmByteArray{}

	defer func() {
		cAppTemplate.Free()
		cErr.SvmFree()
	}()

	if res := C.svm_validate_template(
		cRuntime,
		cAppTemplate,
		&cErr,
	); res != cSvmSuccess {
		return cErr.svmError()
	}

	return nil
}

func cSvmDeployTemplate(runtime Runtime, appTemplate []byte, author Address, gasMetering bool, gasLimit uint64) ([]byte, error) {
	cReceipt := cSvmByteArray{}
	cRuntime := runtime.p
	cAppTemplate := bytesCloneToSvmByteArray(appTemplate)
	cAuthor := bytesCloneToSvmByteArray(author[:])
	cGasMetering := C.bool(gasMetering)
	cGasLimit := C.uint64_t(gasLimit)
	cErr := cSvmByteArray{}

	defer func() {
		cReceipt.SvmFree()
		cAppTemplate.Free()
		cAuthor.Free()
		cErr.SvmFree()
	}()

	if res := C.svm_deploy_template(
		&cReceipt,
		cRuntime,
		cAppTemplate,
		cAuthor,
		cGasMetering,
		cGasLimit,
		&cErr,
	); res != cSvmSuccess {
		return nil, cErr.svmError()
	}

	return svmByteArrayCloneToBytes(cReceipt), nil
}

//func cSvmTemplateReceiptAddr(receipt []byte) (Address, error) {
//	cTemplateAddr := cSvmByteArray{}
//	cReceipt := bytesCloneToSvmByteArray(receipt)
//	cErr := cSvmByteArray{}
//
//	defer func() {
//		cTemplateAddr.SvmFree()
//		cReceipt.Free()
//		cErr.SvmFree()
//	}()
//
//	if res := C.svm_template_receipt_addr(
//		&cTemplateAddr,
//		cReceipt,
//		&cErr,
//	); res != cSvmSuccess {
//		return Address{}, cErr.svmError()
//	}
//
//	return svmByteArrayCloneToAddress(cTemplateAddr), nil
//}

//func cSvmTemplateReceiptGas(receipt []byte) (uint64, error) {
//	var cGasUsed C.uint64_t
//	cReceipt := bytesCloneToSvmByteArray(receipt)
//	cErr := cSvmByteArray{}
//
//	defer func() {
//		cReceipt.Free()
//		cErr.SvmFree()
//	}()
//
//	if res := C.svm_template_receipt_gas(
//		&cGasUsed,
//		cReceipt,
//		&cErr,
//	); res != cSvmSuccess {
//		return 0, cErr.svmError()
//	}
//
//	return uint64(cGasUsed), nil
//}

func cSvmSpawnApp(runtime Runtime, spawnApp []byte, creator Address, gasMetering bool, gasLimit uint64) ([]byte, error) {
	cReceipt := cSvmByteArray{}
	cRuntime := runtime.p
	cSpawnApp := bytesCloneToSvmByteArray(spawnApp)
	cCreator := bytesCloneToSvmByteArray(creator[:])
	cGasMetering := C.bool(gasMetering)
	cGasLimit := C.uint64_t(gasLimit)
	cErr := cSvmByteArray{}

	defer func() {
		cReceipt.SvmFree()
		cSpawnApp.Free()
		cCreator.Free()
		cErr.SvmFree()
	}()

	if res := C.svm_spawn_app(
		&cReceipt,
		cRuntime,
		cSpawnApp,
		cCreator,
		cGasMetering,
		cGasLimit,
		&cErr,
	); res != cSvmSuccess {
		return nil, cErr.svmError()
	}

	return svmByteArrayCloneToBytes(cReceipt), nil
}

//func cSvmAppReceiptState(receipt []byte) ([]byte, error) {
//	cInitialState := cSvmByteArray{}
//	cReceipt := bytesCloneToSvmByteArray(receipt)
//	cErr := cSvmByteArray{}
//
//	defer func() {
//		cInitialState.SvmFree()
//		cReceipt.Free()
//		cErr.SvmFree()
//	}()
//
//	if res := C.svm_app_receipt_state(
//		&cInitialState,
//		cReceipt,
//		&cErr,
//	); res != cSvmSuccess {
//		return nil, cErr.svmError()
//	}
//
//	return svmByteArrayCloneToBytes(cInitialState), nil
//}

//func cSvmAppReceiptAddr(receipt []byte) (Address, error) {
//	cAppAddr := cSvmByteArray{}
//	cReceipt := bytesCloneToSvmByteArray(receipt)
//	cErr := cSvmByteArray{}
//
//	defer func() {
//		cAppAddr.SvmFree()
//		cReceipt.Free()
//		cErr.SvmFree()
//	}()
//
//	if res := C.svm_app_receipt_addr(
//		&cAppAddr,
//		cReceipt,
//		&cErr,
//	); res != cSvmSuccess {
//		return Address{}, cErr.svmError()
//	}
//
//	return svmByteArrayCloneToAddress(cAppAddr), nil
//}

//func cSvmAppReceiptGas(receipt []byte) (uint64, error) {
//	var cGasUsed C.uint64_t
//	cReceipt := bytesCloneToSvmByteArray(receipt)
//	cErr := cSvmByteArray{}
//
//	defer func() {
//		cReceipt.Free()
//		cErr.SvmFree()
//	}()
//
//	if res := C.svm_app_receipt_gas(
//		&cGasUsed,
//		cReceipt,
//		&cErr,
//	); res != cSvmSuccess {
//		return 0, cErr.svmError()
//	}
//
//	return uint64(cGasUsed), nil
//}

func cSvmEncodeSpawnApp(version int, templateAddr Address, name string, ctorName string, calldata []byte) ([]byte, error) {
	spawnApp := cSvmByteArray{}
	cVersion := C.uint(version)
	cTemplateAddr := bytesCloneToSvmByteArray(templateAddr[:])
	cName := bytesCloneToSvmByteArray([]byte(ctorName))
	cCtorName := bytesCloneToSvmByteArray([]byte(ctorName))
	cCalldata := bytesCloneToSvmByteArray(calldata)
	cErr := cSvmByteArray{}

	defer func() {
		spawnApp.SvmFree()
		cTemplateAddr.Free()
		cErr.SvmFree()
	}()

	if res := C.svm_encode_spawn_app(
		&spawnApp,
		cVersion,
		cTemplateAddr,
		cName,
		cCtorName,
		cCalldata,
		&cErr,
	); res != cSvmSuccess {
		return nil, cErr.svmError()
	}

	return svmByteArrayCloneToBytes(spawnApp), nil
}

func cSvmValidateApp(runtime Runtime, app []byte) error {
	cRuntime := runtime.p
	cApp := bytesCloneToSvmByteArray(app)
	cErr := cSvmByteArray{}

	defer func() {
		cApp.Free()
		cErr.SvmFree()
	}()

	if res := C.svm_validate_app(
		cRuntime,
		cApp,
		&cErr,
	); res != cSvmSuccess {
		return cErr.svmError()
	}

	return nil
}

func cSvmEncodeAppTx(
	version int,
	AppAddr Address,
	funcName string,
	calldata []byte,
) ([]byte, error) {
	appTx := cSvmByteArray{}
	cVersion := C.uint(version)
	cTemplateAddr := bytesCloneToSvmByteArray(AppAddr[:])
	cFuncName := bytesCloneToSvmByteArray([]byte(funcName))
	cCalldata := bytesCloneToSvmByteArray(calldata)
	cErr := cSvmByteArray{}

	defer func() {
		appTx.SvmFree()
		cTemplateAddr.Free()
		cFuncName.Free()
		cCalldata.Free()
		cErr.SvmFree()
	}()

	if res := C.svm_encode_app_tx(
		&appTx,
		cVersion,
		cTemplateAddr,
		cFuncName,
		cCalldata,
		&cErr,
	); res != cSvmSuccess {
		return nil, cErr.svmError()
	}

	return svmByteArrayCloneToBytes(appTx), nil
}

func cSvmValidateTx(runtime Runtime, appTx []byte) (Address, error) {
	cAppAddr := cSvmByteArray{}
	cRuntime := runtime.p
	cAppTx := bytesCloneToSvmByteArray(appTx)
	cErr := cSvmByteArray{}

	defer func() {
		cAppAddr.SvmFree()
		cAppTx.Free()
		cErr.SvmFree()
	}()

	if res := C.svm_validate_tx(
		&cAppAddr,
		cRuntime,
		cAppTx,
		&cErr,
	); res != cSvmSuccess {
		return Address{}, cErr.svmError()
	}

	return svmByteArrayCloneToAddress(cAppAddr), nil
}

func cSvmExecApp(runtime Runtime, appTx []byte, appState []byte, gasMetering bool,
	gasLimit uint64) ([]byte, error) {
	cReceipt := cSvmByteArray{}
	cRuntime := runtime.p
	cAppTx := bytesCloneToSvmByteArray(appTx)
	cAppState := bytesCloneToSvmByteArray(appState)
	cGasMetering := C.bool(gasMetering)
	cGasLimit := C.uint64_t(gasLimit)
	cErr := cSvmByteArray{}

	defer func() {
		cReceipt.SvmFree()
		cAppTx.Free()
		cAppState.Free()
		cErr.SvmFree()
	}()

	if res := C.svm_exec_app(
		&cReceipt,
		cRuntime,
		cAppTx,
		cAppState,
		cGasMetering,
		cGasLimit,
		&cErr,
	); res != cSvmSuccess {
		return nil, cErr.svmError()
	}

	return svmByteArrayCloneToBytes(cReceipt), nil
}

//func cSvmExecReceiptState(receipt []byte) ([]byte, error) {
//	cNewState := cSvmByteArray{}
//	cReceipt := bytesCloneToSvmByteArray(receipt)
//	cErr := cSvmByteArray{}
//
//	defer func() {
//		cNewState.SvmFree()
//		cReceipt.Free()
//		cErr.SvmFree()
//	}()
//
//	if res := C.svm_exec_receipt_state(
//		&cNewState,
//		cReceipt,
//		&cErr,
//	); res != cSvmSuccess {
//		return nil, cErr.svmError()
//	}
//
//	return svmByteArrayCloneToBytes(cNewState), nil
//}

//func cSvmExecReceiptReturns(receipt []byte) (Values, error) {
//	cReturns := cSvmByteArray{}
//	cReceipt := bytesCloneToSvmByteArray(receipt)
//	cErr := cSvmByteArray{}
//
//	defer func() {
//		cReturns.SvmFree()
//		cReceipt.Free()
//		cErr.SvmFree()
//	}()
//
//	if res := C.svm_exec_receipt_returns(
//		&cReturns,
//		cReceipt,
//		&cErr,
//	); res != cSvmSuccess {
//		return nil, cErr.svmError()
//	}
//
//	var nativeReturns Values
//	if err := (&nativeReturns).Decode(svmByteArrayCloneToBytes(cReturns)); err != nil {
//		return nil, fmt.Errorf("failed to decode returns: %v", err)
//	}
//
//	return nativeReturns, nil
//}

//func cSvmExecReceiptGas(receipt []byte) (uint64, error) {
//	var cGasUsed C.uint64_t
//	cReceipt := bytesCloneToSvmByteArray(receipt)
//	cErr := cSvmByteArray{}
//
//	defer func() {
//		cReceipt.Free()
//		cErr.SvmFree()
//	}()
//
//	if res := C.svm_exec_receipt_gas(
//		&cGasUsed,
//		cReceipt,
//		&cErr,
//	); res != cSvmSuccess {
//		return 0, cErr.svmError()
//	}
//
//	return uint64(cGasUsed), nil
//}

func cSvmInstanceContextHostGet(ctx unsafe.Pointer) unsafe.Pointer {
	return nil //C.svm_instance_context_host_get(ctx)
}

func cSvmByteArrayDestroy(ba cSvmByteArray) {
	C.svm_byte_array_destroy(ba)
}

func cSvmRuntimeDestroy(runtime Runtime) {
	C.svm_runtime_destroy(runtime.p)
}

func cSvmImportsDestroy(imports Imports) {
	C.svm_imports_destroy(imports.p)
}

func cSvmMemKVDestroy(kv MemKVStore) {
	C.svm_state_kv_destroy(kv.p)
}

func cFree(p unsafe.Pointer) {
	C.free(p)
}
