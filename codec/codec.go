package codec

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wasmerio/wasmer-go/wasmer"
	"go-svm/common"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

type (
	ReceiptDeployTemplate = common.ReceiptDeployTemplate
	ReceiptSpawnApp       = common.ReceiptSpawnApp
	ReceiptExecApp        = common.ReceiptExecApp
)

type SvmCodec struct {
	instance *wasmer.Instance
}

const (
	OkMarker  = 1
	ErrMarker = 0
)

var (
	codec *SvmCodec
)

func init() {
	var err error
	codec, err = newCodec(codecWasmFilePath())
	if err != nil {
		panic(err)
	}
}

func Get() *SvmCodec {
	return codec
}

func newCodec(filename string) (*SvmCodec, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	module, err := wasmer.NewModule(store, bytes)
	if err != nil {
		return nil, err
	}

	instance, err := wasmer.NewInstance(module, wasmer.NewImportObject())
	if err != nil {
		return nil, err
	}

	return &SvmCodec{instance}, nil
}

func (c *SvmCodec) EncodeTxDeployTemplate(version int, name string, code []byte, data []byte) ([]byte, error) {
	txJson, err := json.Marshal(struct {
		Version int    `json:"version"`
		Name    string `json:"name"`
		Code    string `json:"code"`
		Data    string `json:"data"`
	}{
		Version: version,
		Name:    name,
		Code:    hex.EncodeToString(code),
		Data:    hex.EncodeToString(data),
	})
	if err != nil {
		return nil, err
	}

	argPtr, err := c.newBuffer(txJson)
	if err != nil {
		return nil, err
	}

	fn, err := c.instance.Exports.GetFunction("wasm_deploy_template")
	if err != nil {
		return nil, err
	}

	retPtr, err := fn(argPtr)
	if err != nil {
		return nil, err
	}

	return c.loadBuffer(retPtr.(int32))
}

func (c *SvmCodec) EncodeTxSpawnApp(version int, templateAddr []byte, name string, ctorName string, calldata []byte) ([]byte, error) {
	txJson, err := json.Marshal(struct {
		Version      int    `json:"version"`
		TemplateAddr string `json:"template"`
		Name         string `json:"name"`
		CtorName     string `json:"ctor_name"`
		Calldata     string `json:"calldata"`
	}{
		Version:      version,
		TemplateAddr: hex.EncodeToString(templateAddr),
		Name:         name,
		CtorName:     ctorName,
		Calldata:     hex.EncodeToString(calldata),
	})
	if err != nil {
		return nil, err
	}

	argPtr, err := c.newBuffer(txJson)
	if err != nil {
		return nil, err
	}

	fn, err := c.instance.Exports.GetFunction("wasm_encode_spawn_app")
	if err != nil {
		return nil, err
	}

	retPtr, err := fn(argPtr)
	if err != nil {
		return nil, err
	}

	return c.loadBuffer(retPtr.(int32))
}

func (c *SvmCodec) EncodeTxExecApp(version int, appAddr []byte, funcName string, calldata []byte) ([]byte, error) {
	txJson, err := json.Marshal(struct {
		Version  int    `json:"version"`
		AppAddr  string `json:"app"`
		FuncName string `json:"func_name"`
		Calldata string `json:"calldata"`
	}{
		Version:  version,
		AppAddr:  hex.EncodeToString(appAddr),
		FuncName: funcName,
		Calldata: hex.EncodeToString(calldata),
	})
	if err != nil {
		return nil, err
	}

	argPtr, err := c.newBuffer(txJson)
	if err != nil {
		return nil, err
	}

	fn, err := c.instance.Exports.GetFunction("wasm_encode_exec_app")
	if err != nil {
		return nil, err
	}

	retPtr, err := fn(argPtr)
	if err != nil {
		return nil, err
	}

	return c.loadBuffer(retPtr.(int32))
}

func (c *SvmCodec) EncodeCallData(abi []string, data []int) ([]byte, error) {
	calldataJson, err := json.Marshal(struct {
		ABI  []string `json:"abi"`
		Data []int    `json:"data"`
	}{
		ABI:  abi,
		Data: data,
	})
	if err != nil {
		return nil, err
	}

	argPtr, err := c.newBuffer(calldataJson)
	if err != nil {
		return nil, err
	}

	fn, err := c.instance.Exports.GetFunction("wasm_encode_calldata")
	if err != nil {
		return nil, err
	}

	retPtr, err := fn(argPtr)
	if err != nil {
		return nil, err
	}

	ret, err := c.loadBuffer(retPtr.(int32))
	if err != nil {
		return nil, err
	}

	var v map[string]interface{}
	if err := json.Unmarshal(ret, &v); err != nil {
		return nil, err
	}

	calldata, ok := v["calldata"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid json format: %s", ret)
	}

	bytes, err := hex.DecodeString(calldata)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %s", calldata)
	}

	return bytes[1:], nil
}

func (c *SvmCodec) DecodeReturndata(rawReturndata []byte) (string, error) {
	//	a := append([]byte{1}, rawReturndata...)
	calldataJson, err := json.Marshal(struct {
		Calldata string `json:"calldata"`
	}{
		Calldata: hex.EncodeToString(rawReturndata),
	})

	if err != nil {
		return "", err
	}

	argPtr, err := c.newBuffer(calldataJson)
	if err != nil {
		return "", err
	}

	fn, err := c.instance.Exports.GetFunction("wasm_decode_calldata")
	if err != nil {
		return "", err
	}

	retPtr, err := fn(argPtr)
	if err != nil {
		return "", err
	}

	ret, err := c.loadBuffer(retPtr.(int32))
	if err != nil {
		return "", err
	}

	var v map[string]interface{}
	if err := json.Unmarshal(ret, &v); err != nil {
		return "", err
	}

	return string(ret), nil
}

func (c *SvmCodec) DecodeReceiptDeployTemplate(rawReceipt []byte) (*ReceiptDeployTemplate, error) {
	v, err := c.decodeReceipt(rawReceipt)
	if err != nil {
		return nil, err
	}

	return v.(*ReceiptDeployTemplate), nil
}

func (c *SvmCodec) DecodeReceiptSpawnApp(rawReceipt []byte) (*ReceiptSpawnApp, error) {
	v, err := c.decodeReceipt(rawReceipt)
	if err != nil {
		return nil, err
	}

	return v.(*ReceiptSpawnApp), nil
}

func (c *SvmCodec) DecodeReceiptExecApp(rawReceipt []byte) (*ReceiptExecApp, error) {
	v, err := c.decodeReceipt(rawReceipt)
	if err != nil {
		return nil, err
	}

	return v.(*ReceiptExecApp), nil
}

func (c *SvmCodec) decodeReceipt(rawReceipt []byte) (interface{}, error) {
	decodeReceiptJson, err := json.Marshal(struct {
		Data string `json:"data"`
	}{
		Data: hex.EncodeToString(rawReceipt),
	})
	if err != nil {
		return nil, err
	}

	bufPtr, err := c.newBuffer(decodeReceiptJson)
	if err != nil {
		return nil, err
	}

	fn, err := c.instance.Exports.GetFunction("wasm_decode_receipt")
	if err != nil {
		return nil, err
	}

	retBufPtr, err := fn(bufPtr)
	if err != nil {
		return nil, err
	}

	ret, err := c.loadBuffer(retBufPtr.(int32))
	if err != nil {
		return nil, err
	}

	fmt.Printf("## receipt json: %s\n", ret)

	return c.decodeReceiptJSON(ret)

}

func (c *SvmCodec) decodeReceiptJSON(jsonReceipt []byte) (interface{}, error) {
	var v map[string]interface{}
	if err := json.Unmarshal(jsonReceipt, &v); err != nil {
		return nil, err
	}

	errType, ok := v["err_type"].(string)
	if ok {
		switch errType {
		case "oog":
			return nil, errors.New("oog")
		case "template-not-found":
			templateAddr := v["template_addr"].(string)
			return nil, fmt.Errorf("template not found; template address: %v", templateAddr)
		case "app-not-found":
			appAddr := v["app_addr"].(string)
			return nil, fmt.Errorf("template not found; app address: %v", appAddr)
		case "compilation-failed":
			templateAddr := v["template_addr"].(string)
			appAddr := v["app_addr"].(string)
			msg := v["message"].(string)
			return nil, fmt.Errorf("compilation failed; template address: %v, app address: %v, msg: %v",
				templateAddr, appAddr, msg)
		case "instantiation-failed":
			templateAddr := v["template_addr"].(string)
			appAddr := v["app_addr"].(string)
			msg := v["message"].(string)
			return nil, fmt.Errorf("instantiation failed; template address: %v, app address: %v, msg: %v",
				templateAddr, appAddr, msg)
		case "function-not-found":
			templateAddr := v["template_addr"].(string)
			appAddr := v["app_addr"].(string)
			fnc := v["func"].(string)
			return nil, fmt.Errorf("function not found; template address: %v, app address: %v, func: %v",
				templateAddr, appAddr, fnc)
		case "function-failed":
			templateAddr := v["template_addr"].(string)
			appAddr := v["app_addr"].(string)
			fnc := v["func"].(string)
			msg := v["message"].(string)
			return nil, fmt.Errorf("function failed; template address: %v, app address: %v, func: %v, msg: %v",
				templateAddr, appAddr, fnc, msg)
		default:
			panic(fmt.Sprintf("invalid error type: %v", errType))
		}
	} else {
		ty := v["type"].(string)
		switch ty {
		case "deploy-template":
			success := v["success"].(bool)
			gasUsed := v["gas_used"].(float64)
			addr := v["addr"].(string)

			return &common.ReceiptDeployTemplate{
				Success:      success,
				TemplateAddr: common.BytesToAddress(mustDecodeHexString(addr)),
				GasUsed:      uint64(gasUsed),
			}, nil

		case "spawn-app":
			success := v["success"].(bool)
			app := v["app"].(string)
			state := v["state"].(string)
			returndata := v["returndata"].(string)
			logs := v["logs"].([]interface{})
			gasUsed := v["gas_used"].(float64)

			strLogs := make([]string, len(logs))
			for i, log := range logs {
				strLogs[i] = log.(string)
			}

			return &common.ReceiptSpawnApp{
				Success:    success,
				AppAddr:    common.BytesToAddress(mustDecodeHexString(app)),
				State:      mustDecodeHexString(state),
				Returndata: mustDecodeHexString(returndata),
				Logs:       strLogs,
				GasUsed:    uint64(gasUsed),
			}, nil

		case "exec-app":
			success := v["success"].(bool)
			newState := v["new_state"].(string)
			returndata := v["returndata"].(string)
			logs := v["logs"].([]interface{})
			gasUsed := v["gas_used"].(float64)

			strLogs := make([]string, len(logs))
			for i, log := range logs {
				strLogs[i] = log.(string)
			}

			return &common.ReceiptExecApp{
				Success:    success,
				NewState:   mustDecodeHexString(newState),
				Returndata: mustDecodeHexString(returndata),
				Logs:       strLogs,
				GasUsed:    uint64(gasUsed),
			}, nil

		default:
			panic(fmt.Sprintf("invalid receipt type: %v", ty))
		}
	}
}

func (c *SvmCodec) newBuffer(data []byte) (int32, error) {
	length := int32(len(data))
	ptr, err := c.bufferAlloc(length)
	if err != nil {
		return 0, err
	}

	bufferLength, err := c.bufferLength(ptr)
	if err != nil {
		return 0, err
	}
	if length != bufferLength {
		panic(fmt.Sprintf("allocated buffer size isn't sufficient; allocated: %v, got: %v", length, bufferLength))
	}

	dataPtr, err := c.bufferDataPtr(ptr)
	if err != nil {
		return 0, err
	}

	mem, err := c.instance.Exports.GetMemory("memory")
	if err != nil {
		return 0, err
	}

	copy(mem.Data()[dataPtr:], data)

	return ptr, nil
}

func (c *SvmCodec) loadBuffer(ptr int32) ([]byte, error) {
	length, err := c.bufferLength(ptr)
	if err != nil {
		return nil, err
	}

	dataPtr, err := c.bufferDataPtr(ptr)
	if err != nil {
		return nil, err
	}

	mem, err := c.instance.Exports.GetMemory("memory")
	if err != nil {
		return nil, err
	}

	buf := mem.Data()[dataPtr : dataPtr+length]
	marker := buf[0]
	data := buf[1:]

	switch marker {
	case ErrMarker:
		return nil, errors.New(string(data))
	case OkMarker:
		return data, nil
	default:
		panic("invalid marker")
	}
}

func (c *SvmCodec) bufferAlloc(size int32) (int32, error) {
	fn, err := c.instance.Exports.GetFunction("wasm_alloc")
	if err != nil {
		return 0, err
	}

	buf, err := fn(size)
	return buf.(int32), err
}

func (c *SvmCodec) bufferLength(buf int32) (int32, error) {
	fn, err := c.instance.Exports.GetFunction("wasm_buffer_length")
	if err != nil {
		return 0, err
	}

	bufLen, err := fn(buf)
	return bufLen.(int32), err
}

func (c *SvmCodec) bufferDataPtr(buf int32) (int32, error) {
	fn, err := c.instance.Exports.GetFunction("wasm_buffer_data")
	if err != nil {
		return 0, err
	}

	dataPtr, err := fn(buf)
	return dataPtr.(int32), err
}

func codecWasmFilePath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(file, "../../svm/svm_codec.wasm")
}

func mustDecodeHexString(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	return b
}
