package main

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://base-rpc.publicnode.com")
	if err != nil {
		panic(err)
	}

	defer client.Close()

	block, err := client.BlockByHash(context.Background(), common.HexToHash("0x69fcdba289817a374e43e989a36ffb09d323c9aeb8f22716685de0b897c01f95"))
	if err != nil {
		fmt.Printf("failed to get block by hash: %v", err)
		return
	}

	fmt.Printf("Block#%v\n", block.Number().Int64())
}
