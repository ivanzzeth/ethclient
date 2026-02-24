package subscriber_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/subscriber"
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

func testHeaderSubscriber(t *testing.T, sim *simulated.Backend) {
	sim.Commit()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	client := sim.Client()
	if cs, ok := client.Subscriber.(*subscriber.ChainSubscriber); ok {
		defer cs.Close()
	}

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
