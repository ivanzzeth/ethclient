package ctf_exchange

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (e CtfExchangeOrderFilled) MarshalJSON() ([]byte, error) {
	type js struct {
		OrderHash         string
		Maker             common.Address
		Taker             common.Address
		MakerAssetId      string
		TakerAssetId      string
		MakerAmountFilled string
		TakerAmountFilled string
		Fee               string
		BlockNumber       uint64
		Raw               types.Log // Blockchain specific contextual infos
	}

	obj := js{
		OrderHash:         common.Hash(e.OrderHash).Hex(),
		Maker:             e.Maker,
		Taker:             e.Taker,
		MakerAssetId:      e.MakerAssetId.String(),
		TakerAssetId:      e.TakerAssetId.String(),
		MakerAmountFilled: e.MakerAmountFilled.String(),
		TakerAmountFilled: e.TakerAmountFilled.String(),
		Fee:               e.Fee.String(),
		BlockNumber:       e.Raw.BlockNumber,
		Raw:               e.Raw,
	}

	return json.Marshal(obj)
}
