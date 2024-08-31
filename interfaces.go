package ethclient

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Message struct {
	id       common.Hash
	From     common.Address  // the sender of the 'transaction', who will be overwrite if private key not nil
	To       *common.Address // the destination contract (nil for contract creation)
	Value    *big.Int        // amount of wei sent along with the call
	Gas      uint64          // if 0, the call executes with near-infinite gas
	GasPrice *big.Int        // wei <-> gas exchange ratio
	Data     []byte          // input data, usually an ABI-encoded contract method invocation

	AccessList types.AccessList // EIP-2930 access list.

	AfterMsg   common.Hash // message id. Used for making sure the msg was executed after it.
	status     MessageStatus
	PrivateKey *ecdsa.PrivateKey // must be not nil if send the message
}

type MessageStatus uint8

const (
	MessageStatusPending MessageStatus = iota
	MessageStatusQueued
	MessageStatusNonceAssigned
	MessageStatusNonceConsumed
	MessageStatusInflight // Broadcasted but not on chain
	MessageStatusOnChain
	MessageStatusFinalized
)

type MessageResponse struct {
	id         common.Hash
	Tx         *types.Transaction
	ReturnData []byte // not nil if using SafeBatchSendMsg and no err
	Err        error
}

type MessageStorage interface {
	AddMsg(msg Message) error
	GetMsg(msgId common.Hash) (Message, error)
	UpdateMsg(msg Message) error
	UpdateMsgStatus(msgId common.Hash, status MessageStatus) error
}

type MessageSequencer interface {
	PushMsg(msg Message) error
	// block if no any msgs return
	PopMsg() (Message, error)
	PeekMsg() (Message, error)
	QueuedMsgCount() (uint, error)
}

// Subscriber represents a set of methods about chain subscription
type Subscriber interface {
	SubscribeFilterlogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) error
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) error
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []types.Log, err error)
}

// TransactFunc represents the transact call of Smart Contract.
type TransactFunc func() (*types.Transaction, error)

// ExpectedEventsFunc returns true if event is expected.
type ExpectedEventsFunc func(event interface{}) bool

type resubscribeFunc func() (ethereum.Subscription, error)
