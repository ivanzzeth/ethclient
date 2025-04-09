package gnosissafe

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var _ Signer = &PrivateKeySigner{}

type SafeSignFn func(hash []byte, address common.Address) ([]byte, error)

type Signer interface {
	GetSignerFn() SafeSignFn
	RegisterSignerFn(signerFn SafeSignFn)
}

type PrivateKeySigner struct {
	Address     common.Address
	safeSignFns []SafeSignFn
}

func NewPrivateKeySigner(key *ecdsa.PrivateKey) *PrivateKeySigner {
	signer := &PrivateKeySigner{
		safeSignFns: make([]SafeSignFn, 0),
	}

	signer.Address = crypto.PubkeyToAddress(key.PublicKey)

	signerFn := func(hash []byte, address common.Address) ([]byte, error) {
		if address != signer.Address {
			return nil, bind.ErrNotAuthorized
		}

		signature, err := crypto.Sign(hash, key)
		if err != nil {
			return nil, err
		}

		// EIP-155 compatible
		signature[64] += 27

		return signature, nil
	}

	signer.RegisterSignerFn(signerFn)

	return signer
}

func (signer *PrivateKeySigner) GetSignerFn() SafeSignFn {
	return func(hash []byte, address common.Address) ([]byte, error) {
		if len(signer.safeSignFns) == 0 {
			return nil, fmt.Errorf("no signerFn registered")
		}

		for _, fn := range signer.safeSignFns {

			signature, err := fn(hash, address)
			if err != nil {
				continue
			}

			return signature, nil
		}

		return nil, bind.ErrNotAuthorized
	}
}

func (signer *PrivateKeySigner) RegisterSignerFn(signerFn SafeSignFn) {
	signer.safeSignFns = append(signer.safeSignFns, signerFn)
}
