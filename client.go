package ethclient

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ivanzzeth/ethclient/account"
	"github.com/ivanzzeth/ethclient/common/consts"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/nonce"
	"github.com/ivanzzeth/ethclient/subscriber"
)

var _ bind.ContractBackend = (*Client)(nil)
var _ bind.PendingContractCaller = (*Client)(nil)
var _ bind.DeployBackend = (*Client)(nil)

var _ ethereum.ChainReader = (*Client)(nil)
var _ ethereum.TransactionReader = (*Client)(nil)
var _ ethereum.ChainStateReader = (*Client)(nil)
var _ ethereum.ContractCaller = (*Client)(nil)
var _ ethereum.LogFilterer = (*Client)(nil)
var _ ethereum.TransactionSender = (*Client)(nil)
var _ ethereum.GasPricer = (*Client)(nil)
var _ ethereum.GasPricer1559 = (*Client)(nil)
var _ ethereum.FeeHistoryReader = (*Client)(nil)
var _ ethereum.PendingStateReader = (*Client)(nil)
var _ ethereum.PendingContractCaller = (*Client)(nil)
var _ ethereum.GasEstimator = (*Client)(nil)

// TODO:
// var _ ethereum.PendingStateEventer = (*Client)(nil)
var _ ethereum.BlockNumberReader = (*Client)(nil)
var _ ethereum.ChainIDReader = (*Client)(nil)

type Client struct {
	*ethclient.Client
	rpcClient   *rpc.Client
	accRegistry account.Registry
	nonce.Manager
	msgManager      message.Manager
	broadcaster     message.Broadcaster
	reqChannel      chan message.Request
	scheduleChannel chan message.Request
	respChannel     chan message.Response
	receiptChannel  chan message.Receipt
	msgBuffer       int
	msgStore        message.Storage
	msgSequencer    message.Sequencer
	abi             abi.ABI

	subscriber.Subscriber
}

func NewMemoryClient(c *rpc.Client) (*Client, error) {
	ethc := ethclient.NewClient(c)

	chainId, err := ethc.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	accRegistry := account.NewSimpleRegistry(chainId)

	nm, err := nonce.NewSimpleManager(ethc, nonce.NewMemoryStorage())
	if err != nil {
		return nil, err
	}

	msgStore, err := message.NewMemoryStorage()
	if err != nil {
		return nil, err
	}

	msgSequencer := message.NewMemorySequencer(ethc, msgStore, consts.DefaultMsgBuffer)

	msgManager := message.NewSimpleManager(ethc, nm, accRegistry, msgStore)

	subscriber, err := subscriber.NewChainSubscriber(ethc, subscriber.NewMemoryStorage(chainId))
	if err != nil {
		return nil, err
	}

	return NewClient(c, accRegistry, msgStore, nm, msgManager, subscriber, msgSequencer)
}

func NewClient(
	c *rpc.Client,
	accRegistry account.Registry,
	msgStore message.Storage,
	nonceManager nonce.Manager,
	msgManager message.Manager,
	subscriber subscriber.Subscriber,
	sequencer message.Sequencer,
) (*Client, error) {
	ethc := ethclient.NewClient(c)

	cli := &Client{
		Client:          ethc,
		rpcClient:       c,
		accRegistry:     accRegistry,
		reqChannel:      make(chan message.Request, consts.DefaultMsgBuffer),
		scheduleChannel: make(chan message.Request, consts.DefaultMsgBuffer),
		respChannel:     make(chan message.Response, consts.DefaultMsgBuffer),
		receiptChannel:  make(chan message.Receipt, consts.DefaultMsgBuffer),
		msgBuffer:       consts.DefaultMsgBuffer,
		msgStore:        msgStore,
		msgSequencer:    sequencer,
		Manager:         nonceManager,
		msgManager:      msgManager,
		broadcaster:     message.NewSimpleBroadcaster(msgManager),
		Subscriber:      subscriber,
	}

	go cli.sendMsgTask(context.Background())

	return cli, nil
}

func (c *Client) Close() {
	log.Info("close client..")

	c.CloseSendMsg()

	c.Subscriber.Close()

	log.Debug("subscriber closed")

	c.Client.Close()

	log.Debug("underlying ethclient closed")

	log.Info("client closed..")
}

func (c *Client) CloseSendMsg() {
	close(c.reqChannel)
	log.Info("reqChannel closed")
}

// RawClient returns underlying ethclient
func (c *Client) RawClient() *ethclient.Client {
	return c.Client
}

func (c *Client) SetNonceManager(nm nonce.Manager) {
	c.Manager = nm
}

func (c *Client) SetSubscriber(s subscriber.Subscriber) {
	c.Subscriber = s
}

func (c *Client) GetSigner() bind.SignerFn {
	return c.accRegistry.GetSigner()
}

func (c *Client) RegisterSigner(signerFn bind.SignerFn) {
	c.accRegistry.RegisterSigner(signerFn)
}

// Registers the private key used for signing txs.
func (c *Client) RegisterPrivateKey(ctx context.Context, key *ecdsa.PrivateKey) error {
	return c.accRegistry.RegisterPrivateKey(ctx, key)
}

func (c *Client) SetMsgBuffer(buffer int) {
	c.msgBuffer = buffer
}

func (c *Client) GetMsg(msgId common.Hash) (message.Message, error) {
	return c.msgStore.GetMsg(msgId)
}

func (c *Client) NewMethodData(a abi.ABI, methodName string, args ...interface{}) ([]byte, error) {
	return a.Pack(methodName, args...)
}

func (c *Client) ScheduleMsg(req *message.Request) {
	c.reqChannel <- *req.Copy()
}

func (c *Client) ReplayMsg(msgId common.Hash) (newMsgId common.Hash, err error) {
	msg, err := c.msgStore.GetMsg(msgId)
	if err != nil {
		return
	}

	copiedReq := msg.Req.CopyWithoutId()

	message.AssignMessageId(copiedReq)

	c.reqChannel <- *copiedReq

	newMsgId = copiedReq.Id()
	return
}

func (c *Client) Response() <-chan message.Response {
	return c.respChannel
}

func (c *Client) CallMsg(ctx context.Context, msg message.Request, blockNumber *big.Int) (returnData []byte, err error) {
	resp := c.msgManager.CallMsg(ctx, msg, blockNumber)
	return resp.ReturnData, resp.Err
}

func (c *Client) sendMsgTask(ctx context.Context) {
	// Pipepline: reqChannel => scheduler => sequencer => broadcaster

	go c.schedule()

	go c.sequence()

	go c.broadcast(ctx)
}

func (c *Client) schedule() {
	for req := range c.reqChannel {
		log.Debug("start scheduling msg...", "msgId", req.Id())
		func() {
			var err error
			var resp message.Response
			resp.Id = req.Id()
			defer func() {
				if err != nil {
					log.Debug("Client.schedule UpdateResponse", "resp", resp)

					resp.Err = err
					c.msgStore.UpdateResponse(resp.Id, resp)
					c.respChannel <- resp
				}
			}()

			if req.Id() == common.BytesToHash([]byte{}) {
				panic(fmt.Errorf("no msgId provided"))
			}

			if c.msgStore.HasMsg(req.Id()) {
				resp.Err = fmt.Errorf("already known")
				return
			}

			// message.MessageStatusSubmitted
			err = c.msgStore.AddMsg(req)
			if err != nil {
				err = fmt.Errorf("no msgId provided: %v", err)
				return
			}

			msg, err := c.msgStore.GetMsg(req.Id())
			if err != nil {
				return
			}

			now := time.Now().UnixNano()

			if msg.Req.Interval != 0 {
				if msg.Req.StartTime == 0 {
					msg.Req.StartTime = now
				}
				err = c.msgStore.UpdateMsg(msg)
				if err != nil {
					return
				}
			}

			if msg.Req.ExpirationTime != 0 && msg.Req.ExpirationTime < now {
				// timeout
				err = c.msgStore.UpdateMsgStatus(req.Id(), message.MessageStatusExpired)
				if err != nil {
					return
				}

				err = fmt.Errorf("timeout")
			}

			req = *msg.Req

			if msg.Req.StartTime >= now {
				log.Debug("scheduler found it's not time for executing the msg", "msg", msg.Id().Hex())
				go func() {
					duration := msg.Req.StartTime - time.Now().UnixNano()
					time.Sleep(time.Duration(duration))
					c.reqChannel <- *msg.Req
				}()
				return
			}

			c.scheduleChannel <- *msg.Req

			err = c.msgStore.UpdateMsgStatus(req.Id(), message.MessageStatusScheduled)
			if err != nil {
				err = fmt.Errorf("no msgId provided")
				return
			}

			if msg.Req.Interval != 0 {
				newReq := req.CopyWithoutId()

				newReq.AfterMsg = nil
				newReq.StartTime = now + int64(req.Interval)

				message.AssignMessageId(newReq)
				log.Debug("scheduler creates new one for long-term ticker task", "msg", msg.Id().Hex(), "new_msg", newReq.Id().Hex())

				err = c.msgStore.AddMsg(*newReq)
				if err != nil {
					return
				}

				newMsg, err := c.msgStore.GetMsg(newReq.Id())
				if err != nil {
					resp.Id = newReq.Id()
					return
				}
				parent := req.Id()
				newMsg.Parent = &parent
				if msg.Root != nil {
					newMsg.Root = msg.Root
				} else {
					newMsg.Root = &parent
				}

				err = c.msgStore.UpdateMsg(newMsg)
				if err != nil {
					resp.Id = newReq.Id()
					return
				}

				c.reqChannel <- *newReq
			}
		}()
	}

	log.Debug("close scheduler...")
	close(c.scheduleChannel)
}

func (c *Client) sequence() {
	for msg := range c.scheduleChannel {
		log.Debug("sequence msg...", "msg", msg)
		func() {
			var err error
			var resp message.Response
			resp.Id = msg.Id()
			defer func() {
				if err != nil {
					log.Debug("Client.sequence UpdateResponse", "resp", resp)

					resp.Err = err
					c.msgStore.UpdateResponse(resp.Id, resp)
					c.respChannel <- resp
				}
			}()
			err = c.msgSequencer.PushMsg(msg)
			if err != nil {
				return
			}

			err = c.msgStore.UpdateMsgStatus(msg.Id(), message.MessageStatusQueued)
			if err != nil {
				return
			}
		}()
	}

	log.Debug("close sequencer...")

	c.msgSequencer.Close()
}

func (c *Client) broadcast(ctx context.Context) {
	for {
		exit := func() (exit bool) {
			var err error

			msg, err := c.msgSequencer.PopMsg()
			if err != nil {
				if errors.Is(err, message.ErrPendingChannelClosed) {
					log.Debug("close responseChannel...")
					close(c.respChannel)
					close(c.receiptChannel)
					return true
				}
				log.Error("unexpected broadcast case", "err", err)
				return true
			}

			var resp message.Response
			resp.Id = msg.Id()
			defer func() {
				log.Debug("Client.broadcast UpdateResponse", "resp", resp, "msgId", msg.Id())

				c.msgStore.UpdateResponse(resp.Id, resp)
				c.respChannel <- resp
			}()

			if msg.SimulationOn {
				resp = c.msgManager.CallMsg(ctx, msg, nil)
			}

			if resp.Err == nil {
				sendResp := c.broadcaster.SendMsg(ctx, msg)
				log.Debug("broadcaster.SendMsg resp", "resp", sendResp)
				resp.Id = sendResp.Id
				resp.Err = sendResp.Err
				resp.Tx = sendResp.Tx
			}

			return
		}()

		if exit {
			return
		}
	}
}

func (c *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return c.Manager.PendingNonceAt(ctx, account)
}

func (c *Client) SuggestGasPrice(ctx context.Context) (gasPrice *big.Int, err error) {
	return c.Manager.SuggestGasPrice(ctx)
}

func (c *Client) WaitTxReceipt(txHash common.Hash, confirmations uint64, timeout time.Duration) (*types.Receipt, bool) {
	return c.msgManager.WaitTxReceipt(txHash, confirmations, timeout)
}

func (c *Client) WaitMsgResponse(msgId common.Hash, timeout time.Duration) (*message.Response, bool) {
	return c.msgManager.WaitMsgResponse(msgId, timeout)
}

func (c *Client) WaitMsgReceipt(msgId common.Hash, confirmations uint64, timeout time.Duration) (*message.Receipt, bool) {
	return c.msgManager.WaitMsgReceipt(msgId, confirmations, timeout)
}

// MessageToTransactOpts .
// NOTE: You must provide private key for signature.
func (c *Client) MessageToTransactOpts(ctx context.Context, msg message.Request) (*bind.TransactOpts, error) {
	return c.msgManager.MessageToTransactOpts(ctx, msg)
}

func (c *Client) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	return c.Subscriber.SubscribeFilterLogs(ctx, query, ch)
}

func (c *Client) FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []types.Log, err error) {
	return c.Subscriber.FilterLogs(ctx, q)
}

func (c *Client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return c.Subscriber.SubscribeNewHead(ctx, ch)
}

func (c *Client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	gas, err := c.Manager.EstimateGas(ctx, msg)
	if err != nil {
		return 0, c.DecodeJsonRpcError(err)
	}

	return gas, nil
}

func (c *Client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	ret, err := c.Client.CallContract(ctx, msg, blockNumber)
	if err != nil {
		return nil, c.DecodeJsonRpcError(err)
	}

	return ret, nil
}

func (c *Client) CallContractAtHash(ctx context.Context, msg ethereum.CallMsg, blockHash common.Hash) ([]byte, error) {
	ret, err := c.Client.CallContractAtHash(ctx, msg, blockHash)
	if err != nil {
		return nil, c.DecodeJsonRpcError(err)
	}

	return ret, nil
}

func (c *Client) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	ret, err := c.Client.PendingCallContract(ctx, msg)
	if err != nil {
		return nil, c.DecodeJsonRpcError(err)
	}

	return ret, nil
}

func (c *Client) DebugTransactionOnChain(ctx context.Context, txHash common.Hash) ([]byte, error) {
	receipt, confirmed := c.WaitTxReceipt(txHash, 3, 30*time.Second)
	if !confirmed {
		return nil, fmt.Errorf("tx receipt not found")
	}

	tx, _, err := c.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, err
	}

	from, err := c.TransactionSender(ctx, tx, receipt.BlockHash, receipt.TransactionIndex)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		From:          from,
		To:            tx.To(),
		Value:         tx.Value(),
		Data:          tx.Data(),
		Gas:           tx.Gas(),
		GasPrice:      tx.GasPrice(),
		AccessList:    tx.AccessList(),
		BlobGasFeeCap: tx.BlobGasFeeCap(),
		BlobHashes:    tx.BlobHashes(),
	}
	ret, err := c.CallContractAtHash(ctx, msg, receipt.BlockHash)

	return ret, err
}

// Trying to decode some data using the abi if specific
func (c *Client) AddABI(intf abi.ABI) {
	if c.abi.Errors == nil {
		c.abi.Errors = make(map[string]abi.Error)
	}

	if c.abi.Methods == nil {
		c.abi.Methods = make(map[string]abi.Method)
	}

	if c.abi.Events == nil {
		c.abi.Events = make(map[string]abi.Event)
	}

	// errors
	for k, v := range intf.Errors {
		c.abi.Errors[k] = v
	}

	for k, v := range intf.Methods {
		c.abi.Methods[k] = v
	}

	for k, v := range intf.Events {
		c.abi.Events[k] = v
	}
}

func (c *Client) DecodeJsonRpcError(err error) error {
	jsonErr := &consts.JsonRpcError{Abi: c.abi, RawError: err.Error()}
	ec, ok := err.(rpc.Error)
	if ok {
		jsonErr.Code = ec.ErrorCode()
	}

	de, ok := err.(rpc.DataError)
	if ok {
		jsonErr.Data = de.ErrorData()
	}

	return jsonErr
}
