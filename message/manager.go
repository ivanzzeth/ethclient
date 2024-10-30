package message

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Manager interface {
	Storage
	ScheduleManager
}

type ScheduleManager interface {
	CallAndSendMsg(ctx context.Context, msg Request) (resp Response)
	CallMsg(ctx context.Context, msg Request, blockNumber *big.Int) (resp Response)

	SendMsg(ctx context.Context, msg Request) (resp Response)
	ReplaceMsgWithHigherGasPrice(ctx context.Context, msgId common.Hash) (resp Response)
	// replace old msg with same nonce.
	// mark old one as MessageStatusNonceReleased
	// ReplaceMsg(ctx context.Context, msgId common.Hash, newMsg Request) (resp Response)

	NewTransaction(ctx context.Context, msg Request) (*types.Transaction, error)
	MessageToTransactOpts(ctx context.Context, msg Request) (*bind.TransactOpts, error)

	WaitTxReceipt(txHash common.Hash, confirmations uint64, timeout time.Duration) (*types.Receipt, bool)
	WaitMsgResponse(msgId common.Hash, timeout time.Duration) (*Response, bool)
	WaitMsgReceipt(msgId common.Hash, confirmations uint64, timeout time.Duration) (*Receipt, bool)
}

// type StatusManager interface {
// 	SubmitMsg(msg Request) (err error)
// 	ScheduleMsg(msgId common.Hash) (err error)
// 	QueueMsg(msgId common.Hash) (err error)
// 	AssignNonceToMsg(msgId common.Hash) (err error)
// 	InflightMsg(msgId common.Hash) (err error)
// 	OnChainMsg(msgId common.Hash) (err error)
// 	FinalizeMsg(msgId common.Hash) (err error)
// 	ExpireMsg(msgId common.Hash) (err error)
// }
