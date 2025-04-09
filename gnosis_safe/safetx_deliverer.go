package gnosissafe

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethclient/message"
)

var _ SafeTxDeliverer = &SafeTxDelivererByEthClient{}

type SafeTxDeliverer interface {
	Deliverer(req *message.Request, safeNonce uint64) error
}

type SafeTxDelivererByEthClient struct {
	ethClient        *ethclient.Client
	clientSendTxAddr common.Address
}

func NewSafeTxDelivererByEthClient(ethClient *ethclient.Client, clientSendTxAddr common.Address) *SafeTxDelivererByEthClient {
	return &SafeTxDelivererByEthClient{
		ethClient:        ethClient,
		clientSendTxAddr: clientSendTxAddr,
	}
}

func (deliverer *SafeTxDelivererByEthClient) Deliverer(req *message.Request, safeNonce uint64) error {

	if req.From != deliverer.clientSendTxAddr {
		return errors.New("from address do not match")
	}

	safeContract, err := NewDefaultSafelContractCallerByAddress(*req.To, deliverer.ethClient.Client)
	if err != nil {
		return err
	}

	nonceInChain, err := safeContract.GetNonce()
	if err != nil {
		return err
	}

	if nonceInChain < safeNonce {
		req.AfterMsg = message.GenerateMessageIdByAddressAndNonce(*req.To, int64(safeNonce-1))
	} else if nonceInChain > safeNonce {
		return errors.New("safeNonce is invalid")
	}

	// sync schedule
	deliverer.ethClient.ScheduleMsg(req)

	return nil
}
