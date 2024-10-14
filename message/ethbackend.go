package message

import "github.com/ethereum/go-ethereum"

type ethBackend interface {
	ethereum.ContractCaller
	ethereum.BlockNumberReader
	ethereum.TransactionSender
	ethereum.TransactionReader
	ethereum.PendingStateReader
	ethereum.GasPricer
	ethereum.GasEstimator
}
