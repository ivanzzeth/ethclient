package ethclient

import (
	"github.com/ethereum/go-ethereum/rpc"

	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
)

type EthClientInterface interface {
	bind.ContractBackend
	bind.DeployBackend
	bind.PendingContractCaller
	ethereum.ChainReader
	ethereum.TransactionReader
	ethereum.ChainStateReader
	ethereum.ContractCaller
	ethereum.LogFilterer
	ethereum.TransactionSender
	ethereum.GasPricer
	ethereum.GasPricer1559
	ethereum.FeeHistoryReader
	ethereum.PendingStateReader
	ethereum.PendingContractCaller
	ethereum.GasEstimator
	ethereum.BlockNumberReader
	ethereum.ChainIDReader
}

type Dialer = func(rawurl string) (EthClientInterface, error)

var dialer Dialer = func(rawurl string) (EthClientInterface, error) {
	return Dial(rawurl)
}

func SetDialer(d Dialer) {
	dialer = d
}

var dialerOnceMap sync.Map
var clientMap sync.Map

func DialOnce(url string) EthClientInterface {
	o, _ := dialerOnceMap.LoadOrStore(url, &sync.Once{})
	once := o.(*sync.Once)
	once.Do(func() {
		for {
			backend, err := dialer(url)
			if err != nil {
				log.Error("Dial backend failed", "err", err, "url", url)
				time.Sleep(3 * time.Second)
				continue
			}

			clientMap.Store(url, backend)
		}
	})

	backend, _ := clientMap.Load(url)
	return backend.(EthClientInterface)
}

func Dial(rawurl string) (*Client, error) {
	rpcClient, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	return NewMemoryClient(rpcClient)
}
