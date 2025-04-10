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
	"github.com/ivanzzeth/ethclient/gnosis_safe/contract/v1.3/compatibilityFallbackHandler"
	"github.com/ivanzzeth/ethclient/gnosis_safe/contract/v1.3/safel2contract"

	safeproxycontract "github.com/ivanzzeth/ethclient/gnosis_safe/contract/v1.3/safeproxy"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/simulated"
)

var (
	PrivateKey1, _ = crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	Addr1          = crypto.PubkeyToAddress(PrivateKey1.PublicKey)
	PrivateKey2, _ = crypto.HexToECDSA("59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d")
	Addr2          = crypto.PubkeyToAddress(PrivateKey2.PublicKey)
	PrivateKey3, _ = crypto.HexToECDSA("5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a")
	Addr3          = crypto.PubkeyToAddress(PrivateKey3.PublicKey)
	PrivateKey4, _ = crypto.HexToECDSA("7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6")
	Addr4          = crypto.PubkeyToAddress(PrivateKey4.PublicKey)

	AddrZero = common.HexToAddress("0x0000000000000000000000000000000000000000")
)

func SetUpClient(t *testing.T) (sim *simulated.Backend) {
	// For more details about logs
	handler := log.NewTerminalHandlerWithLevel(os.Stdout, log.LevelDebug, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	alloc := types.GenesisAlloc{
		Addr1: types.Account{
			Balance: big.NewInt(0).Mul(big.NewInt(100000000), big.NewInt(1000000000000000000)),
		},
		Addr2: types.Account{
			Balance: big.NewInt(0).Mul(big.NewInt(100000000), big.NewInt(1000000000000000000)),
		},
		Addr3: types.Account{
			Balance: big.NewInt(0).Mul(big.NewInt(100000000), big.NewInt(1000000000000000000)),
		},
		Addr4: types.Account{
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

	err = client.RegisterPrivateKey(context.Background(), PrivateKey1)
	if err != nil {
		t.Fatal(err)
	}

	client.AddABI(contracts.GetTestContractABI())

	t.Logf("Setup Client successful")
	return
}

func DeployTestContract(t *testing.T, ctx context.Context, backend *simulated.Backend) (common.Address, *types.Transaction, *contracts.Contracts) {
	auth, err := backend.Client().MessageToTransactOpts(ctx, message.Request{From: Addr1})
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

func DeploySafeContract(t *testing.T, backend *simulated.Backend) (common.Address, *safel2contract.Safel2contract) {
	auth, err := backend.Client().MessageToTransactOpts(context.Background(), message.Request{From: Addr1})
	if err != nil {
		t.Fatal(err)
	}

	fallbackHandlerAddr, txOfContractCreation, _, err := compatibilityFallbackHandler.DeployCompatibilityFallbackHandler(auth, backend.Client())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", fallbackHandlerAddr.Hex())

	backend.CommitAndExpectTx(txOfContractCreation.Hash())

	_, contains := backend.Client().WaitTxReceipt(txOfContractCreation.Hash(), 0, 5*time.Second)
	if !contains {
		t.Fatal("deploy fallbackHandler contract failed")
	}

	auth, err = backend.Client().MessageToTransactOpts(context.Background(), message.Request{From: Addr1})
	if err != nil {
		t.Fatal(err)
	}

	singletonAddr, txOfContractCreation, _, err := safel2contract.DeploySafel2contract(auth, backend.Client())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", singletonAddr.Hex())

	backend.CommitAndExpectTx(txOfContractCreation.Hash())

	_, contains = backend.Client().WaitTxReceipt(txOfContractCreation.Hash(), 0, 5*time.Second)
	if !contains {
		t.Fatal("deploy singleton contract failed")
	}

	auth, err = backend.Client().MessageToTransactOpts(context.Background(), message.Request{From: Addr1})
	if err != nil {
		t.Fatal(err)
	}

	safeAddr, txOfContractCreation, _, err := safeproxycontract.DeploySafeproxycontract(auth, backend.Client(), singletonAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", singletonAddr.Hex())

	backend.CommitAndExpectTx(txOfContractCreation.Hash())

	_, contains = backend.Client().WaitTxReceipt(txOfContractCreation.Hash(), 0, 5*time.Second)
	if !contains {
		t.Fatal("deploy safeProxy contract failed")
	}

	auth, err = backend.Client().MessageToTransactOpts(context.Background(), message.Request{From: Addr1})
	if err != nil {
		t.Fatal(err)
	}

	safeContract, err := safel2contract.NewSafel2contract(safeAddr, backend.Client())
	if err != nil {
		t.Fatal(err)
	}

	owners := []common.Address{Addr1, Addr2, Addr3, Addr4}
	threshold := big.NewInt(3)
	txOfSetup, err := safeContract.Setup(auth, owners, threshold, AddrZero, []byte{}, fallbackHandlerAddr, AddrZero, big.NewInt(0), AddrZero)
	if err != nil {
		t.Fatal(err)
	}
	backend.CommitAndExpectTx(txOfSetup.Hash())

	_, contains = backend.Client().WaitTxReceipt(txOfSetup.Hash(), 0, 5*time.Second)
	if !contains {
		t.Fatal("Setup safeProxy failed")
	}

	return safeAddr, safeContract
}
