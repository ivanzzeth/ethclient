package consts

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoAnyKeyStores       = errors.New("No any keystores")
	ErrMessagePrivateKeyNil = errors.New("PrivateKey is nil")
)

type RevertError struct {
	Id            string
	FuncSignature string
	Params        []interface{}
}

func (e RevertError) Error() string {
	paramsJs, _ := json.Marshal(e.Params)

	paramsStr := strings.ReplaceAll(string(paramsJs), "\"", "")
	paramsStr = strings.ReplaceAll(paramsStr, "[", "")
	paramsStr = strings.ReplaceAll(paramsStr, "]", "")
	return fmt.Sprintf("%s %s(%v)", e.Id, e.FuncSignature, paramsStr)
}
