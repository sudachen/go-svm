package main

import (
	"fmt"
	"go-svm/codec"
	"go-svm/svm"
	"io/ioutil"
	"runtime"
	"unsafe"
)

// Declare `inc` and `get` C function signatures in the cgo preamble.
// Doing so is required for SVM to be able to invoke their Go implementation.

// #include <stdlib.h>
//
// extern void inc(void *ctx, int value);
// extern int get(void *ctx);
import "C"

//Define `inc` and `get` Go implementation.
//The first argument is the runtime context, and must be included by all import functions.
//Notice the `//export` comment which is the way cgo uses to map Go code to C code.
//
func inc(args []svm.Value) ([]svm.Value, error) {
	closure.value += args[0].ToI32()
	return []svm.Value{}, nil
}

func give(_ []svm.Value) ([]svm.Value, error) {
	return []svm.Value{svm.I32(host.value)}, nil
}

type counter struct {
	value int32
}

var closure counter

var host counter

func main() {
	// Initialize svmRuntime.
	ib := svm.NewImportsBuilder()

	ib, err := ib.AppendFunction(
		"counter_mul",
		svm.ValueTypes{svm.TypeI32, svm.TypeI32},
		svm.ValueTypes{svm.TypeI32},
		func(args []svm.Value) ([]svm.Value, error) {
			//closure.value += args[0].ToI32()
			//	res := args[0].ToI32() * args[1].ToI32()
			return []svm.Value{svm.I32(args[0].ToI32())}, nil
		},
	)
	noError(err)

	imports, err := ib.Build()
	noError(err)
	defer imports.Free()

	kv, err := svm.NewMemKVStore()
	noError(err)
	defer kv.Free()

	host := unsafe.Pointer(&host)

	svmRuntime, err := svm.NewRuntimeBuilder().
		WithImports(imports).
		WithMemKVStore(kv).
		WithHost(host).
		Build()
	noError(err)
	defer svmRuntime.Free()
	fmt.Printf("1) Runtime: %v\n\n", svmRuntime)

	version := 0
	gasMetering := false
	gasLimit := uint64(0)

	// Generate Deploy Template tx.
	code, err := ioutil.ReadFile("./wasm/counter.wasm")
	noError(err)
	name := "name"
	dataLayout := svm.DataLayout{4}
	tx, err := codec.Get().EncodeTxDeployTemplate(version, name, code, dataLayout.Encode())
	noError(err)

	// Deploy Template.
	author := svm.Address{}
	deployTemplateReceipt, err := deployTemplate(
		svmRuntime,
		tx,
		author,
		gasMetering,
		gasLimit,
	)
	noError(err)
	fmt.Printf("2) %v\n", deployTemplateReceipt)

	// Generate Spawn App tx.
	calldata, err := codec.Get().EncodeCallData([]string{"u32"}, []int{10})
	noError(err)
	tx, err = codec.Get().EncodeTxSpawnApp(
		version,
		deployTemplateReceipt.TemplateAddr[:],
		name,
		"initialize",
		calldata,
	)
	noError(err)

	// Spawn App.
	creator := svm.Address{}
	spawnAppReceipt, err := spawnApp(
		svmRuntime,
		tx,
		creator,
		gasMetering,
		gasLimit,
	)
	noError(err)
	fmt.Printf("3) %v\n", spawnAppReceipt)

	// Generate Exec App tx.
	calldata, err = codec.Get().EncodeCallData(
		[]string{"u32", "u32"},
		[]int{3, 5},
	)
	noError(err)
	tx, err = codec.Get().EncodeTxExecApp(
		version,
		spawnAppReceipt.AppAddr[:],
		"add_and_mul",
		calldata,
	)
	noError(err)

	// Exec App.
	receiptExecApp, err := execApp(
		svmRuntime,
		tx,
		spawnAppReceipt.State,
		gasMetering,
		gasLimit,
	)
	fmt.Printf("4.0) %v\n", receiptExecApp)

	returndata, err := codec.Get().DecodeReturndata(receiptExecApp.Returndata)
	noError(err)

	fmt.Printf("4.0) %v\n", returndata)

	runtime.KeepAlive(ib)

	//// 4.1) Storage value get.
	//execAppResult, err = execApp(
	//	svmRuntime,
	//	version,
	//	spawnAppReceipt.AppAddr,
	//	"storage_get",
	//	[]byte(nil), //svm.Values(nil), // TODO: FIX
	//	execAppResult.NewState,
	//	gasMetering,
	//	gasLimit,
	//)
	//fmt.Printf("4.1) %v\n", execAppResult)
	//
	//// 4.2) Host import function value increment.
	//execAppResult, err = execApp(
	//	svmRuntime,
	//	version,
	//	spawnAppReceipt.AppAddr,
	//	"host_inc",
	//	[]byte(nil), //svm.Values{svm.I32(25)}, TODO: FIX
	//	execAppResult.NewState,
	//	gasMetering,
	//	gasLimit,
	//)
	//fmt.Printf("4.2) %v\n", execAppResult)
	//
	//// 4.3) Host import function value get.
	//execAppResult, err = execApp(
	//	svmRuntime,
	//	version,
	//	spawnAppReceipt.AppAddr,
	//	"host_get",
	//	[]byte(nil), // svm.Values(nil), TODO: FIX
	//	execAppResult.NewState,
	//	gasMetering,
	//	gasLimit,
	//)
	//fmt.Printf("4.3) %v\n", execAppResult)
}

func deployTemplate(
	runtime svm.Runtime,
	appTemplate []byte,
	author svm.Address,
	gasMetering bool,
	gasLimit uint64,
) (*svm.DeployTemplateReceipt, error) {

	// TODO: validation is temporarily disabled due to pending issues. Should be re-enabled.
	//if err = svm.ValidateTemplate(runtime, appTemplate); err != nil {
	//	return nil, err
	//}

	res, err := svm.DeployTemplate(
		runtime,
		appTemplate,
		author,
		gasMetering,
		gasLimit,
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func spawnApp(
	runtime svm.Runtime,
	tx []byte,
	creator svm.Address,
	gasMetering bool,
	gasLimit uint64,
) (*svm.SpawnAppReceipt, error) {
	if err := svm.ValidateApp(runtime, tx); err != nil {
		return nil, err
	}

	res, err := svm.SpawnApp(
		runtime,
		tx,
		creator,
		gasMetering,
		gasLimit,
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func execApp(
	runtime svm.Runtime,
	tx []byte,
	appState []byte,
	gasMetering bool,
	gasLimit uint64,
) (*svm.ExecAppReceipt, error) {
	if _, err := svm.ValidateAppTx(runtime, tx); err != nil {
		return nil, err
	}

	return svm.ExecApp(runtime, tx, appState, gasMetering, gasLimit)
}

func noError(err error) {
	if err != nil {
		panic(err)
	}
}
