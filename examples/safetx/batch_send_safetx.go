package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethclient"
	gnosissafe "github.com/ivanzzeth/ethclient/gnosis_safe"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/nonce"
)

func main() {

	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	key, _ := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	safeContractAddress := common.HexToAddress("0xc582Bc0317dbb0908203541971a358c44b1F3766")

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Crit(err.Error())
	}
	defer client.Close()

	err = client.RegisterPrivateKey(context.Background(), key)
	if err != nil {
		log.Crit(err.Error())
	}

	safeContract, err := gnosissafe.NewSafeContractVersion1_3_0(safeContractAddress, client.Client)
	if err != nil {
		log.Crit(err.Error())
	}

	safeOwnerKey1, _ := crypto.HexToECDSA("59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d")
	safeOwnerKey2, _ := crypto.HexToECDSA("5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a")
	safeOwnerKey3, _ := crypto.HexToECDSA("7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6")

	safeOwnerKeys := []*ecdsa.PrivateKey{safeOwnerKey1, safeOwnerKey2, safeOwnerKey3}

	signers := make(map[common.Address]gnosissafe.Signer)
	for _, ownerKey := range safeOwnerKeys {
		signers[crypto.PubkeyToAddress(ownerKey.PublicKey)] = gnosissafe.NewPrivateKeySigner(ownerKey)
	}

	nonceStorage := nonce.NewMemoryStorage()
	builder, err := gnosissafe.NewSafeTxBuilderByContract(safeContract, signers, nonceStorage)
	if err != nil {
		log.Crit(err.Error())
	}

	locker := nonceStorage.NonceLockFrom(safeContractAddress)
	locker.Lock()
	startNonce, err := nonceStorage.GetNonce(safeContractAddress)
	if err != nil {
		log.Crit(err.Error())
	}
	locker.Unlock()

	deliverer := gnosissafe.NewSafeTxDelivererByEthClient(client, crypto.PubkeyToAddress(key.PublicKey), nil)

	to := common.HexToAddress("0xa0Ee7A142d267C1f36714E4a8F75612F20a79720")

	var loopCount = 3
	go func() {
		for range loopCount {

			// value is greater than the balance of the contract
			{
				safeTxParam := gnosissafe.SafeTxParamV1_3{
					To:             to,
					Value:          big.NewInt(0).Mul(big.NewInt(1000000000000000000), big.NewInt(10000000000)),
					Calldata:       []byte{},
					Operation:      0,
					SafeTxGas:      big.NewInt(220000),
					BaseGas:        big.NewInt(50000),
					GasPrice:       big.NewInt(2000),
					GasToken:       common.HexToAddress("0x00"),
					RefundReceiver: common.HexToAddress("0x00"),
				}

				callData, _, safeNonce, err := builder.Build(&safeTxParam)
				if err != nil {
					log.Crit(err.Error())
				}
				log.Info("safe nonce value", "int", safeNonce)

				req := message.Request{
					From:     crypto.PubkeyToAddress(key.PublicKey),
					To:       &safeContractAddress,
					Value:    big.NewInt(0),
					Gas:      500000,
					GasPrice: big.NewInt(3000),
					Data:     callData,
				}
				req.SetId(*message.GenerateMessageIdByAddressAndNonce(safeContractAddress, int64(safeNonce)))
				err = deliverer.Deliverer(&req, safeNonce)
				if err != nil {
					log.Crit(err.Error())
				}
			}

			// value is less than the balance of the contract
			{
				safeTxParam := gnosissafe.SafeTxParamV1_3{
					To:             to,
					Value:          big.NewInt(0).Div(big.NewInt(1000000000000000000), big.NewInt(100)),
					Calldata:       []byte{},
					Operation:      0,
					SafeTxGas:      big.NewInt(220000),
					BaseGas:        big.NewInt(50000),
					GasPrice:       big.NewInt(2000),
					GasToken:       common.HexToAddress("0x00"),
					RefundReceiver: common.HexToAddress("0x00"),
				}

				callData, _, safeNonce, err := builder.Build(&safeTxParam)
				if err != nil {
					log.Crit(err.Error())
				}
				log.Info("safe nonce value", "int", safeNonce)

				req := message.Request{
					From:     crypto.PubkeyToAddress(key.PublicKey),
					To:       &safeContractAddress,
					Value:    big.NewInt(0),
					Gas:      500000,
					GasPrice: big.NewInt(3000),
					Data:     callData,
				}
				req.SetId(*message.GenerateMessageIdByAddressAndNonce(safeContractAddress, int64(safeNonce)))
				err = deliverer.Deliverer(&req, safeNonce)
				if err != nil {
					log.Crit(err.Error())
				}
			}

		}
		log.Info("send req done")

		time.Sleep(10 * time.Second)
		client.CloseSendMsg()
	}()

	for resp := range client.Response() {
		fmt.Println("respInfo", "id", resp.Id, "err", resp.Err)
	}

	locker = nonceStorage.NonceLockFrom(safeContractAddress)
	locker.Lock()
	endNonce, err := nonceStorage.GetNonce(safeContractAddress)
	if err != nil {
		log.Crit(err.Error())
	}
	locker.Unlock()

	if endNonce-startNonce != uint64(loopCount*2) {
		log.Crit("safe nonce has mistake")
	}
	log.Info("safe nonce As expected!")
}
