package client_test

import (
	"testing"

	"github.com/ivanzzeth/ethclient/tests/helper"
)

func TestScheduleMsg(t *testing.T) {
	sim := helper.SetUpClient(t)

	testScheduleMsg(t, sim, false)
}

func TestScheduleMsgConcurrent(t *testing.T) {
	sim := helper.SetUpClient(t)

	testScheduleMsg(t, sim, true)
}

func Test_ScheduleMsg_RandomlyReverted(t *testing.T) {
	sim := helper.SetUpClient(t)

	test_ScheduleMsg_RandomlyReverted(t, sim)
}
