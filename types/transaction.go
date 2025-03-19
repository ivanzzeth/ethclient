package types

import "github.com/ethereum/go-ethereum/common/hexutil"

type MethodSelector [4]byte

const MethodSelectorLength = 4

func NewMethodSelector(selector string) MethodSelector {
	return MethodSelector(hexutil.MustDecode(selector))
}

func (s MethodSelector) Hex() string {
	return hexutil.Encode(s[:])
}
