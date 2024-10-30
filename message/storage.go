package message

import (
	"github.com/ethereum/go-ethereum/common"
)

type Storage interface {
	StorageReader
	StorageWriter
}

type StorageReader interface {
	HasMsg(msgId common.Hash) bool
	GetMsg(msgId common.Hash) (Message, error)
	GetNonce(msgId common.Hash) (uint64, error)
}

type StorageWriter interface {
	AddMsg(req Request) error
	UpdateMsg(msg Message) error
	UpdateResponse(msgId common.Hash, resp Response) error
	UpdateReceipt(msgId common.Hash, receipt Receipt) error
	UpdateMsgStatus(msgId common.Hash, status MessageStatus) error
}
