package ethclient

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
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
	log.Debug("DialOnce...", "url", url)

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
			break
		}
	})

	backend, _ := clientMap.Load(url)

	log.Debug("DialOnce successful...", "url", url)

	return backend.(EthClientInterface)
}

func Dial(rawurl string) (*Client, error) {
	rpcClient, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	client := NewClient(rpcClient)
	return client, nil
}

// DialWithTransport creates an RPC client that uses the given transport for HTTP/HTTPS endpoints.
// When transport is non-nil and rawurl has scheme "http" or "https", rpc.DialHTTPWithClient is used.
// Otherwise Dial(rawurl) is used (e.g. for ws/wss or when transport is nil).
func DialWithTransport(rawurl string, transport *http.Transport) (*Client, error) {
	if transport != nil {
		u, err := url.Parse(rawurl)
		if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
			httpClient := &http.Client{Transport: transport}
			rpcClient, err := rpc.DialHTTPWithClient(rawurl, httpClient)
			if err != nil {
				return nil, err
			}
			return NewClient(rpcClient), nil
		}
	}
	return Dial(rawurl)
}
