package ethclient

import (
	"errors"
	"fmt"
)

var (
	ErrNoAnyKeyStores       = errors.New("No any keystores")
	ErrMessagePrivateKeyNil = errors.New("PrivateKey is nil")
)

type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err *JsonRpcError) Error() string {
	return fmt.Sprintf("json-rpc error(code=%d, msg=\"%s\", data=%v)", err.Code, err.Message, err.Data)
}

func (err *JsonRpcError) ErrorCode() int {
	return err.Code
}

func (err *JsonRpcError) ErrorData() interface{} {
	return err.Data
}
