package message

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type Broadcaster interface {
	CallAndSendMsg(ctx context.Context, msg Request) (resp Response)
	SendMsg(ctx context.Context, msg Request) (resp Response)
}

// SimpleBroadcaster makes sure that every message broadcasted could be consumed(on-chain) correctly.
type SimpleBroadcaster struct {
	msgManager         Manager
	blockConfirmations uint64
	timeout            time.Duration
}

func NewSimpleBroadcaster(msgManager Manager) *SimpleBroadcaster {
	return &SimpleBroadcaster{
		msgManager:         msgManager,
		blockConfirmations: 0, // TODO:
		timeout:            20 * time.Second,
	}
}

func (b SimpleBroadcaster) CallAndSendMsg(ctx context.Context, msg Request) (resp Response) {
	resp = b.msgManager.CallAndSendMsg(ctx, msg)

	go b.protect(ctx, msg.Id())
	return
}

func (b SimpleBroadcaster) SendMsg(ctx context.Context, msg Request) (resp Response) {
	resp = b.msgManager.SendMsg(ctx, msg)

	go b.protect(ctx, msg.Id())
	return
}

func (b SimpleBroadcaster) protect(ctx context.Context, msgId common.Hash) {
	resp, ok := b.msgManager.WaitMsgResponse(msgId, b.timeout)
	if !ok {
		log.Error("no need to protect error response", "msgId", msgId)
		return
	}

	if resp.Err != nil {
		return
	}

	log.Info("protect msg", "msgId", msgId.Hex(), "txHash", resp.Tx.Hash().Hex(), "resp", *resp)

	txReceipt, ok := b.msgManager.WaitTxReceipt(resp.Tx.Hash(), b.blockConfirmations, b.timeout)
	if !ok {
		b.msgManager.ReplaceMsgWithHigherGasPrice(ctx, msgId)
		b.protect(ctx, msgId)
	} else {
		b.msgManager.UpdateReceipt(msgId, Receipt{Id: msgId, TxReceipt: txReceipt})
	}
}
