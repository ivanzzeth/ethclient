package message

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ivanzzeth/ethclient/account"
	"github.com/ivanzzeth/ethclient/nonce"
)

var _ Manager = (*SimpleManager)(nil)

type SimpleManager struct {
	backend ethBackend
	nm      nonce.Manager
	account.Registry
	Storage
}

func NewSimpleManager(backend ethBackend, nm nonce.Manager, accountRegistry account.Registry, storage Storage) *SimpleManager {
	return &SimpleManager{
		backend:  backend,
		nm:       nm,
		Registry: accountRegistry,
		Storage:  storage,
	}
}

func (c *SimpleManager) CallAndSendMsg(ctx context.Context, msg Request) (resp Response) {
	// err := c.AddMsg(msg)
	// if err != nil {
	// 	resp.Err = err
	// 	return
	// }

	return c.callAndSendMsg(ctx, msg)
}

func (c *SimpleManager) CallMsg(ctx context.Context, msg Request, blockNumber *big.Int) (resp Response) {
	resp.Id = msg.Id()

	if msg.AfterMsg != nil {
		resp.Err = fmt.Errorf("field AfterMsg ONLY available on calling ScheduleMsg")
		return
	}

	ethMesg := ethereum.CallMsg{
		From:       msg.From,
		To:         msg.To,
		Gas:        msg.Gas,
		GasPrice:   msg.GasPrice,
		Value:      msg.Value,
		Data:       msg.Data,
		AccessList: msg.AccessList,
	}

	returnData, err := c.backend.CallContract(ctx, ethMesg, blockNumber)
	resp.Err = err
	resp.ReturnData = returnData
	return
}

func (m SimpleManager) SendMsg(ctx context.Context, msg Request) (resp Response) {
	resp.Id = msg.Id()

	// err := m.AddMsg(msg)
	// if err != nil {
	// 	resp.Err = err
	// 	return
	// }

	signedTx, err := m.sendMsg(ctx, msg)
	if err != nil {
		resp.Err = err
		return
	}

	resp = Response{
		Id:  msg.Id(),
		Tx:  signedTx,
		Err: err,
	}

	return
}

func (m SimpleManager) ReplaceMsgWithHigherGasPrice(ctx context.Context, msgId common.Hash) (resp Response) {
	log.Info("replace message with higher gas price", "msgId", msgId)
	resp.Id = msgId

	signedTx, err := m.replaceMsgWithHigherGasPrice(ctx, msgId)
	if err != nil {
		resp.Err = err
		return
	}

	resp = Response{
		Id:  msgId,
		Tx:  signedTx,
		Err: err,
	}

	return
}

// func (m SimpleManager) ReplaceMsg(msgId common.Hash, newMsg Request) (resp Response) {
// 	return Response{}
// }

func (c SimpleManager) NewTransaction(ctx context.Context, msg Request) (*types.Transaction, error) {
	return c.newTransactionWithNonce(ctx, msg, 0)
}

func (c SimpleManager) MessageToTransactOpts(ctx context.Context, msg Request) (*bind.TransactOpts, error) {
	nonce, err := c.nm.PendingNonceAt(ctx, msg.From)
	if err != nil {
		return nil, err
	}

	auth := &bind.TransactOpts{}

	auth.From = msg.From
	auth.Signer = c.GetSigner()
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = msg.Value
	auth.GasLimit = msg.Gas
	auth.GasPrice = msg.GasPrice

	return auth, nil
}

func (c SimpleManager) WaitTxReceipt(txHash common.Hash, confirmations uint64, timeout time.Duration) (*types.Receipt, bool) {
	startTime := time.Now()
	retryCount := 0
	for ; ; retryCount++ {
		log.Debug("wait tx receipt", "txHash", txHash.Hex(), "retryCount", retryCount)
		currTime := time.Now()
		elapsedTime := currTime.Sub(startTime)
		if elapsedTime >= timeout {
			return nil, false
		}

		receipt, err := c.backend.TransactionReceipt(context.Background(), txHash)
		// log.Debug("WaitTxReceipt.TransactionReceipt", "receipt", receipt, "err", err)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		block, err := c.backend.BlockNumber(context.Background())
		// log.Debug("WaitTxReceipt.BlockNumber", "block", block, "err", err)

		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if block >= receipt.BlockNumber.Uint64()+confirmations {
			return receipt, true
		}

		time.Sleep(1 * time.Second)
	}
}

func (c SimpleManager) WaitMsgResponse(msgId common.Hash, timeout time.Duration) (*Response, bool) {
	startTime := time.Now()
	for {
		log.Debug("wait msg response", "msgId", msgId.Hex())
		currTime := time.Now()
		elapsedTime := currTime.Sub(startTime)
		if elapsedTime >= timeout {
			return nil, false
		}

		msg, err := c.GetMsg(msgId)
		if err != nil {
			// log.Debug("WaitMsgResponse GetMsg failed", "err", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if msg.Resp == nil {
			// log.Debug("WaitMsgResponse msg.Resp is nil")
			time.Sleep(1 * time.Second)
			continue
		}

		return msg.Resp, true
	}
}

func (c SimpleManager) WaitMsgReceipt(msgId common.Hash, confirmations uint64, timeout time.Duration) (*Receipt, bool) {
	startTime := time.Now()
	for {
		currTime := time.Now()
		elapsedTime := currTime.Sub(startTime)
		if elapsedTime >= timeout {
			return nil, false
		}

		msg, err := c.GetMsg(msgId)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if msg.Resp != nil && msg.Resp.Tx != nil {
			log.Debug("wait msg receipt", "msgId", msgId.Hex(), "txHash", msg.Resp.Tx.Hash().Hex())
		}

		if msg.Receipt == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		receipt, err := c.backend.TransactionReceipt(context.Background(), msg.Receipt.TxReceipt.TxHash)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		block, err := c.backend.BlockNumber(context.Background())
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if block >= receipt.BlockNumber.Uint64()+confirmations {
			return msg.Receipt, true
		}

		time.Sleep(1 * time.Second)
	}
}

func (c *SimpleManager) callAndSendMsg(ctx context.Context, msg Request) (resp Response) {
	resp = c.CallMsg(ctx, msg, nil)
	if resp.Err != nil {
		return
	}

	tx, err := c.sendMsg(ctx, msg)
	if err != nil {
		resp.Err = err
		return
	}

	resp.Tx = tx

	return
}

func (m SimpleManager) sendMsg(ctx context.Context, msg Request) (signedTx *types.Transaction, err error) {
	log.Debug("broadcast msg", "msg", msg)
	tx, err := m.NewTransaction(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("NewTransaction err: %v", err)
	}

	err = m.UpdateMsgStatus(msg.Id(), MessageStatusNonceAssigned)
	if err != nil {
		return nil, err
	}

	signedTx, err = m.signMsgAndBroadcast(ctx, msg.Id(), msg.From, tx)

	if err != nil {
		return nil, err
	}

	log.Info("Send Message successfully", "msgId", msg.Id(), "txHash", signedTx.Hash().Hex(), "from", msg.From.Hex(),
		"to", msg.To.Hex(), "value", msg.Value, "nonce", signedTx.Nonce())

	return signedTx, nil
}

func (m SimpleManager) replaceMsgWithHigherGasPrice(ctx context.Context, msgId common.Hash) (signedTx *types.Transaction, err error) {
	log.Debug("replace msg with higher gas price", "msgId", msgId)
	msg, err := m.GetMsg(msgId)
	if err != nil {
		return nil, err
	}

	if msg.Resp == nil || msg.Resp.Tx == nil {
		return nil, fmt.Errorf("no nonce assigned")
	}

	msg.Req.GasPrice = big.NewInt(0).Mul(msg.Resp.Tx.GasPrice(), big.NewInt(12))
	msg.Req.GasPrice.Div(msg.Req.GasPrice, big.NewInt(10))

	tx, err := m.newTransactionWithNonce(ctx, *msg.Req, msg.Resp.Tx.Nonce())
	if err != nil {
		return nil, fmt.Errorf("NewTransaction err: %v", err)
	}

	err = m.UpdateMsgStatus(msg.Id(), MessageStatusNonceAssigned)
	if err != nil {
		return nil, err
	}

	signedTx, err = m.signMsgAndBroadcast(ctx, msg.Id(), msg.Req.From, tx)

	if err != nil {
		return nil, err
	}

	log.Info("Replace and send Message successfully", "msgId", msgId, "txHash", signedTx.Hash().Hex(), "from", msg.Req.From.Hex(),
		"to", msg.Req.To.Hex(), "value", msg.Req.Value)

	return signedTx, nil
}

func (m SimpleManager) signMsgAndBroadcast(ctx context.Context, msgId common.Hash, from common.Address, tx *types.Transaction) (signedTx *types.Transaction, err error) {
	// chainID, err := c.Client.ChainID(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("get Chain ID err: %v", err)
	// }

	// signedTx, err = types.SignTx(tx, types.NewEIP2930Signer(chainID), msg.PrivateKey)
	// if err != nil {
	// 	return nil, fmt.Errorf("SignTx err: %v", err)
	// }

	signerFn := m.GetSigner()
	signedTx, err = signerFn(from, tx)
	if err != nil {
		return nil, err
	}

	err = m.backend.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("SendTransaction err: %v", err)
	}
	log.Info("broadcasted transaction", "txHash", signedTx.Hash().Hex(), "from", from, "nonce", tx.Nonce())

	err = m.UpdateMsgStatus(msgId, MessageStatusInflight)
	if err != nil {
		return nil, err
	}

	return
}

func (c SimpleManager) newTransactionWithNonce(ctx context.Context, msg Request, nonce uint64) (tx *types.Transaction, err error) {
	if msg.To == nil {
		to := common.HexToAddress("0x0")
		msg.To = &to
	}

	if msg.Gas == 0 {
		ethMesg := ethereum.CallMsg{
			From:       msg.From,
			To:         msg.To,
			GasPrice:   msg.GasPrice,
			Value:      msg.Value,
			Data:       msg.Data,
			AccessList: msg.AccessList,
		}

		gas, err := c.nm.EstimateGas(ctx, ethMesg)
		if err != nil {
			if msg.GasOnEstimationFailed == nil {
				return nil, err
			}

			msg.Gas = *msg.GasOnEstimationFailed
		} else {
			// Multiplier 1.5
			msg.Gas = gas * 1500 / 1000
			// reach out max gas, then replace gas estimated with GasOnEstimationFailed
			if msg.GasOnEstimationFailed != nil && msg.Gas > *msg.GasOnEstimationFailed {
				log.Warn("reach out max gas, then replace gas estimated with GasOnEstimationFailed", "msgId",
					msg.Id().Hex(), "gasOnEstimationFailed", *msg.GasOnEstimationFailed)
				msg.Gas = *msg.GasOnEstimationFailed
			}
		}
	}

	if msg.GasPrice == nil || msg.GasPrice.Uint64() == 0 {
		var err error
		msg.GasPrice, err = c.nm.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
	}

	if nonce == 0 {
		nonce, err = c.nm.PendingNonceAt(ctx, msg.From)
		if err != nil {
			return nil, err
		}
	}

	log.Debug("nonce assign msg", "nonce", nonce, "ID", msg.Id())

	tx = types.NewTransaction(nonce, *msg.To, msg.Value, msg.Gas, msg.GasPrice, msg.Data)

	return
}
