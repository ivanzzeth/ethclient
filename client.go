package ethclient

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/uuid"
	"github.com/ivanzz/ethclient/nonce"
)

type Client struct {
	*ethclient.Client
	rpcClient *rpc.Client
	nonce.Manager
	msgBuffer int
	abi       abi.ABI
	Subscriber
}

func Dial(rawurl string) (*Client, error) {
	rpcClient, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	c := ethclient.NewClient(rpcClient)

	nm, err := nonce.NewSimpleManager(c, nonce.NewMemoryStorage())
	if err != nil {
		return nil, err
	}

	subscriber, err := NewChainSubscriber(c)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:     c,
		rpcClient:  rpcClient,
		msgBuffer:  DefaultMsgBuffer,
		Manager:    nm,
		Subscriber: subscriber,
	}, nil
}

func DialWithNonceManager(rawurl string, nm nonce.Manager) (*Client, error) {
	rpcClient, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	c := ethclient.NewClient(rpcClient)

	subscriber, err := NewChainSubscriber(c)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:     c,
		rpcClient:  rpcClient,
		msgBuffer:  DefaultMsgBuffer,
		Manager:    nm,
		Subscriber: subscriber,
	}, nil
}

func NewClient(c *rpc.Client) (*Client, error) {
	ethc := ethclient.NewClient(c)

	nm, err := nonce.NewSimpleManager(ethc, nonce.NewMemoryStorage())
	if err != nil {
		return nil, err
	}

	subscriber, err := NewChainSubscriber(ethc)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:     ethc,
		rpcClient:  c,
		msgBuffer:  DefaultMsgBuffer,
		Manager:    nm,
		Subscriber: subscriber,
	}, nil
}

func (c *Client) Close() {
	c.Client.Close()
}

// RawClient returns underlying ethclient
func (c *Client) RawClient() *ethclient.Client {
	return c.Client
}

func (c *Client) SetNonceManager(nm nonce.Manager) {
	c.Manager = nm
}

func (c *Client) SetMsgBuffer(buffer int) {
	c.msgBuffer = buffer
}

type Message struct {
	Id       uuid.UUID
	From     common.Address  // the sender of the 'transaction', who will be overwrite if private key not nil
	To       *common.Address // the destination contract (nil for contract creation)
	Gas      uint64          // if 0, the call executes with near-infinite gas
	GasPrice *big.Int        // wei <-> gas exchange ratio
	Value    *big.Int        // amount of wei sent along with the call
	Data     []byte          // input data, usually an ABI-encoded contract method invocation

	AccessList types.AccessList // EIP-2930 access list.

	PrivateKey *ecdsa.PrivateKey // must be not nil if send the message
}

func AssignMessageId(msg *Message) *Message {
	msg.Id, _ = uuid.NewUUID()
	return msg
}

type MessageResponse struct {
	Id         uuid.UUID
	Tx         *types.Transaction
	ReturnData []byte // not nil if using SafeBatchSendMsg and no err
	Err        error
}

func (c *Client) NewMethodData(a abi.ABI, methodName string, args ...interface{}) ([]byte, error) {
	return a.Pack(methodName, args...)
}

func (c *Client) SafeBatchSendMsg(ctx context.Context, msgs <-chan Message) <-chan MessageResponse {
	msgResChan := make(chan MessageResponse, c.msgBuffer)
	go func() {
		for msg := range msgs {
			tx, returnData, err := c.SafeSendMsg(ctx, msg)
			// TODO: Release nonce...
			msgResChan <- MessageResponse{
				Id:         msg.Id,
				Tx:         tx,
				ReturnData: returnData,
				Err:        err,
			}
		}

		close(msgResChan)
	}()

	return msgResChan
}

func (c *Client) BatchSendMsg(ctx context.Context, msgs <-chan Message) <-chan MessageResponse {
	msgResChan := make(chan MessageResponse, c.msgBuffer)
	go func() {
		for msg := range msgs {
			tx, err := c.SendMsg(ctx, msg)
			msgResChan <- MessageResponse{
				Id:  msg.Id,
				Tx:  tx,
				Err: err,
			}
		}

		close(msgResChan)
	}()

	return msgResChan
}

func (c *Client) CallMsg(ctx context.Context, msg Message, blockNumber *big.Int) (returnData []byte, err error) {
	if msg.PrivateKey != nil {
		msg.From = crypto.PubkeyToAddress(msg.PrivateKey.PublicKey)
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

	return c.Client.CallContract(ctx, ethMesg, blockNumber)
}

func (c *Client) SafeSendMsg(ctx context.Context, msg Message) (*types.Transaction, []byte, error) {
	returnData, err := c.CallMsg(ctx, msg, nil)
	if err != nil {
		return nil, nil, err
	}

	tx, err := c.SendMsg(ctx, msg)
	if err != nil {
		return nil, nil, err
	}

	return tx, returnData, err
}

func (c *Client) SendMsg(ctx context.Context, msg Message) (signedTx *types.Transaction, err error) {
	if msg.PrivateKey == nil {
		return nil, ErrMessagePrivateKeyNil
	}

	msg.From = crypto.PubkeyToAddress(msg.PrivateKey.PublicKey)

	ethMesg := ethereum.CallMsg{
		From:       msg.From,
		To:         msg.To,
		Gas:        msg.Gas,
		GasPrice:   msg.GasPrice,
		Value:      msg.Value,
		Data:       msg.Data,
		AccessList: msg.AccessList,
	}

	tx, err := c.NewTransaction(ctx, ethMesg)
	if err != nil {
		return nil, fmt.Errorf("NewTransaction err: %v", err)
	}

	chainID, err := c.Client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get Chain ID err: %v", err)
	}

	signedTx, err = types.SignTx(tx, types.NewEIP2930Signer(chainID), msg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("SignTx err: %v", err)
	}

	err = c.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("SendTransaction err: %v", err)
	}

	log.Debug("Send Message successfully", "txHash", signedTx.Hash().Hex(), "from", msg.From.Hex(),
		"to", msg.To.Hex(), "value", msg.Value)

	return signedTx, nil
}

func (c *Client) NewTransaction(ctx context.Context, msg ethereum.CallMsg) (*types.Transaction, error) {
	if msg.To == nil {
		to := common.HexToAddress("0x0")
		msg.To = &to
	}

	if msg.Gas == 0 {
		gas, err := c.EstimateGas(ctx, msg)
		if err != nil {
			return nil, err
		}

		// Multiplier 1.5
		msg.Gas = gas * 1500 / 1000
	}

	if msg.GasPrice == nil || msg.GasPrice.Uint64() == 0 {
		var err error
		msg.GasPrice, err = c.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
	}

	nonce, err := c.PendingNonceAt(ctx, msg.From)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(nonce, *msg.To, msg.Value, msg.Gas, msg.GasPrice, msg.Data)

	return tx, nil
}

func (c *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return c.Manager.PendingNonceAt(ctx, account)
}

func (c *Client) SuggestGasPrice(ctx context.Context) (gasPrice *big.Int, err error) {
	return c.Manager.SuggestGasPrice(ctx)
}

func (c *Client) WaitTxReceipt(txHash common.Hash, confirmations uint64, timeout time.Duration) (*types.Receipt, bool) {
	startTime := time.Now()
	for {
		currTime := time.Now()
		elapsedTime := currTime.Sub(startTime)
		if elapsedTime >= timeout {
			return nil, false
		}

		receipt, err := c.Client.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			continue
		}

		block, err := c.Client.BlockNumber(context.Background())
		if err != nil {
			continue
		}

		if block >= receipt.BlockNumber.Uint64()+confirmations {
			return receipt, true
		}
	}
}

// MessageToTransactOpts .
// NOTE: You must provide private key for signature.
func (c *Client) MessageToTransactOpts(ctx context.Context, msg Message) (*bind.TransactOpts, error) {
	if msg.PrivateKey == nil {
		return nil, ErrMessagePrivateKeyNil
	}
	msg.From = crypto.PubkeyToAddress(msg.PrivateKey.PublicKey)

	nonce, err := c.PendingNonceAt(ctx, msg.From)
	if err != nil {
		return nil, err
	}

	chainID, err := c.Client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(msg.PrivateKey, chainID)
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = msg.Value
	auth.GasLimit = msg.Gas
	auth.GasPrice = msg.GasPrice

	return auth, nil
}

func (c *Client) SubscribeFilterlogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) error {
	return c.Subscriber.SubscribeFilterlogs(ctx, query, ch)
}

func (c *Client) FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []types.Log, err error) {
	return c.Subscriber.FilterLogs(ctx, q)
}

func (c *Client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) error {
	return c.Subscriber.SubscribeNewHead(ctx, ch)
}

func (c *Client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	gas, err := c.Client.EstimateGas(ctx, msg)
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

// TODO: Implement c.Client.PendingCallContract()

// Trying to decode some data using the abi if specific
func (c *Client) SetABI(abi abi.ABI) {
	c.abi = abi
}

func (c *Client) DecodeJsonRpcError(err error) error {
	jsonErr := &JsonRpcError{abi: c.abi}
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
