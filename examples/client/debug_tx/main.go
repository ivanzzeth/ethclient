package main

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/ethclient"
)

func main() {
	// For more details about logs
	// handler := log.NewTerminalHandler(os.Stdout, true)
	// logger := log.NewLogger(handler)
	// log.SetDefault(logger)

	client, err := ethclient.Dial("wss://opbnb-rpc.publicnode.com")
	if err != nil {
		panic(err)
	}

	// client.SetABI()
	ret, err := client.DebugTransactionOnChain(context.Background(), common.HexToHash("0x8b1becff129aa708913bd3278c3581f34833227561c5bf3a54ece3334f5d4a47"))
	fmt.Printf("got ret 0x%x, err %v\n", ret, err)
}
