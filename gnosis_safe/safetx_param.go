package gnosissafe

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var _ SafeTxParam = &SafeTxParamV1_3{}

type SafeTxParam interface {
	Version() string
}

type SafeTxParamV1_3 struct {
	To             common.Address
	Value          *big.Int
	Calldata       []byte
	Operation      uint8
	SafeTxGas      *big.Int
	BaseGas        *big.Int
	GasPrice       *big.Int
	GasToken       common.Address
	RefundReceiver common.Address
}

func (param *SafeTxParamV1_3) Version() string {
	return "1.3.0"
}
