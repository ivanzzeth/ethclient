package client_test

import (
	"testing"

	"github.com/ivanzzeth/ethclient/tests/helper"
)

func TestBatchSendMsg(t *testing.T) {
	client := helper.SetUpClient(t)

	testScheduleMsg(t, client)
}

func Test_BatchSendMsg_RandomlyReverted(t *testing.T) {
	client := helper.SetUpClient(t)

	test_ScheduleMsg_RandomlyReverted(t, client)
}
