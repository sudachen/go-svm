package svm

import (
	"fmt"
	"go-svm/codec"
)

type ExecAppResult struct {
	Receipt  []byte
	NewState []byte
	Returns  Values
	GasUsed  uint64
}

func (r ExecAppResult) String() string {
	return fmt.Sprintf(
		"ExecApp result:\n"+
			"  Receipt: %x\n"+
			"  NewState: %x\n"+
			"  Returns: %v\n"+
			"  GasUsed: %v\n",
		r.Receipt, r.NewState, r.Returns, r.GasUsed)
}

func DeployTemplate(runtime Runtime, appTemplate []byte, author Address, gasMetering bool, gasLimit uint64) (*DeployTemplateReceipt, error) {
	rawReceipt, err := cSvmDeployTemplate(runtime, appTemplate, author, gasMetering, gasLimit)
	if err != nil {
		return nil, err
	}

	c, err := codec.NewCodec()
	if err != nil {
		return nil, err
	}

	return c.DecodeDeployTemplateReceipt(rawReceipt)
}

func SpawnApp(runtime Runtime, spawnAppData []byte, creator Address,
	gasMetering bool, gasLimit uint64) (*SpawnAppReceipt, error) {
	rawReceipt, err := cSvmSpawnApp(runtime, spawnAppData, creator, gasMetering, gasLimit)
	if err != nil {
		return nil, err
	}

	c, err := codec.NewCodec()
	if err != nil {
		return nil, err
	}

	return c.DecodeSpawnAppReceipt(rawReceipt)
}

func ExecApp(runtime Runtime, appTx, appState []byte, gasMetering bool, gasLimit uint64) (*ExecAppResult, error) {
	receipt, err := cSvmExecApp(runtime, appTx, appState, gasMetering, gasLimit)
	if err != nil {
		return nil, err
	}
	//
	//newState, err := cSvmExecReceiptState(receipt)
	//if err != nil {
	//	return nil, err
	//}
	//
	//returns, err := cSvmExecReceiptReturns(receipt)
	//if err != nil {
	//	return nil, err
	//}
	//
	//gasUsed, err := cSvmExecReceiptGas(receipt)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return &ExecAppResult{
	//	Receipt:  receipt,
	//	NewState: newState,
	//	Returns:  returns,
	//	GasUsed:  gasUsed,
	//}, nil

	return &ExecAppResult{
		Receipt:  receipt,
		NewState: nil,
		Returns:  nil,
		GasUsed:  0,
	}, nil
}
