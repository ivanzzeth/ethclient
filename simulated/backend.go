//// This file forked from go-ethereum repo and do some tiny changes.

// Copyright 2023 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package simulated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/catalyst"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ivanzzeth/ethclient"
)

// Backend is a simulated blockchain. You can use it to test your contracts or
// other code that interacts with the Ethereum chain.
type Backend struct {
	node   *node.Node
	beacon *catalyst.SimulatedBeacon
	client *ethclient.Client
}

type MineOption struct {
	Miner    common.Address
	Duration time.Duration
}

func NewAutoMineBackend(alloc types.GenesisAlloc, mineOption MineOption) *Backend {
	sim := NewBackend(alloc, func(nodeConf *node.Config, ethConf *ethconfig.Config) {})

	chainID, err := sim.Client().ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	pendingTxs := make(chan *types.Transaction)
	_, err = sim.Client().SubscribeFullPendingTransactions(context.Background(), pendingTxs)
	if err != nil {
		panic(err)
	}

	go func() {
		for tx := range pendingTxs {
			from, err := types.Sender(types.LatestSignerForChainID(chainID), tx)
			if err != nil {
				panic(err)
			}
			log.Info("incoming transaction", "txHash", tx.Hash().Hex(), "from", from, "nonce", tx.Nonce())
		}
	}()

	// if mineOption.Duration != 0 {
	// 	go func() {
	// 		ticker := time.NewTicker(mineOption.Duration)

	// 		for range ticker.C {
	// 			blockHash := sim.Commit()
	// 			block, err := sim.Client().BlockByHash(context.Background(), blockHash)
	// 			if err != nil {
	// 				log.Error("Simulated backend call BlockByHash failed", "err", err)
	// 				continue
	// 			}

	// 			lastBlock, _ := sim.Client().BlockNumber(context.Background())

	// 			log.Info("mined block", "blockNumber", block.Number().String(), "blockHash", blockHash.Hex(), "latestBlock", lastBlock, "txs", len(block.Transactions()))
	// 			for _, tx := range block.Transactions() {
	// 				log.Info("mined tx", "blockNumber", block.Number(), "txHash", tx.Hash().Hex())
	// 			}
	// 		}
	// 	}()
	// }

	return sim
}

// NewBackend creates a new simulated blockchain that can be used as a backend for
// contract bindings in unit tests.
//
// A simulated backend always uses chainID 1337.
func NewBackend(alloc types.GenesisAlloc, options ...func(nodeConf *node.Config, ethConf *ethconfig.Config)) *Backend {
	// Create the default configurations for the outer node shell and the Ethereum
	// service to mutate with the options afterwards
	nodeConf := node.DefaultConfig
	nodeConf.DataDir = ""
	nodeConf.P2P = p2p.Config{NoDiscovery: true}

	ethConf := ethconfig.Defaults
	ethConf.Genesis = &core.Genesis{
		Config:   params.AllDevChainProtocolChanges,
		GasLimit: ethconfig.Defaults.Miner.GasCeil,
		Alloc:    alloc,
	}
	ethConf.SyncMode = downloader.FullSync
	ethConf.TxPool.NoLocals = true

	for _, option := range options {
		option(&nodeConf, &ethConf)
	}
	// Assemble the Ethereum stack to run the chain with
	stack, err := node.New(&nodeConf)
	if err != nil {
		panic(err) // this should never happen
	}
	sim, err := newWithNode(stack, &ethConf, 0)
	if err != nil {
		panic(err) // this should never happen
	}
	return sim
}

func NewBackendFromConfig(conf ethconfig.Config) *Backend {
	// Setup the node object
	nodeConf := node.DefaultConfig
	nodeConf.DataDir = ""
	nodeConf.P2P = p2p.Config{NoDiscovery: true}
	stack, err := node.New(&nodeConf)
	if err != nil {
		// This should never happen, if it does, please open an issue
		panic(err)
	}

	conf.SyncMode = downloader.FullSync
	conf.TxPool.NoLocals = true
	sim, err := newWithNode(stack, &conf, 0)
	if err != nil {
		// This should never happen, if it does, please open an issue
		panic(err)
	}
	return sim
}

// newWithNode sets up a simulated backend on an existing node. The provided node
// must not be started and will be started by this method.
func newWithNode(stack *node.Node, conf *eth.Config, blockPeriod uint64) (*Backend, error) {
	backend, err := eth.New(stack, conf)
	if err != nil {
		return nil, err
	}
	// Register the filter system
	filterSystem := filters.NewFilterSystem(backend.APIBackend, filters.Config{})
	stack.RegisterAPIs([]rpc.API{{
		Namespace: "eth",
		Service:   filters.NewFilterAPI(filterSystem),
	}})
	// Start the node
	if err := stack.Start(); err != nil {
		return nil, err
	}
	// Set up the simulated beacon
	beacon, err := catalyst.NewSimulatedBeacon(blockPeriod, backend)
	if err != nil {
		return nil, err
	}
	// Reorg our chain back to genesis
	if err := beacon.Fork(backend.BlockChain().GetCanonicalHash(0)); err != nil {
		return nil, err
	}
	return &Backend{
		node:   stack,
		beacon: beacon,
		client: ethclient.NewClient(stack.Attach()),
	}, nil
}

// Close shuts down the simBackend.
// The simulated backend can't be used afterwards.
func (n *Backend) Close() error {
	if n.client.Client != nil {
		n.client.Close()
		n.client = &ethclient.Client{}
	}
	var err error
	if n.beacon != nil {
		err = n.beacon.Stop()
		n.beacon = nil
	}
	if n.node != nil {
		err = errors.Join(err, n.node.Close())
		n.node = nil
	}
	return err
}

// Commit seals a block and moves the chain forward to a new empty block.
func (n *Backend) Commit() common.Hash {
	blockHash := n.beacon.Commit()

	n.expectTxAtBlock(common.Hash{}, blockHash)
	return blockHash
}

func (n *Backend) CommitAndExpectTx(txHash common.Hash) common.Hash {
	blockHash := n.beacon.Commit()

	if !n.expectTxAtBlock(txHash, blockHash) {
		panic(fmt.Errorf("committed block not including txHash: %v", txHash.Hex()))
	}

	return blockHash
}

func (n *Backend) expectTxAtBlock(txHash, blockHash common.Hash) bool {
	block, err := n.client.BlockByHash(context.Background(), blockHash)
	if err != nil {
		log.Error("Simulated backend call BlockByHash failed", "err", err)
		return false
	} else {
		lastBlock, _ := n.client.BlockNumber(context.Background())

		log.Info("mined block", "blockNumber", block.Number().String(), "blockHash", blockHash.Hex(), "latestBlock", lastBlock, "txs", len(block.Transactions()))
		for _, tx := range block.Transactions() {
			log.Info("mined tx", "blockNumber", block.Number(), "txHash", tx.Hash().Hex())
			if txHash.Cmp(common.Hash{}) != 0 && tx.Hash().Cmp(txHash) == 0 {
				return true
			}
		}
	}

	return false
}

// Rollback removes all pending transactions, reverting to the last committed state.
func (n *Backend) Rollback() {
	n.beacon.Rollback()
}

// Fork creates a side-chain that can be used to simulate reorgs.
//
// This function should be called with the ancestor block where the new side
// chain should be started. Transactions (old and new) can then be applied on
// top and Commit-ed.
//
// Note, the side-chain will only become canonical (and trigger the events) when
// it becomes longer. Until then CallContract will still operate on the current
// canonical chain.
//
// There is a % chance that the side chain becomes canonical at the same length
// to simulate live network behavior.
func (n *Backend) Fork(parentHash common.Hash) error {
	return n.beacon.Fork(parentHash)
}

// AdjustTime changes the block timestamp and creates a new block.
// It can only be called on empty blocks.
func (n *Backend) AdjustTime(adjustment time.Duration) error {
	return n.beacon.AdjustTime(adjustment)
}

// Client returns a client that accesses the simulated chain.
func (n *Backend) Client() *ethclient.Client {
	return n.client
}
