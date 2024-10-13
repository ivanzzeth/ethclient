package client_test

import (
	"context"
	"testing"

	"github.com/ivanzzeth/ethclient/contracts"
	"github.com/ivanzzeth/ethclient/tests/helper"
)

func Test_DecodeJsonRpcError(t *testing.T) {
	sim := helper.SetUpClient(t)
	defer sim.Close()

	client := sim.Client()
	client.AddABI(contracts.GetTestContractABI())

	ctx := context.Background()
	// Deploy Test contract.
	contractAddr, txOfContractCreation, _ := helper.DeployTestContract(t, ctx, sim)

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	testContract, err := contracts.NewContracts(contractAddr, client)
	if err != nil {
		t.Fatal(err)
	}

	err = testContract.TestReverted(nil, true)
	t.Log("TestReverted err: ", err)

	err = testContract.TestRevertedString(nil, true)
	t.Log("TestRevertedString err: ", err)
}
