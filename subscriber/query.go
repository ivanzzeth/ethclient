package subscriber

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetQueryKey(query ethereum.FilterQuery) string {
	hash := GetQueryHash(query)

	return hash.Hex()
}

func GetQueryHash(query ethereum.FilterQuery) common.Hash {
	json, _ := json.Marshal(query)
	hash := crypto.Keccak256Hash(json)
	return hash
}
