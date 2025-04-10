package gnosissafe

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func TestSafeContractVersion1_3_0(t *testing.T) {
	sim := helper.SetUpClient(t)

	safeAddr, safeContract := helper.DeploySafeContract(t, sim)

	safeContractV1_3, err := NewSafeContractVersion1_3_0(safeAddr, sim.Client())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, safeAddr, safeContractV1_3.GetAddress())

	wantThreshold, err := safeContract.GetThreshold(nil)
	if err != nil {
		t.Fatal(err)
	}

	getThreshold, err := safeContractV1_3.GetThreshold()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wantThreshold.Uint64(), getThreshold)

	wantOwners := []common.Address{helper.Addr1, helper.Addr2, helper.Addr3, helper.Addr4}
	if err != nil {
		t.Fatal(err)
	}

	getOwners, err := safeContractV1_3.GetOwners()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wantOwners, getOwners)
}
