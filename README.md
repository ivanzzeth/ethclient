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

## Setup local node for testing

you should install foundry before running the script below:

```bash
./cmd/run_test_node.sh
```

## License
The ethclient library is licensed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included in our repository in the COPYING.LESSER file.