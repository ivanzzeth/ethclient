package main

import (
	"context"
	"crypto/ecdsa"
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
	"github.com/ivanzzeth/ethclient/tests/helper"
)

func main() {

	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	// This example is only for proving how it works.
	// TODO: This address should be replaced with your deployed Safe Proxy contract address.
	// The owner addresses of the test contract include helper.Addr1, helper.Addr2, helper.Addr3, and helper.Addr4, with a threshold of 3.
	safeContractAddress := common.HexToAddress("0xc582Bc0317dbb0908203541971a358c44b1F3766")

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Crit(err.Error())
	}
	defer client.Close()

	err = client.RegisterPrivateKey(context.Background(), helper.PrivateKey1)
	if err != nil {
		log.Crit(err.Error())
	}

	safeContract, err := gnosissafe.NewSafeContractVersion1_3_0(safeContractAddress, client.Client)
	if err != nil {
		log.Crit(err.Error())
	}
	startNonceOnChain, err := safeContract.GetNonce()
	if err != nil {
		log.Crit("nonce has err")
	}
	log.Debug("safe nonce in chain for start", "nonce", startNonceOnChain)

	safeOwnerKeys := []*ecdsa.PrivateKey{helper.PrivateKey2, helper.PrivateKey3, helper.PrivateKey4}

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
	locker.Unlock()
	if err != nil {
		log.Crit(err.Error())
	}

	deliverer := gnosissafe.NewSafeTxDelivererByEthClient(client, helper.Addr1)

	var loopCount = 3
	go func() {
		// there's 6=3*2 requests that will be scheduled.
		for range loopCount {
			go func() {

				// value is greater than the balance of the contract.
				/*
					After this tx is executed:
						The transfer will fail (insufficient contract balance)
						The Safe nonce still increments by 1
						Thus, any previously submitted transactions with higher nonces can still execute without causing signature verification failures.
				*/
				{
					safeTxParam := gnosissafe.SafeTxParamV1_3{
						To:             helper.Addr4,
						Value:          big.NewInt(0).Mul(big.NewInt(1000000000000000000), big.NewInt(10000000000)), // 1e10 eth
						Calldata:       []byte{},
						Operation:      0,
						SafeTxGas:      big.NewInt(220000), // TODO: Adjust these gas settings according to your network conditions!
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
						From:  helper.Addr1,
						To:    &safeContractAddress,
						Value: big.NewInt(0),
						Gas:   500000, // TODO: ensure gas meets the minimum requirement for successful Safe contract execution (avoiding reverts).
						Data:  callData,
					}
					err = deliverer.Deliver(&req, safeNonce)
					if err != nil {
						log.Crit(err.Error())
					}
				}

				// value is less than the balance of the contract.
				// this transaction is expected to execute successfully on-chain.
				{
					safeTxParam := gnosissafe.SafeTxParamV1_3{
						To:             helper.Addr4,
						Value:          big.NewInt(0).Div(big.NewInt(1000000000000000000), big.NewInt(100)),
						Calldata:       []byte{},
						Operation:      0,
						SafeTxGas:      big.NewInt(220000), // TODO: Adjust these gas settings according to your network conditions!
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
						From:  helper.Addr1,
						To:    &safeContractAddress,
						Value: big.NewInt(0),
						Gas:   500000, // TODO: ensure gas meets the minimum requirement for successful Safe contract execution (avoiding reverts).
						Data:  callData,
					}
					err = deliverer.Deliver(&req, safeNonce)
					if err != nil {
						log.Crit(err.Error())
					}
				}

			}()
		}
		log.Info("send req done")

		time.Sleep(10 * time.Second)
		client.CloseSendMsg()
	}()

	for resp := range client.Response() {
		log.Info("respInfo", "id", resp.Id, "err", resp.Err)
	}

	locker = nonceStorage.NonceLockFrom(safeContractAddress)
	locker.Lock()
	endNonce, err := nonceStorage.GetNonce(safeContractAddress)
	locker.Unlock()
	if err != nil {
		log.Crit(err.Error())
	}

	if endNonce-startNonce != uint64(loopCount*2) {
		log.Crit("safe nonce has mistake")
	}
	log.Info("safe nonce As expected!")

	endNonceOnChain, err := safeContract.GetNonce()
	if err != nil {
		log.Crit("nonce has err")
	}
	log.Debug("safe nonce in chain for end", "nonce", endNonceOnChain)

	req3 := &message.Request{From: helper.Addr1, To: &safeContractAddress}
	err = deliverer.Deliver(req3, endNonce-1)
	if err == nil {
		log.Crit("nonce has err")
	}

	if err.Error() != "safeNonce is invalid" {
		log.Crit(err.Error())
	} else {
		log.Info(err.Error())
	}

	req4 := &message.Request{From: helper.Addr2, To: &safeContractAddress}
	err = deliverer.Deliver(req4, endNonce)

	if err == nil {
		log.Crit("from has err")
	}

	if err.Error() != "from address do not match" {
		log.Crit(err.Error())
	} else {
		log.Info(err.Error())
	}

}
