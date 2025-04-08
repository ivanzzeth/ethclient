package gnosissafe

import (
	"errors"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/ethclient/nonce"
)

var _ SafeTxBuilder = &SafeTxBuilderByContract{}

type SafeTxBuilder interface {
	Build(safeTxParams ...any) (callData []byte, signatures []byte, nonce uint64, err error)
	GetContractAddress() (common.Address, error)
}

type SafeTxBuilderByContract struct {
	safeContract    SafeContract
	addressToSigner map[common.Address]Signer
	sortAddresses   []common.Address
	nonceStorage    nonce.Storage
}

func NewSafeTxBuilderByContract(safe SafeContract, signers map[common.Address]Signer, nonceStorage nonce.Storage) (SafeTxBuilder, error) {
	threshold, err := safe.GetThreshold()
	if err != nil {
		return nil, err
	}

	if threshold > uint64(len(signers)) {
		return nil, errors.New("safe contract threshold > len(signers)")
	}

	nonce, err := safe.GetNonce()
	if err != nil {
		return nil, err
	}

	safeAddress := safe.GetAddress()

	locker := nonceStorage.NonceLockFrom(safeAddress)
	locker.Lock()
	defer locker.Unlock()

	nonceStorage.SetNonce(safeAddress, nonce)

	sortAddresses := make([]common.Address, 0, len(signers))
	for addr := range signers {
		sortAddresses = append(sortAddresses, addr)
	}

	sort.Slice(sortAddresses, func(i, j int) bool {
		return sortAddresses[i].Cmp(sortAddresses[j]) < 0
	})

	return &SafeTxBuilderByContract{
		safeContract:    safe,
		addressToSigner: signers,
		sortAddresses:   sortAddresses,
		nonceStorage:    nonceStorage,
	}, nil
}

func (builder *SafeTxBuilderByContract) Build(safeTxParams ...any) ([]byte, []byte, uint64, error) {

	locker := builder.nonceStorage.NonceLockFrom(builder.safeContract.GetAddress())
	locker.Lock()
	defer locker.Unlock()

	safeNonce, err := builder.nonceStorage.GetNonce(builder.safeContract.GetAddress())
	if err != nil {
		return nil, nil, 0, err
	}

	safeTxHash, err := builder.safeContract.GetTransactionHash(safeNonce, safeTxParams...)
	if err != nil {
		return nil, nil, 0, err
	}

	signatures := make([]byte, 0)
	for _, address := range builder.sortAddresses {

		signer, ok := builder.addressToSigner[address]
		if !ok {
			return nil, nil, 0, errors.New("unknown signer address")
		}

		signerFn := signer.GetSignerFn()
		signature, err := signerFn(safeTxHash, address)
		if err != nil {
			return nil, nil, 0, err
		}
		signatures = append(signatures, signature...)
	}

	callData, err := builder.safeContract.EncodeExecTransactionData(signatures, safeTxParams...)
	if err != nil {
		return nil, nil, 0, err
	}

	err = builder.nonceStorage.SetNonce(builder.safeContract.GetAddress(), safeNonce+1)
	if err != nil {
		return nil, nil, 0, err
	}

	return callData, signatures, safeNonce, nil
}

func (builder *SafeTxBuilderByContract) GetContractAddress() (common.Address, error) {
	return builder.safeContract.GetAddress(), nil
}
