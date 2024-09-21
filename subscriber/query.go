package subscriber

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Query struct {
	ChainID *big.Int
	ethereum.FilterQuery
}

func NewQuery(chainId *big.Int, q ethereum.FilterQuery) Query {
	return Query{
		ChainID:     big.NewInt(0).Set(chainId),
		FilterQuery: q,
	}
}

func (q Query) Hash() common.Hash {
	return GetQueryHash(q.ChainID, q.FilterQuery)
}

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
