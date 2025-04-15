package gnosissafe

import (
	"errors"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/ethclient/nonce"
)

var _ SafeTxBuilder = &SafeTxBuilderByContract{}

// SafeTxBuilder builds transaction calldata and maintains Safe nonce sequence.
type SafeTxBuilder interface {
	// Build return transaction calldata, verifiable signatures, and safe nonce.
	Build(safeTxParams SafeTxParam) (callData []byte, signatures []byte, nonce uint64, err error)
	BuildCustomNonce(safeTxParams SafeTxParam, nonce uint64) (callData []byte, signatures []byte, err error)
	GetContractAddress() (common.Address, error)
}

type SafeTxBuilderByContract struct {
	safeContract    SafeContract
	addressToSigner map[common.Address]Signer
	sortedAddresses []common.Address
	nonceStorage    nonce.Storage
}

func NewSafeTxBuilderByContract(safe SafeContract, signers map[common.Address]Signer, nonceStorage nonce.Storage) (*SafeTxBuilderByContract, error) {
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
		sortedAddresses: sortAddresses,
		nonceStorage:    nonceStorage,
	}, nil
}

func (builder *SafeTxBuilderByContract) Build(safeTxParams SafeTxParam) ([]byte, []byte, uint64, error) {

	safeContractAddr := builder.safeContract.GetAddress()

	locker := builder.nonceStorage.NonceLockFrom(safeContractAddr)
	locker.Lock()
	defer locker.Unlock()

	safeNonce, err := builder.nonceStorage.GetNonce(safeContractAddr)
	if err != nil {
		return nil, nil, 0, err
	}

	callData, signatures, err := builder.buildCustomNonce(safeTxParams, safeNonce)
	if err != nil {
		return nil, nil, 0, err
	}

	err = builder.nonceStorage.SetNonce(safeContractAddr, safeNonce+1)
	if err != nil {
		return nil, nil, 0, err
	}

	return callData, signatures, safeNonce, nil
}

func (builder *SafeTxBuilderByContract) BuildCustomNonce(safeTxParams SafeTxParam, safeNonce uint64) (callData []byte, signatures []byte, err error) {
	return builder.buildCustomNonce(safeTxParams, safeNonce)
}

func (builder *SafeTxBuilderByContract) buildCustomNonce(safeTxParams SafeTxParam, nonce uint64) (callData []byte, signatures []byte, err error) {
	contractVersion, err := builder.safeContract.GetVersion()
	if err != nil {
		return nil, nil, err
	}
	if safeTxParams.Version() != contractVersion {
		return nil, nil, ErrSafeParamVersionNotMatch
	}

	safeTxHash, err := builder.safeContract.GetTransactionHash(nonce, safeTxParams)
	if err != nil {
		return nil, nil, err
	}

	signatures = make([]byte, 0)
	for _, address := range builder.sortedAddresses {

		signer, ok := builder.addressToSigner[address]
		if !ok {
			return nil, nil, errors.New("unknown signer address")
		}

		signerFn := signer.GetSignerFn()
		signature, err := signerFn(common.BytesToHash(safeTxHash), address)
		if err != nil {
			return nil, nil, err
		}
		signatures = append(signatures, signature...)
	}

	callData, err = builder.safeContract.EncodeExecTransactionData(signatures, safeTxParams)
	if err != nil {
		return nil, nil, err
	}

	return callData, signatures, nil
}

func (builder *SafeTxBuilderByContract) GetContractAddress() (common.Address, error) {
	return builder.safeContract.GetAddress(), nil
}
