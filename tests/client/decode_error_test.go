package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/ivanzzeth/ethclient/contracts"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func Test_DecodeJsonRpcError(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	client.AddABI(contracts.GetTestContractABI())

	ctx := context.Background()
	// Deploy Test contract.
	contractAddr, txOfContractCreation, _ := helper.DeployTestContract(t, ctx, client)

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	assert.Equal(t, true, contains)

	testContract, err := contracts.NewContracts(contractAddr, client)
	if err != nil {
		t.Fatal(err)
	}

	err = testContract.TestReverted(nil, true)
	t.Log("TestReverted err: ", err)

	err = testContract.TestRevertedString(nil, true)
	t.Log("TestRevertedString err: ", err)
}
