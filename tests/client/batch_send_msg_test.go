package client_test

import (
	"testing"

	"github.com/ivanzzeth/ethclient/tests/helper"
)

func TestScheduleMsg(t *testing.T) {
	client := helper.SetUpClient(t)

	testScheduleMsg(t, client)
}

func Test_ScheduleMsg_RandomlyReverted(t *testing.T) {
	client := helper.SetUpClient(t)

	test_ScheduleMsg_RandomlyReverted(t, client)
}
