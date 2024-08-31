package message

import (
	"github.com/ethereum/go-ethereum/common"
)

type Storage interface {
	AddMsg(msg Request) error
	GetMsg(msgId common.Hash) (Request, error)
	UpdateMsg(msg Request) error
	UpdateMsgStatus(msgId common.Hash, status MessageStatus) error
}
