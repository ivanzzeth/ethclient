package helper

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ivanzzeth/ethclient/contracts"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/simulated"
)

var (
	PrivateKey, _ = crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	Addr          = crypto.PubkeyToAddress(PrivateKey.PublicKey)
)

func SetUpClient(t *testing.T) (sim *simulated.Backend) {
	// For more details about logs
	handler := log.NewTerminalHandlerWithLevel(os.Stdout, log.LevelDebug, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	alloc := types.GenesisAlloc{
		Addr: types.Account{
			Balance: big.NewInt(0).Mul(big.NewInt(100000000), big.NewInt(1000000000000000000)),
		},
	}
	sim = simulated.NewBackend(alloc)

	client := sim.Client()
	// client, err := ethclient.Dial("http://localhost:8545")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if latestBlock != 0 {
		panic(fmt.Errorf("unexpected block number: %v", latestBlock))
	}

	err = client.RegisterPrivateKey(context.Background(), PrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	client.AddABI(contracts.GetTestContractABI())

	t.Logf("Setup Client successful")
	return
}

func DeployTestContract(t *testing.T, ctx context.Context, backend *simulated.Backend) (common.Address, *types.Transaction, *contracts.Contracts) {
	auth, err := backend.Client().MessageToTransactOpts(ctx, message.Request{From: Addr})
	if err != nil {
		t.Fatal(err)
	}

	contractAddr, txOfContractCreation, contract, err := contracts.DeployContracts(auth, backend.Client())
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	backend.CommitAndExpectTx(txOfContractCreation.Hash())

	_, contains := backend.Client().WaitTxReceipt(txOfContractCreation.Hash(), 0, 5*time.Second)
	if !contains {
		t.Fatal("deploy contract failed")
	}

	return contractAddr, txOfContractCreation, contract
}
