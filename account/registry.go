package account

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type Registry interface {
	GetSigner() bind.SignerFn
	RegisterSigner(signerFn bind.SignerFn)
	RegisterPrivateKey(ctx context.Context, key *ecdsa.PrivateKey) error
}
