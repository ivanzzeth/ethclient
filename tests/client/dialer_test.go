package client_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ivanzzeth/ethclient"
)

func test_Dialer_DialOnce(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	client := ethclient.DialOnce("http://localhost:8545")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	block, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(block)
}
