package message

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

type Message struct {
	Req    *Request
	Resp   *Response
	Status MessageStatus
}

func (m *Message) Id() common.Hash {
	return m.Req.id
}

type Request struct {
	id       common.Hash
	From     common.Address  // the sender of the 'transaction'
	To       *common.Address // the destination contract (nil for contract creation)
	Value    *big.Int        // amount of wei sent along with the call
	Gas      uint64          // if 0, the call executes with near-infinite gas
	GasPrice *big.Int        // wei <-> gas exchange ratio
	Data     []byte          // input data, usually an ABI-encoded contract method invocation

	AccessList types.AccessList // EIP-2930 access list.

	AfterMsg common.Hash // message id. Used for making sure the msg was executed after it.
}

type MessageStatus uint8

const (
	MessageStatusPending MessageStatus = iota
	MessageStatusQueued
	MessageStatusNonceAssigned
	MessageStatusInflight // Broadcasted but not on chain
	MessageStatusOnChain
	MessageStatusFinalized
)

type Response struct {
	Id         common.Hash
	Tx         *types.Transaction
	ReturnData []byte // not nil if using SafeBatchSendMsg and no err
	Err        error
}

func AssignMessageId(msg *Request) *Request {
	uid, _ := uuid.NewUUID()
	uidBytes, _ := uid.MarshalBinary()
	msg.id = crypto.Keccak256Hash(uidBytes)
	return msg
}

func (m *Request) Id() common.Hash {
	return m.id
}
