package common

import "fmt"

type ReceiptDeployTemplate struct {
	Success      bool
	Version      int
	TemplateAddr Address
	GasUsed      uint64
}

//func (r ReceiptDeployTemplate) String() string {
//	return fmt.Sprintf(
//			"Success: %v\n"+
//			"Version: %v\n"+
//			"Template Address: %v\n"+
//			"Gas Used: %v",
//		r.Success, r.Version, r.TemplateAddr, r.GasUsed)
//}

type ReceiptSpawnApp struct {
	Success    bool
	Version    int
	AppAddr    Address
	State      []byte
	Returndata []byte
	Logs       []string
	GasUsed    uint64
}

//func (r ReceiptSpawnApp) String() string {
//	return fmt.Sprintf(
//			"App Address: %x\n"+
//			"State: %x\n"+
//			"Returndata: %x\n"+
//			"Gas Used: %v\n" +
//			"Logs: %#v\n",
//		r.State, r.AppAddr, r.Returndata, r.GasUsed, r.Logs)
//}

type ReceiptExecApp struct {
	Success    bool
	Version    int
	NewState   []byte
	Returndata []byte
	Logs       []string
	GasUsed    uint64
}

func (r ReceiptExecApp) String() string {
	return fmt.Sprintf(
		"New State: %x\n"+
			"Returndata: %x\n"+
			"Logs: %v\n"+
			"GasUsed: %v\n",
		r.NewState, r.Returndata, r.Logs, r.GasUsed)
}
