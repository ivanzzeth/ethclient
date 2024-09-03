# ethclient

## Description
Extension ethclient.

## Prerequisites
golang

## Quick Start
```go
package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ivanzz/ethclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	// The private key of your wallet address.
	privateKey, _ := crypto.HexToECDSA("9a01f5c57e377e0239e6036b7b2d700454b760b2dab51390f1eeb2f64fe98b68")

	// Dial Client.
	chainUrl := "ws://localhost:8546"
	client, err := ethclient.Dial(chainUrl)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// The address your want to send to.
	to := common.HexToAddress("0x06514D014e997bcd4A9381bF0C4Dc21bD32718D4")

	// Send single transaction.
	tx, err := client.SendMsg(ctx, ethclient.Message{
		To:         &to,
		PrivateKey: privateKey,
		Value:      big.NewInt(0),
	})

	if err != nil {
		fmt.Printf("Send single message err: %v\n", err)
		return
	}

	// Waiting n confirmations.
	contains, err := client.ConfirmTx(tx.Hash(), 2, 5*time.Second)
	if err != nil {
		panic(err)
	}

	if !contains {
		fmt.Printf("The transaction %v is not contained at blockchain", tx.Hash().Hex())
	} else {
		receipt, err := client.RawClient().TransactionReceipt(ctx, tx.Hash())
		// do something.
		_, _ = receipt, err
	}

	fmt.Println("Send single message successful, txHash:", tx.Hash().Hex())

	// Send multiple transactions.
	mesgs := make(chan ethclient.Message)
	txs, errs := client.ScheduleMsg(ctx, mesgs)
	go func() {
		for i := 0; i < 5; i++ {
			mesgs <- ethclient.Message{
				PrivateKey: privateKey,
				To:         &to,
			}
		}

		close(mesgs)
	}()

	for tx := range txs {
        fmt.Printf("Send multiple message successful, txHash: %v, nonce: %v, err: %v\n",
            tx.Hash().Hex(), tx.Nonce(), <-errs)
	}
}
```

## Setup local node for testing

you should install foundry before running the script below:

```bash
./run_local_node.sh
```

## License
The ethclient library is licensed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included in our repository in the COPYING.LESSER file.