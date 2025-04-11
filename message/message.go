package message

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

type Message struct {
	Root    *common.Hash
	Parent  *common.Hash // it's created by parent if not nil
	Req     *Request
	Resp    *Response // not nil if inflight
	Receipt *Receipt  // not nil if on-chain
	Status  MessageStatus
}

func (m *Message) Id() common.Hash {
	return m.Req.id
}

type Request struct {
	id                    common.Hash
	From                  common.Address  // the sender of the 'transaction'
	To                    *common.Address // the destination contract (nil for contract creation)
	Value                 *big.Int        // amount of wei sent along with the call
	Gas                   uint64          // if 0, the call executes with near-infinite gas
	GasOnEstimationFailed *uint64         // how much gas you wanna provide when the msg estimation failed. As much as possible, so you can debug on-chain
	GasPrice              *big.Int        // wei <-> gas exchange ratio
	Data                  []byte          // input data, usually an ABI-encoded contract method invocation

	AccessList types.AccessList // EIP-2930 access list.

	SimulationOn bool // contains return data of msg call if true
	// ONLY available on function ScheduleMsg
	AfterMsg       *common.Hash  // message id or txHash. Used for making sure the msg was executed after it.
	StartTime      int64         // the msg was executed after the time. It's useful for one-time task.
	ExpirationTime int64         // the msg will be not included on-chain if timeout.
	Interval       time.Duration // the msg will be executed every interval.
}

type MessageStatus uint8

const (
	MessageStatusSubmitted MessageStatus = iota + 1
	MessageStatusScheduled
	MessageStatusQueued
	MessageStatusNonceAssigned
	MessageStatusInflight // Broadcasted but not on chain
	MessageStatusOnChain
	MessageStatusFinalized
	// it was broadcasted but not included on-chain until timeout, so the nonce was released
	MessageStatusNonceReleased
	MessageStatusExpired
)

type Response struct {
	Id         common.Hash
	Tx         *types.Transaction
	ReturnData []byte // not nil if using SafeScheduleMsg and no err
	Err        error
}

type Receipt struct {
	Id        common.Hash
	TxReceipt *types.Receipt
}

func AssignMessageId(msg *Request) *Request {
	uid, _ := uuid.NewUUID()
	uidBytes, _ := uid.MarshalBinary()
	msg.id = crypto.Keccak256Hash(uidBytes)
	return msg
}

func AssignMessageIdWithNonce(msg *Request, nonce int64) *Request {
	msg.id = *GenerateMessageIdByNonce(nonce)
	return msg
}

func GenerateMessageIdByNonce(nonce int64) *common.Hash {
	hash := crypto.Keccak256Hash(big.NewInt(nonce).Bytes())
	return &hash
}

func (m *Request) Id() common.Hash {
	return m.id
}

func (q *Request) SetId(id common.Hash) *Request {
	q.id = id
	return q
}

func (q *Request) SetIdWithNonce(nonce int64) *Request {
	q.id = *GenerateMessageIdByNonce(nonce)
	return q
}

func (q *Request) SetRandomId() *Request {
	AssignMessageId(q)
	return q
}

func (q *Request) Copy() *Request {
	req := q.CopyWithoutId()
	req.id = q.id

	return req
}

func (q *Request) CopyWithoutId() *Request {
	var (
		gasOnEstimationFailed *uint64
		value, gasPrice       *big.Int
	)

	if q.GasOnEstimationFailed != nil {
		gas := *q.GasOnEstimationFailed
		gasOnEstimationFailed = &gas
	}

	if q.Value != nil {
		value = big.NewInt(0).Set(q.Value)
	}

	if q.GasPrice != nil {
		gasPrice = big.NewInt(0).Set(q.GasPrice)
	}

	req := Request{
		From:                  q.From,
		To:                    q.To,
		Value:                 value,
		Gas:                   q.Gas,
		GasOnEstimationFailed: gasOnEstimationFailed,
		GasPrice:              gasPrice,
		Data:                  q.Data,
		AccessList:            q.AccessList,
		SimulationOn:          q.SimulationOn,

		AfterMsg:       q.AfterMsg,
		StartTime:      q.StartTime,
		ExpirationTime: q.ExpirationTime,
		Interval:       q.Interval,
	}

	return &req
}
