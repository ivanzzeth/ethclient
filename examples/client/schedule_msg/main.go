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
			From:      helper.Addr1,
			To:        &helper.Addr1,
			StartTime: time.Now().Add(5 * time.Second).UnixNano(),
		}).SetRandomId())

		client.ScheduleMsg((&message.Request{
			From: helper.Addr1,
			To:   &helper.Addr1,
			// StartTime:      time.Now().Add(5 * time.Second).UnixNano(),
			ExpirationTime: time.Now().UnixNano() - int64(5*time.Second),
		}).SetRandomId())

		client.ScheduleMsg((&message.Request{
			From:           helper.Addr1,
			To:             &helper.Addr1,
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
