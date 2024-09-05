package message

import (
	"errors"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func Test_Sequencer(t *testing.T) {
	// handler := log.NewTerminalHandler(os.Stdout, true)
	// logger := log.NewLogger(handler)
	// log.SetDefault(logger)

	type testcase struct {
		inputs      []Request
		wantPushErr error
	}

	id1 := common.HexToHash("0x1")
	id2 := common.HexToHash("0x2")
	id3 := common.HexToHash("0x3")
	id4 := common.HexToHash("0x4")

	testcases := []testcase{
		{
			inputs: []Request{
				{
					id: id1,
				},
				{
					id: id2,
				},
				{
					id: id3,
				},
				{
					id: id4,
				},
			},
		},
		{
			inputs: []Request{
				{
					id:       id1,
					AfterMsg: &id2,
				},
				{
					id: id2,
				},
				{
					id: id3,
				},
				{
					id:       id4,
					AfterMsg: &id3,
				},
			},
		},
		{
			inputs: []Request{
				{
					id:       id1,
					AfterMsg: &id2,
				},
				{
					id: id2,
				},
				{
					id:       id3,
					AfterMsg: &id4,
				},
				{
					id:       id4,
					AfterMsg: &id1,
				},
			},
		},
		{
			inputs: []Request{
				{
					id:       id1,
					AfterMsg: &id2,
				},
				{
					id:       id2,
					AfterMsg: &id3,
				},
				{
					id:       id3,
					AfterMsg: &id4,
				},
				{
					id:       id4,
					AfterMsg: &id1,
				},
			},
		},
	}

	for i, tt := range testcases {
		t.Logf("run case#%d", i)
		storage, err := NewMemoryStorage()
		if err != nil {
			t.Fatal(err)
		}
		sequencer := NewMemorySequencer(nil, storage, 5)

		pushMsg := func(msg Request) error {
			err := storage.AddMsg(msg)
			if err != nil {
				return err
			}
			err = sequencer.PushMsg(msg)
			if err != nil {
				return err
			}

			return nil
		}

		for _, msg := range tt.inputs {
			err = pushMsg(msg)
			if !errors.Is(err, tt.wantPushErr) {
				t.Fatal("push:", err)
			}
		}

		time.Sleep(2 * time.Second)

		pendingCount, err := sequencer.PendingMsgCount()
		if err != nil {
			t.Fatal("PendingMsgCount:", err)
		}

		t.Log("pendingCount", pendingCount)

		got := []string{}
		for j := 0; j < pendingCount; j++ {
			res, err := sequencer.PopMsg()
			if err != nil {
				t.Fatal("PopMsg:", err)
			}
			got = append(got, res.Id().Hex())
			// assert.Equal(t, tt.wantExecutionSequence[j], res.Id(), "incorrect sequence")
		}

		t.Logf("Got sequence: %v", got)
	}
}
