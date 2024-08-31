package message

import (
	"github.com/ethereum/go-ethereum/common"
)

type Storage interface {
	AddMsg(req Request) error
	GetMsg(msgId common.Hash) (Message, error)
	UpdateMsg(msg Message) error
	UpdateResponse(msgId common.Hash, resp Response) error
	UpdateMsgStatus(msgId common.Hash, status MessageStatus) error
}
