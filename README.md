# ethclient

## Description
Extension ethclient.

## Prerequisites
golang

## Feautres
- [x] Schedule message
- [x] Sequence message
- [x] Protect message
- [x] Nonce management
- [x] Concurrent Transaction in Safe Multisig Wallets
- [ ] Multiple rpc url supported

## Quick Start
```go
package main

import (
	"fmt"
	"time"

	"github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/tests/helper"
)

func main() {
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	go func() {
		client.ScheduleMsg((&message.Request{
			From:      helper.Addr,
			To:        &helper.Addr,
			StartTime: time.Now().Add(5 * time.Second).UnixNano(),
		}).SetRandomId())

		client.ScheduleMsg((&message.Request{
			From: helper.Addr,
			To:   &helper.Addr,
			// StartTime:      time.Now().Add(5 * time.Second).UnixNano(),
			ExpirationTime: time.Now().UnixNano() - int64(5*time.Second),
		}).SetRandomId())

		client.ScheduleMsg((&message.Request{
			From:           helper.Addr,
			To:             &helper.Addr,
			ExpirationTime: time.Now().Add(10 * time.Second).UnixNano(),
			Interval:       2 * time.Second,
		}).SetRandomId())

		time.Sleep(20 * time.Second)
		client.CloseSendMsg()
	}()

	for resp := range client.Response() {
		fmt.Println("execution resp: ", resp)
	}
}

```

## Concurrent Transaction Management in Safe Multisig Wallets 
The Safe multisig contract also uses a nonce.
Our solution manages this nonce off-chain.
Now you can submit multiple transactions at once - no need to wait for confirmations.

The current implementation uses Safe v1.3, but you can build your own for different Safe contract versions.
```go
func BatchSendSafeTx(from common.Address, builder gnosissafe.SafeTxBuilder, deliverer gnosissafe.SafeTxDeliverer, params []gnosissafe.SafeTxParam) {

	safeContractAddress, _ := builder.GetContractAddress()

	for _, param := range params {
		callData, _, safeNonce, err := builder.Build(param)
		if err != nil {
			// Handle error
			return
		}

		req := message.Request{
			From:     from,
			To:       &safeContractAddress,
			Gas:      500000, // ensure gas meets the minimum requirement for successful Safe contract execution (avoiding reverts).
			Data:     callData,
		}
		err = deliverer.Deliver(&req, safeNonce)
		if err != nil {
			// Handle error
			return
		}
	}
}

```

## Running tests

- **Unit + integration** (simulated backend, no real RPC): `go test ./...` (subscriber tests need `-timeout=120s` or they may hit default timeout)
- **E2e** (includes tests that talk to real RPC): `go test -tags=e2e ./...`

E2e tests live in `*_e2e_test.go` files behind the `e2e` build tag.

## Setup local node for testing

you should install foundry before running the script below:

```bash
./cmd/run_test_node.sh
```

## License
The ethclient library is licensed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included in our repository in the COPYING.LESSER file.