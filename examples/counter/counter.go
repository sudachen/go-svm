package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"go-svm/codec"
	"go-svm/svm"
	"io/ioutil"
)

const (
	templateFilename = "./wasm/counter.wasm"
	initialValue     = 10
)

var counter int32 = initialValue

func main() {
	// Build imports.
	imports, err := svm.NewImportsBuilder().
		RegisterFunction(
			"add",
			svm.ValueTypes{svm.TypeI32, svm.TypeI32},
			svm.ValueTypes{svm.TypeI32},
			func(args []svm.Value) ([]svm.Value, error) {
				curr := args[0].ToI32()
				val := args[1].ToI32()

				fmt.Printf("`add` invoked by SVM; args: (%v, %v)\n", curr, val)

				//if curr != counter {
				//	panic(fmt.Sprintf("closure counter value mismatch: expected: %v, found: %v", curr, counter))
				//}

				res := curr + val
				counter = res
				return []svm.Value{svm.I32(res)}, nil
			},
		).RegisterFunction(
		"mul",
		svm.ValueTypes{svm.TypeI32, svm.TypeI32},
		svm.ValueTypes{svm.TypeI32},
		func(args []svm.Value) ([]svm.Value, error) {
			curr := args[0].ToI32()
			val := args[1].ToI32()

			fmt.Printf("`mul` invoked by SVM; args: (%v, %v)\n", curr, val)

			//if curr != counter {
			//	panic(fmt.Sprintf("closure counter value mismatch: expected: %v, found: %v", curr, counter))
			//}

			res := curr * val
			counter = res
			return []svm.Value{svm.I32(res)}, nil
		},
	).Build()
	noError(err)
	defer imports.Free()

	kv, err := svm.NewStateKV_FFI()
	noError(err)
	defer kv.Free()

	kv.RegisterGet(func(key []byte) []byte {
		fmt.Printf("FFI-state-KV `get` invoked by SVM; arg: (%x)\n", key)

		v := make([]byte, svm.KVValueSize)
		v[0] = 1
		return v
	})
	kv.RegisterSet(func(key []byte, val []byte) {
		fmt.Printf("FFI-state-KV `set` invoked by SVM; key: %x, val: %x\n", key, val)
	})
	kv.RegisterDiscard(func() {
		fmt.Printf("FFI-state-KV `discard` invoked by SVM\n")
	})
	kv.RegisterCheckpoint(func() []byte {
		fmt.Printf("FFI-state-KV `checkpoint` invoked by SVM\n")

		state := make([]byte, svm.StateSize)
		state[0] = 2
		return state
	})
	kv.RegisterHead(func() []byte {
		fmt.Printf("FFI-state-KV `head` invoked by SVM\n")

		state := make([]byte, svm.StateSize)
		state[0] = 3
		return state
	})

	// Initialize runtime.
	svmRuntime, err := svm.NewRuntimeBuilder().
		WithImports(imports).
		WithStateKV_FFI(&kv).
		Build()
	noError(err)
	spew.Dump(svmRuntime)
	println()
	defer svmRuntime.Free()

	version := 0
	gasMetering := false
	gasLimit := uint64(0)

	// Deploy Template: generate tx.
	code, err := ioutil.ReadFile(templateFilename)
	noError(err)
	name := "name"
	dataLayout := svm.DataLayout{4}
	tx, err := codec.EncodeTxDeployTemplate(version, name, code, dataLayout.Encode())
	noError(err)

	// Deploy Template: validate tx.
	// TODO: re-enable; temporarily disabled due to pending SVM issue.
	//err = svm.ValidateTemplate(svmRuntime, tx)
	//noError(err)

	// Deploy Template.
	author := svm.Address{}
	receiptDeployTemplate, err := svm.DeployTemplate(
		svmRuntime,
		tx,
		author,
		gasMetering,
		gasLimit,
	)
	noError(err)
	println("DeployTemplate finished successfully.")
	spew.Dump(receiptDeployTemplate)
	println()

	// Spawn App: generate tx.
	calldata, err := codec.EncodeCallData(
		[]string{"u32"},
		[]int{initialValue},
	)
	noError(err)
	tx, err = codec.EncodeTxSpawnApp(
		version,
		receiptDeployTemplate.TemplateAddr[:],
		name,
		"initialize",
		calldata,
	)
	noError(err)

	// Spawn App: validate tx.
	creator := svm.Address{}
	err = svm.ValidateApp(svmRuntime, tx)
	noError(err)

	// Spawn App.
	receiptSpawnApp, err := svm.SpawnApp(
		svmRuntime,
		tx,
		creator,
		gasMetering,
		gasLimit,
	)
	noError(err)
	println("SpawnApp finished successfully.")
	spew.Dump(receiptSpawnApp)
	returndata, err := codec.DecodeReturndata(receiptSpawnApp.Returndata)
	noError(err)
	fmt.Printf("Decoded Returndata: %v\n", returndata)
	println()

	// Exec App: generate tx.
	calldata, err = codec.EncodeCallData(
		[]string{"u32", "u32"},
		[]int{5, 5},
	)
	noError(err)
	tx, err = codec.EncodeTxExecApp(
		version,
		receiptSpawnApp.AppAddr[:],
		"add_and_mul",
		calldata,
	)
	noError(err)

	// Exec App: validate tx.
	_, err = svm.ValidateAppTx(svmRuntime, tx)
	noError(err)

	// Exec App.
	receiptExecApp, err := svm.ExecApp(svmRuntime, tx, receiptSpawnApp.State, gasMetering, gasLimit)
	noError(err)
	println("ExecApp finished successfully.")
	spew.Dump(receiptExecApp)
	returndata, err = codec.DecodeReturndata(receiptExecApp.Returndata)
	noError(err)
	fmt.Printf("Decoded Returndata: %v\n", returndata)
}

func noError(err error) {
	if err != nil {
		panic(err)
	}
}
