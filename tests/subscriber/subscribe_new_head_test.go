package subscriber_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	rawEthclient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	// "github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func Test_HeaderSubscriber(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testHeaderSubscriber(t, sim)
}

func Test_RealHeaderSubscriber(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// wss://opbnb-rpc.publicnode.com
	// ws://localhost:3005/ws/8453
	client, err := rawEthclient.Dial("wss://opbnb-rpc.publicnode.com")
	// client, err := rawEthclient.Dial("ws://localhost:3005/ws/8453")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("dial successful")

	ch := make(chan *types.Header)
	// sub, err := client.RawClient().SubscribeNewHead(ctx, ch)
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

func testHeaderSubscriber(t *testing.T, sim *simulated.Backend) {
	sim.Commit()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	client := sim.Client()

	ch := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(ctx, ch)
	if err != nil {
		t.Fatal(err)
	}

	expectedBlockNumbers := []uint64{2, 3, 4}
	gotBlockNumbers := []uint64{}

	go func() {
		for header := range ch {
			gotBlockNumbers = append(gotBlockNumbers, header.Number.Uint64())
		}
	}()

	sim.Commit()
	sim.Commit()
	sim.Commit()

	sub.Unsubscribe()

	time.Sleep(2 * time.Second)
	assert.Equal(t, expectedBlockNumbers, gotBlockNumbers, "wrong headers received")
}
