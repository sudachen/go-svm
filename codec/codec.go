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

const (
	OkMarker  = 1
	ErrMarker = 0
)

type (
	DeployTemplateReceipt = common.DeployTemplateReceipt
	SpawnAppReceipt       = common.SpawnAppReceipt
)

func codecWasmFilePath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(file, "../../svm/svm_codec.wasm")
}

type Codec struct {
	instance *wasmer.Instance
}

func NewCodec() (*Codec, error) {
	bytes, err := ioutil.ReadFile(codecWasmFilePath())
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

	return &Codec{instance}, nil
}

func (c *Codec) EncodeCallData(abi []string, data []int) ([]byte, error) {
	calldataJson, err := json.Marshal(struct {
		ABI  []string `json:"abi,omitempty"`
		Data []int    `json:"data,omitempty"`
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

	return bytes, nil
}

func (c *Codec) DecodeDeployTemplateReceipt(rawReceipt []byte) (*DeployTemplateReceipt, error) {
	v, err := c.decodeReceipt(rawReceipt)
	if err != nil {
		return nil, err
	}

	return v.(*DeployTemplateReceipt), nil
}

func (c *Codec) DecodeSpawnAppReceipt(rawReceipt []byte) (*SpawnAppReceipt, error) {
	v, err := c.decodeReceipt(rawReceipt)
	if err != nil {
		return nil, err
	}

	return v.(*SpawnAppReceipt), nil
}

func (c *Codec) decodeReceipt(rawReceipt []byte) (interface{}, error) {
	decodeReceiptJson, err := json.Marshal(struct {
		Data string `json:"data,omitempty"`
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

	//fmt.Printf("## receipt: %s\n", ret)

	var v map[string]interface{}
	if err = json.Unmarshal(ret, &v); err != nil {
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

			decodedAddr, _ := hex.DecodeString(addr)

			return &common.DeployTemplateReceipt{
				Success:      success,
				TemplateAddr: common.BytesToAddress(decodedAddr),
				GasUsed:      int(gasUsed),
			}, nil

		default:
			panic(fmt.Sprintf("invalid receipt type: %v", ty))
		}
	}
}

func (c *Codec) newBuffer(data []byte) (int32, error) {
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

func (c *Codec) loadBuffer(ptr int32) ([]byte, error) {
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

func (c *Codec) bufferAlloc(size int32) (int32, error) {
	fn, err := c.instance.Exports.GetFunction("wasm_alloc")
	if err != nil {
		return 0, err
	}

	buf, err := fn(size)
	return buf.(int32), err
}

func (c *Codec) bufferLength(buf int32) (int32, error) {
	fn, err := c.instance.Exports.GetFunction("wasm_buffer_length")
	if err != nil {
		return 0, err
	}

	bufLen, err := fn(buf)
	return bufLen.(int32), err
}

func (c *Codec) bufferDataPtr(buf int32) (int32, error) {
	fn, err := c.instance.Exports.GetFunction("wasm_buffer_data")
	if err != nil {
		return 0, err
	}

	dataPtr, err := fn(buf)
	return dataPtr.(int32), err
}
