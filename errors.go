package ethclient

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	ErrNoAnyKeyStores       = errors.New("No any keystores")
	ErrMessagePrivateKeyNil = errors.New("PrivateKey is nil")
)

type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	abi     abi.ABI
}

// TODO: translate error using abi
func (e *JsonRpcError) Error() string {
	errData := e.Data
	if data, ok := e.Data.(string); ok {
		fmt.Println("JSONRPCERROR111, ", data)
		if strings.HasPrefix(data, "0x") {
			data = data[2:]
		}
		hexData, err := hex.DecodeString(data)
		if err == nil && len(hexData) >= 4 {
			fmt.Println("JSONRPCERROR222, ", data)
			errorDefinition, err := e.abi.ErrorByID([4]byte(hexData))
			if err == nil {
				// name := errorDefinition.Name
				errSignature := errorDefinition.String()
				if strings.HasPrefix(errSignature, "error ") {
					errSignature = errSignature[6:]
				}
				params, _ := errorDefinition.Inputs.Unpack(hexData[4:])
				id := errorDefinition.ID.Hex()

				paramsJs, _ := json.Marshal(params)

				// keysStr := strings.ReplaceAll(string(keysJs), "\"", "")
				paramsStr := strings.ReplaceAll(string(paramsJs), "\"", "")

				paramsStr = strings.ReplaceAll(paramsStr, "[", "")
				paramsStr = strings.ReplaceAll(paramsStr, "]", "")

				errData = fmt.Sprintf("%s %s(%v)", id, errSignature, paramsStr)
			}
		}
	}

	return fmt.Sprintf("json-rpc error(code=%d, msg=\"%s\", data=\"%v\", data_type=%T)", e.Code, e.Message, errData, errData)
}

func (err *JsonRpcError) ErrorCode() int {
	return err.Code
}

func (err *JsonRpcError) ErrorData() interface{} {
	return err.Data
}
