package gnosissafe

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethclient/message"
)

var _ SafeTxDeliverr = &SafeTxDeliverrByEthClient{}

type SafeTxDeliverr interface {
	Deliverr(req *message.Request, safeNonce uint64) error
}

type SafeTxDeliverrByEthClient struct {
	ethClient        *ethclient.Client
	clientSendTxAddr common.Address
}

func NewSafeTxDeliverrByEthClient(ethClient *ethclient.Client, clientSendTxAddr common.Address) *SafeTxDeliverrByEthClient {
	return &SafeTxDeliverrByEthClient{
		ethClient:        ethClient,
		clientSendTxAddr: clientSendTxAddr,
	}
}

func (deliverr *SafeTxDeliverrByEthClient) Deliverr(req *message.Request, safeNonce uint64) error {

	if req.From != deliverr.clientSendTxAddr {
		return errors.New("from address do not match")
	}

	safeContract, err := NewDefaultSafelContractCallerByAddress(*req.To, deliverr.ethClient.Client)
	if err != nil {
		return err
	}

	nonceInChain, err := safeContract.GetNonce()
	if err != nil {
		return err
	}
	if nonceInChain == safeNonce {
		deliverr.ethClient.ScheduleMsg(req)
	} else if nonceInChain < safeNonce {
		req.AfterMsg = message.GenerateMessageIdByAddressAndNonce(*req.To, int64(safeNonce-1))
	} else {
		return errors.New("safeNonce is invalid")
	}

	// sync schedule
	deliverr.ethClient.ScheduleMsg(req)

	return nil
}
