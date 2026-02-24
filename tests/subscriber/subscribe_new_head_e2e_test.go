//go:build e2e

package subscriber_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	rawEthclient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

// Test_RealHeaderSubscriber runs against a real RPC (opbnb). Run with: go test -tags=e2e ./tests/subscriber/... -run Test_RealHeaderSubscriber
func Test_RealHeaderSubscriber(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// wss://opbnb-rpc.publicnode.com
	// ws://localhost:3005/ws/8453
	client, err := rawEthclient.Dial("wss://opbnb-rpc.publicnode.com")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("dial successful")

	ch := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(ctx, ch)
	if err != nil {
		t.Fatal(err)
	}

	defer sub.Unsubscribe()
	go func() {
		for header := range ch {
			t.Logf("===> header: %v", header)
		}
	}()

	time.Sleep(30 * time.Second)
}
