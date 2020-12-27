package main

import (
	"fmt"
	"go-svm/svm"
	"io/ioutil"
	"os"
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
	// 1) Initialize runtime.
	ib := svm.NewImportsBuilder()

	ib, err := ib.AppendFunction(
		"inc",
		svm.ValueTypes{svm.TypeI32},
		svm.ValueTypes{},
		func(args []svm.Value) ([]svm.Value, error) {
			closure.value += args[0].ToI32()
			return []svm.Value{}, nil
		},
	)
	noError(err)

	ib, err = ib.AppendFunction(
		"get",
		svm.ValueTypes{},
		svm.ValueTypes{svm.TypeI32},
		func(_ []svm.Value) ([]svm.Value, error) {
			return []svm.Value{svm.I32(host.value)}, nil
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

	runtime, err := svm.NewRuntimeBuilder().
		WithImports(imports).
		WithMemKVStore(kv).
		WithHost(host).
		Build()
	noError(err)
	defer runtime.Free()
	fmt.Printf("1) Runtime: %v\n\n", runtime)

	version := 0
	gasMetering := false
	gasLimit := uint64(0)

	// 2) Deploy Template.
	// TODO: add on-the-fly wat2wasm translation
	code, err := ioutil.ReadFile("counter_template.wasm")
	dataLayout := svm.DataLayout{4}
	noError(err)
	name := "name"
	author := svm.Address{}
	deployTemplateReceipt, err := deployTemplate(
		runtime,
		code,
		dataLayout,
		version,
		name,
		author,
		gasMetering,
		gasLimit,
	)
	noError(err)
	fmt.Printf("2) %v\n", deployTemplateReceipt)

	os.Exit(0)

	// 3) Spawn App.
	creator := svm.Address{}
	spawnAppResult, err := spawnApp(
		runtime,
		version,
		deployTemplateReceipt.TemplateAddr,
		name,
		"storage_inc",
		[]byte(nil), // TODO: FIX. encode svm.Values{svm.I32(5)},
		creator,
		gasMetering,
		gasLimit,
	)
	noError(err)
	fmt.Printf("3) %v\n", spawnAppResult)

	// 4) Exec App
	// 4.0) Storage value increment.
	execAppResult, err := execApp(
		runtime,
		version,
		spawnAppResult.AppAddr,
		"storage_inc",
		[]byte(nil), // TODO: FIX. encode svm.Values{svm.I32(5)},
		spawnAppResult.InitialState,
		gasMetering,
		gasLimit,
	)
	fmt.Printf("4.0) %v\n", execAppResult)

	// 4.1) Storage value get.
	execAppResult, err = execApp(
		runtime,
		version,
		spawnAppResult.AppAddr,
		"storage_get",
		[]byte(nil), //svm.Values(nil), // TODO: FIX
		execAppResult.NewState,
		gasMetering,
		gasLimit,
	)
	fmt.Printf("4.1) %v\n", execAppResult)

	// 4.2) Host import function value increment.
	execAppResult, err = execApp(
		runtime,
		version,
		spawnAppResult.AppAddr,
		"host_inc",
		[]byte(nil), //svm.Values{svm.I32(25)}, TODO: FIX
		execAppResult.NewState,
		gasMetering,
		gasLimit,
	)
	fmt.Printf("4.2) %v\n", execAppResult)

	// 4.3) Host import function value get.
	execAppResult, err = execApp(
		runtime,
		version,
		spawnAppResult.AppAddr,
		"host_get",
		[]byte(nil), // svm.Values(nil), TODO: FIX
		execAppResult.NewState,
		gasMetering,
		gasLimit,
	)
	fmt.Printf("4.3) %v\n", execAppResult)
}

func deployTemplate(
	runtime svm.Runtime,
	code []byte,
	dataLayout svm.DataLayout,
	version int,
	name string,
	author svm.Address,
	gasMetering bool,
	gasLimit uint64,
) (*svm.DeployTemplateReceipt, error) {
	appTemplate, err := svm.EncodeAppTemplate(
		version,
		name,
		code,
		dataLayout,
	)
	if err != nil {
		return nil, err
	}

	if err = svm.ValidateTemplate(runtime, appTemplate); err != nil {
		return nil, err
	}

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
	version int,
	templateAddr svm.Address,
	name string,
	ctorName string,
	calldata []byte,
	creator svm.Address,
	gasMetering bool,
	gasLimit uint64,
) (*svm.SpawnAppReceipt, error) {
	spawnApp, err := svm.EncodeSpawnApp(
		version,
		templateAddr,
		name,
		ctorName,
		calldata,
	)
	if err != nil {
		return nil, err
	}

	if err = svm.ValidateApp(runtime, spawnApp); err != nil {
		return nil, err
	}

	res, err := svm.SpawnApp(
		runtime,
		spawnApp,
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
	version int,
	appAddr svm.Address,
	funcName string,
	calldata []byte,
	appState []byte,
	gasMetering bool,
	gasLimit uint64,
) (*svm.ExecAppResult, error) {
	appTx, err := svm.EncodeAppTx(
		version,
		appAddr,
		funcName,
		calldata,
	)
	if err != nil {
		return nil, err
	}

	if _, err = svm.ValidateAppTx(runtime, appTx); err != nil {
		return nil, err
	}

	execAppResult, err := svm.ExecApp(runtime, appTx, appState, gasMetering, gasLimit)
	if err != nil {
		return nil, err
	}

	return execAppResult, nil
}

func noError(err error) {
	if err != nil {
		panic(err)
	}
}
