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
)

type Client struct {
	rawClient *ethclient.Client
	rpcClient *rpc.Client
	nm        *NonceManager
	Subscriber
}

func Dial(rawurl string) (*Client, error) {
	rpcClient, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	c := ethclient.NewClient(rpcClient)

	nm, err := NewNonceManager(c)
	if err != nil {
		return nil, err
	}

	subscriber, err := NewChainSubscriber(c)
	if err != nil {
		return nil, err
	}

	return &Client{
		rawClient:  c,
		rpcClient:  rpcClient,
		nm:         nm,
		Subscriber: subscriber,
	}, nil
}

func NewClient(c *rpc.Client) (*Client, error) {
	ethc := ethclient.NewClient(c)

	nm, err := NewNonceManager(ethc)
	if err != nil {
		return nil, err
	}

	subscriber, err := NewChainSubscriber(ethc)
	if err != nil {
		return nil, err
	}

	return &Client{
		rawClient:  ethc,
		rpcClient:  c,
		nm:         nm,
		Subscriber: subscriber,
	}, nil
}

func (c *Client) Close() {
	c.rawClient.Close()
}

// RawClient returns ethclient
func (c *Client) RawClient() *ethclient.Client {
	return c.rawClient
}

type Message struct {
	From       common.Address    // the sender of the 'transaction'
	PrivateKey *ecdsa.PrivateKey // overwrite From if not nil
	To         *common.Address   // the destination contract (nil for contract creation)
	Gas        uint64            // if 0, the call executes with near-infinite gas
	GasPrice   *big.Int          // wei <-> gas exchange ratio
	Value      *big.Int          // amount of wei sent along with the call
	Data       []byte            // input data, usually an ABI-encoded contract method invocation

	AccessList types.AccessList // EIP-2930 access list.
}

func (c *Client) NewMethodData(a abi.ABI, methodName string, args ...interface{}) ([]byte, error) {
	return a.Pack(methodName, args...)
}

func (c *Client) BatchSendMsg(ctx context.Context, msgs <-chan Message) (<-chan *types.Transaction, <-chan error) {
	txs := make(chan *types.Transaction, 10)
	errs := make(chan error, 10)
	go func() {
		for msg := range msgs {
			tx, err := c.SendMsg(ctx, msg)
			txs <- tx
			errs <- err
		}

		close(txs)
		close(errs)
	}()
	return txs, errs
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

	return c.rawClient.CallContract(ctx, ethMesg, blockNumber)
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

func (c *Client) SendMsg(ctx context.Context, msg Message) (*types.Transaction, error) {
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

	chainID, err := c.rawClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("Get Chain ID err: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP2930Signer(chainID), msg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("SignTx err: %v", err)
	}

	err = c.rawClient.SendTransaction(ctx, signedTx)
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
		gas, err := c.rawClient.EstimateGas(ctx, msg)
		if err != nil {
			return nil, err
		}

		msg.Gas = gas
	}

	if msg.GasPrice == nil || msg.GasPrice.Uint64() == 0 {
		var err error
		msg.GasPrice, err = c.rawClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
	}

	nonce, err := c.nm.PendingNonceAt(ctx, msg.From)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(nonce, *msg.To, msg.Value, msg.Gas, msg.GasPrice, msg.Data)

	return tx, nil
}

func (c *Client) WaitTxReceipt(txHash common.Hash, confirmations uint64, timeout time.Duration) (*types.Receipt, bool) {
	startTime := time.Now()
	for {
		currTime := time.Now()
		elapsedTime := currTime.Sub(startTime)
		if elapsedTime >= timeout {
			return nil, false
		}

		receipt, err := c.rawClient.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			continue
		}

		block, err := c.rawClient.BlockNumber(context.Background())
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

	nonce, err := c.nm.PendingNonceAt(ctx, msg.From)
	if err != nil {
		return nil, err
	}

	chainID, err := c.rawClient.ChainID(ctx)
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
