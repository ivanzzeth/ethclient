package subscriber

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetQueryKey(chainId *big.Int, query ethereum.FilterQuery) string {
	hash := GetQueryHash(chainId, query)

	return hash.Hex()
}

func GetQueryHash(chainId *big.Int, query ethereum.FilterQuery) common.Hash {
	type js struct {
		ChainId string
		ethereum.FilterQuery
	}

	var obj js = js{
		ChainId:     chainId.String(),
		FilterQuery: query,
	}
	json, _ := json.Marshal(obj)
	hash := crypto.Keccak256Hash(json)
	return hash
}
