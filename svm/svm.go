package svm

import "go-svm/common"

type (
	DeployTemplateReceipt = common.ReceiptDeployTemplate
	SpawnAppReceipt       = common.ReceiptSpawnApp
	ExecAppReceipt        = common.ReceiptExecApp

	Address = common.Address
)

var (
	BytesToAddress = common.BytesToAddress
)
