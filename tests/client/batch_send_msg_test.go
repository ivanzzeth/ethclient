package client_test

import (
	"testing"

	"github.com/ivanzz/ethclient/tests/helper"
)

func TestBatchSendMsg(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	testScheduleMsg(t, client)
}

func Test_BatchSendMsg_RandomlyReverted(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	test_ScheduleMsg_RandomlyReverted(t, client)
}
