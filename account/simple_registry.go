package account

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

type SimpleRegistry struct {
	chainId *big.Int
	signers []bind.SignerFn // Method to use for signing the transaction (mandatory)
}

func NewSimpleRegistry(chainId *big.Int) *SimpleRegistry {
	return &SimpleRegistry{
		chainId: chainId,
		signers: []bind.SignerFn{},
	}
}

func (r *SimpleRegistry) GetSigner() bind.SignerFn {
	// combine all signerFn
	return func(a common.Address, t *types.Transaction) (tx *types.Transaction, err error) {
		if len(r.signers) == 0 {
			return nil, fmt.Errorf("no signerFn registered")
		}

		for i, fn := range r.signers {
			tx, err = fn(a, t)
			log.Debug("try to call signerFn", "index", i, "err", err, "account", a)

			if err != nil {
				continue
			}

			return tx, nil
		}

		return nil, bind.ErrNotAuthorized
	}
}

func (r *SimpleRegistry) RegisterSigner(signerFn bind.SignerFn) {
	log.Info("register signerFn for signing...")
	r.signers = append(r.signers, signerFn)
}

// Registers the private key used for signing txs.
func (r *SimpleRegistry) RegisterPrivateKey(ctx context.Context, key *ecdsa.PrivateKey) error {
	chainID := r.chainId
	keyAddr := crypto.PubkeyToAddress(key.PublicKey)
	if chainID == nil {
		return bind.ErrNoChainID
	}
	signer := types.LatestSignerForChainID(chainID)
	signerFn := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != keyAddr {
			return nil, bind.ErrNotAuthorized
		}
		signature, err := crypto.Sign(signer.Hash(tx).Bytes(), key)
		if err != nil {
			return nil, err
		}
		return tx.WithSignature(signer, signature)
	}

	r.RegisterSigner(signerFn)

	return nil
}
