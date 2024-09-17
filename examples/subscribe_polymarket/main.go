package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethclient/examples/subscribe_polymarket/contracts/ctf_exchange"
	"github.com/ivanzzeth/ethclient/subscriber"
	goredislib "github.com/redis/go-redis/v9"
)

func main() {
	// For more details about logs
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	client, err := ethclient.Dial("wss://polygon-bor-rpc.publicnode.com")
	if err != nil {
		panic(err)
	}

	defer client.Close()

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	redisClient := goredislib.NewClient(&goredislib.Options{
		Addr:     "localhost:16379",
		Password: "135683271d06e8",
	})

	pool := goredis.NewPool(redisClient)

	subStorage := subscriber.NewRedisStorage(chainId, pool)
	subscriber, err := subscriber.NewChainSubscriber(client.Client, subStorage)
	if err != nil {
		panic(err)
	}

	client.SetSubscriber(subscriber)

	ctfExchange, err := ctf_exchange.NewCtfExchange(common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"), client)
	if err != nil {
		panic(err)
	}

	currBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("Curr block:", currBlock)

	startBlock := uint64(61801957)
	ctfExchangeOrderFilledEventChan := make(chan *ctf_exchange.CtfExchangeOrderFilled)
	subscription, err := ctfExchange.WatchOrderFilled(
		&bind.WatchOpts{Start: &startBlock}, ctfExchangeOrderFilledEventChan, nil, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("WatchOrderFilled: %v", err))
	}

	defer subscription.Unsubscribe()

	fmt.Println("Subscribe successful")

	go CtfExchangeOrderFilledHandler(ctfExchangeOrderFilledEventChan)

	signalChan := make(chan os.Signal, 10)
	// SIGHUP: terminal closed
	// SIGINT: Ctrl+C
	// SIGTERM: program exit
	// SIGQUIT: Ctrl+/
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	waitElegantExit(signalChan)
}

func CtfExchangeOrderFilledHandler(ch <-chan *ctf_exchange.CtfExchangeOrderFilled) {
	for event := range ch {
		eventBytes, _ := json.MarshalIndent(event, "", " ")
		fmt.Printf("OrderFilled: %+v\n", string(eventBytes))
	}
}

func waitElegantExit(signalChan chan os.Signal) {
	for i := range signalChan {
		switch i {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			fmt.Println("receive exit signal ", i.String(), ",exit...")
			os.Exit(0)
		}
	}
}
