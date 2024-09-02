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
	"github.com/ivanzz/ethclient/message"
	"github.com/ivanzz/ethclient/nonce"
)

type Client struct {
	*ethclient.Client
	rpcClient *rpc.Client
	nonce.Manager
	msgBuffer    int
	msgStore     message.Storage
	msgSequencer message.Sequencer
	abi          abi.ABI
	signers      []bind.SignerFn // Method to use for signing the transaction (mandatory)

	Subscriber
}

func Dial(rawurl string) (*Client, error) {
	rpcClient, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	return NewClient(rpcClient)
}

func NewClient(c *rpc.Client) (*Client, error) {
	ethc := ethclient.NewClient(c)

	nm, err := nonce.NewSimpleManager(ethc, nonce.NewMemoryStorage())
	if err != nil {
		return nil, err
	}

	msgStore, err := message.NewMemoryStorage()
	if err != nil {
		return nil, err
	}

	msgSequencer := message.NewMemorySequencer(msgStore, 10)

	subscriber, err := NewChainSubscriber(ethc)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:       ethc,
		rpcClient:    c,
		msgBuffer:    DefaultMsgBuffer,
		msgStore:     msgStore,
		msgSequencer: msgSequencer,
		Manager:      nm,
		Subscriber:   subscriber,
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

func (c *Client) GetSigner() bind.SignerFn {
	// combine all signerFn
	return func(a common.Address, t *types.Transaction) (tx *types.Transaction, err error) {
		if len(c.signers) == 0 {
			return nil, fmt.Errorf("no signerFn registered")
		}

		for i, fn := range c.signers {
			tx, err = fn(a, t)
			log.Debug("try to call signerFn", "index", i, "err", err, "account", a)

			if err != nil {
				continue
			}

			return tx, nil
		}

		return nil, bind.ErrNotAuthorized
	}
}

func (c *Client) RegisterSigner(signerFn bind.SignerFn) {
	log.Info("register signerFn for signing...")
	c.signers = append(c.signers, signerFn)
}

// Registers the private key used for signing txs.
func (c *Client) RegisterPrivateKey(ctx context.Context, key *ecdsa.PrivateKey) error {
	chainID, err := c.ChainID(ctx)
	if err != nil {
		return err
	}
	keyAddr := crypto.PubkeyToAddress(key.PublicKey)
	if chainID == nil {
		return bind.ErrNoChainID
	}
	signer := types.LatestSignerForChainID(chainID)
	signerFn := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != keyAddr {
			return nil, bind.ErrNotAuthorized
		}
		signature, err := crypto.Sign(signer.Hash(tx).Bytes(), key)
		if err != nil {
			return nil, err
		}
		return tx.WithSignature(signer, signature)
	}

	c.RegisterSigner(signerFn)

	return nil
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

func (c *Client) SafeBatchSendMsg(ctx context.Context, msgs <-chan message.Request) <-chan message.Response {
	msgResChan := make(chan message.Response, c.msgBuffer)
	go func() {
		for msg := range msgs {
			tx, returnData, err := c.SafeSendMsg(ctx, msg)
			// TODO: Release nonce...
			msgResChan <- message.Response{
				Id:         msg.Id(),
				Tx:         tx,
				ReturnData: returnData,
				Err:        err,
			}
		}

		close(msgResChan)
	}()

	return msgResChan
}

func (c *Client) BatchSendMsg(ctx context.Context, msgs <-chan message.Request) <-chan message.Response {
	msgResChan := make(chan message.Response, c.msgBuffer)
	go func() {
		for msg := range msgs {
			resp := message.Response{Id: msg.Id()}

			err := c.msgSequencer.PushMsg(msg)
			if err != nil {
				resp.Err = err
				msgResChan <- resp
				continue
			}

			err = c.msgStore.UpdateMsgStatus(msg.Id(), message.MessageStatusQueued)
			if err != nil {
				resp.Err = err
				msgResChan <- resp
				continue
			}
		}

		close(msgResChan)
	}()

	go func() {
		for {
			msg, err := c.msgSequencer.PopMsg()
			if err != nil {
				msgResChan <- message.Response{Id: msg.Id(), Err: err}
				continue
			}

			tx, err := c.sendMsg(ctx, msg)
			msgResChan <- message.Response{Id: msg.Id(), Tx: tx, Err: err}
		}
	}()

	return msgResChan
}

func (c *Client) CallMsg(ctx context.Context, msg message.Request, blockNumber *big.Int) (returnData []byte, err error) {
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

func (c *Client) SafeSendMsg(ctx context.Context, msg message.Request) (*types.Transaction, []byte, error) {
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

func (c *Client) SendMsg(ctx context.Context, msg message.Request) (signedTx *types.Transaction, err error) {
	err = c.msgStore.AddMsg(msg)
	if err != nil {
		return nil, err
	}

	return c.sendMsg(ctx, msg)
}

func (c *Client) sendMsg(ctx context.Context, msg message.Request) (signedTx *types.Transaction, err error) {
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

	err = c.msgStore.UpdateMsgStatus(msg.Id(), message.MessageStatusNonceAssigned)
	if err != nil {
		return nil, err
	}

	// chainID, err := c.Client.ChainID(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("get Chain ID err: %v", err)
	// }

	// signedTx, err = types.SignTx(tx, types.NewEIP2930Signer(chainID), msg.PrivateKey)
	// if err != nil {
	// 	return nil, fmt.Errorf("SignTx err: %v", err)
	// }

	signerFn := c.GetSigner()
	signedTx, err = signerFn(msg.From, tx)
	if err != nil {
		return nil, err
	}

	err = c.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("SendTransaction err: %v", err)
	}

	err = c.msgStore.UpdateMsgStatus(msg.Id(), message.MessageStatusInflight)
	if err != nil {
		return nil, err
	}

	resp := message.Response{
		Id:  msg.Id(),
		Tx:  signedTx,
		Err: err,
	}

	err = c.msgStore.UpdateResponse(msg.Id(), resp)
	if err != nil {
		log.Debug("update msg response failed", "err", err, "msgId", msg.Id().Hex())
		return nil, err
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
func (c *Client) MessageToTransactOpts(ctx context.Context, msg message.Request) (*bind.TransactOpts, error) {
	nonce, err := c.PendingNonceAt(ctx, msg.From)
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

func (c *Client) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	ret, err := c.Client.PendingCallContract(ctx, msg)
	if err != nil {
		return nil, c.DecodeJsonRpcError(err)
	}

	return ret, nil
}

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
