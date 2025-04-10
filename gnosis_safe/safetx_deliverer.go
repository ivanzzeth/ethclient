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
	ethClient                         *ethclient.Client
	clientSendTxAddr                  common.Address
	addrToCaller                      map[common.Address]SafelContractCaller
	defaultSafelContractCallerCreator SafelContractCallerCreator
}

type Option func(*SafeTxDelivererByEthClient)

func SetDefaultSafelContractCallerCreator(deliverer *SafeTxDelivererByEthClient) {
	deliverer.defaultSafelContractCallerCreator = NewDefaultSafelContractCallerCreator
}

func NewSafeTxDelivererByEthClient(ethClient *ethclient.Client, clientSendTxAddr common.Address, options []Option) SafeTxDeliverer {
	out := &SafeTxDelivererByEthClient{
		ethClient:                         ethClient,
		clientSendTxAddr:                  clientSendTxAddr,
		addrToCaller:                      make(map[common.Address]SafelContractCaller),
		defaultSafelContractCallerCreator: NewDefaultSafelContractCallerCreator,
	}

	for _, option := range options {
		option(out)
	}
	return out
}

func (deliverer *SafeTxDelivererByEthClient) Deliverer(req *message.Request, safeNonce uint64) (err error) {

	if req.From != deliverer.clientSendTxAddr {
		return errors.New("from address do not match")
	}

	safelContractCaller, ok := deliverer.addrToCaller[*req.To]
	if !ok {
		safelContractCaller, err = deliverer.defaultSafelContractCallerCreator(*req.To, deliverer.ethClient.Client)
		if err != nil {
			return err
		}
		deliverer.addrToCaller[*req.To] = safelContractCaller
	}

	nonceInChain, err := safelContractCaller.GetNonce()
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
