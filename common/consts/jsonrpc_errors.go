package consts

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/rpc"
)

type JsonRpcErrorCode int

const (
	// Standard errors
	JsonRpcErrorCodeParseError      JsonRpcErrorCode = -32700 // The JSON request is invalid, this can be due to syntax errors.
	JsonRpcErrorCodeInvalidRequest  JsonRpcErrorCode = -32600 // The JSON request is possibly malformed.
	JsonRpcErrorCodeMethodNotFound  JsonRpcErrorCode = -32601 // The method does not exist, often due to a typo in the method name or the method not being supported.
	JsonRpcErrorCodeInvalidArgument JsonRpcErrorCode = -32602 // Invalid method parameters. For example, "error":{"code":-32602,"message":"invalid argument 0: json: cannot unmarshal hex string without 0x prefix into Go value of type common.Hash"} indicates the 0x prefix is missing from the hexadecimal address.
	JsonRpcErrorCodeInternalError   JsonRpcErrorCode = -32603 // An internal JSON-RPC error, often caused by a bad or invalid payload.

	// Non-standard errors for infura
	JsonRpcErrorCodeInvalidInput               JsonRpcErrorCode = -32000 // Missing or invalid parameters, possibly due to server issues or a block not being processed yet.
	JsonRpcErrorCodeResourceNotFound           JsonRpcErrorCode = -32001 // The requested resource cannot be found, possibly when calling an unsupported method.
	JsonRpcErrorCodeResourceUnavailable        JsonRpcErrorCode = -32002 // The requested resource is not available.
	JsonRpcErrorCodeTransactionRejected        JsonRpcErrorCode = -32003 // The transaction could not be created.
	JsonRpcErrorCodeMethodNotSupported         JsonRpcErrorCode = -32004 // The requested method is not implemented.
	JsonRpcErrorCodeLimitExceeded              JsonRpcErrorCode = -32005 // The request exceeds your request limit. For more information, refer to Avoid rate limiting.
	JsonRpcErrorCodeJsonRpcVersionNotSupported JsonRpcErrorCode = -32006 // The version of the JSON-RPC protocol is not supported.
)

type JsonRpcError struct {
	Code    JsonRpcErrorCode `json:"code"`
	Message string           `json:"message"`
	Data    interface{}      `json:"data,omitempty"`

	// Provides additional information for json rpc error.
	// It's not standard for JSON RPC spec.
	DecodedData interface{} `json:"decoded_data,omitempty"`
	RawError    string      `json:"raw_error"`
	Abi         abi.ABI     `json:"-"`

	// RevertReason stores the raw revert reason string extracted from Error(string) format
	// This is useful when the revert reason contains binary data (e.g., address bytes)
	RevertReason string `json:"revert_reason,omitempty"`
	// IsRevertError indicates whether this error is a revert error (Error(string) format)
	IsRevertError bool `json:"is_revert_error,omitempty"`
}

func (e *JsonRpcError) Error() string {
	if e.DecodedData == nil {
		errData := e.Data
		if errData == nil {
			errData = e.RawError
		} else {
			if data, ok := e.Data.(string); ok {
				if strings.HasPrefix(data, "0x") {
					data = data[2:]
				}
				hexData, err := hex.DecodeString(data)
				if err == nil && len(hexData) >= 4 {
					errorDefinition, err := e.Abi.ErrorByID([4]byte(hexData))
					// error defined in ABI
					if err == nil {
						// name := errorDefinition.Name
						errSignature := errorDefinition.String()
						if strings.HasPrefix(errSignature, "error ") {
							errSignature = errSignature[6:]
						}
						params, _ := errorDefinition.Inputs.Unpack(hexData[4:])
						id := errorDefinition.ID.Hex()

						errData = RevertError{
							Id:            id,
							FuncSignature: errSignature,
							Params:        params,
						}
					} else {
						// try to decode using abi.UnpackRevert
						revertReason, err := abi.UnpackRevert(hexData)
						if err == nil {
							// Store the raw revert reason for use by callers
							// For calculateCreateProxyWithNonceAddress, revertReason is the address bytes (20 bytes)
							// as a string (not hex string, but raw bytes as string)
							e.RevertReason = revertReason
							e.IsRevertError = true
							errData = fmt.Sprintf(`reverted with [%s]`, revertReason)
						}
					}
				}
			}
		}

		e.DecodedData = errData
	}

	return fmt.Sprintf("json-rpc error(code=%d, msg=\"%s\", data=\"%v\", decoded_data=\"%v\", decoded_data_type=%T)",
		e.Code, e.Message, e.Data, e.DecodedData, e.DecodedData)
}

func (err *JsonRpcError) ErrorCode() JsonRpcErrorCode {
	return err.Code
}

func (err *JsonRpcError) ErrorData() interface{} {
	return err.Data
}

// GetRevertReason returns the raw revert reason string extracted from Error(string) format.
// This can be used to extract binary data (e.g., address bytes) from revert errors.
func (err *JsonRpcError) GetRevertReason() string {
	return err.RevertReason
}

// Decode json rpc errors and provide additional information using abi.
func DecodeJsonRpcError(err error, evmABI abi.ABI) error {
	if err == nil {
		return nil
	}
	jsonErr := &JsonRpcError{
		Abi:      evmABI,
		RawError: err.Error(),
		// default error for internal errors from client side
		Code:    JsonRpcErrorCodeInternalError,
		Message: err.Error(),
	}

	ec, ok := err.(rpc.Error)
	if ok {
		jsonErr.Code = JsonRpcErrorCode(ec.ErrorCode())
		jsonErr.Message = ec.Error()
	}

	de, ok := err.(rpc.DataError)
	if ok {
		jsonErr.Data = de.ErrorData()
	}

	// Force decode the error to set the RevertReason and IsRevertError fields
	_ = jsonErr.Error()

	return jsonErr
}
