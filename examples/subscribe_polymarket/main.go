package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzz/ethclient"
	"github.com/ivanzz/ethclient/examples/subscribe_polymarket/contracts/ctf_exchange"
)

func main() {
	// For more details about logs
	// handler := log.NewTerminalHandler(os.Stdout, true)
	// logger := log.NewLogger(handler)
	// log.SetDefault(logger)

	client, err := ethclient.Dial("wss://polygon-bor-rpc.publicnode.com")
	if err != nil {
		panic(err)
	}

	ctfExchange, err := ctf_exchange.NewCtfExchange(common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"), client)
	if err != nil {
		panic(err)
	}

	startBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}

	ctfExchangeOrderFilledEventChan := make(chan *ctf_exchange.CtfExchangeOrderFilled)
	subscription, err := ctfExchange.WatchOrderFilled(
		&bind.WatchOpts{Start: &startBlock}, ctfExchangeOrderFilledEventChan, nil, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("WatchOrderFilled: %v", err))
	}

	defer subscription.Unsubscribe()

	fmt.Println("Subscribe successful")

	go func() {
		for event := range ctfExchangeOrderFilledEventChan {
			fmt.Printf("Got event: %+v\n", event)
		}
	}()

	time.Sleep(2 * time.Minute)
}
