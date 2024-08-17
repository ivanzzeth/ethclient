package ethclient

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/external"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/accounts/scwallet"
	"github.com/ethereum/go-ethereum/accounts/usbwallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/miner"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/pathdb"
)

func NewTestEthBackend(privateKey *ecdsa.PrivateKey, alloc types.GenesisAlloc, dataDir string) (*node.Node, error) {
	// Generate test chain.
	etherbase := crypto.PubkeyToAddress(privateKey.PublicKey)
	genesis := generateTestGenesis(etherbase, alloc)
	// Create node
	// nodeConfig := node.DefaultConfig
	nodeConfig := node.Config{}
	// nodeConfig.DataDir = dataDir
	nodeConfig.DataDir = "" // unless explicitly requested, use memory databases
	nodeConfig.UseLightweightKDF = true
	// p2p
	nodeConfig.P2P.MaxPeers = 0
	nodeConfig.P2P.ListenAddr = ""
	nodeConfig.P2P.NoDial = true
	nodeConfig.P2P.NoDiscovery = true
	nodeConfig.P2P.DiscoveryV5 = false

	stack, err := node.New(&nodeConfig)
	if err != nil {
		return nil, fmt.Errorf("can't create new node: %v", err)
	}

	// Setup genesis block
	for _, name := range []string{"chaindata", "lightchaindata"} {
		chaindb, err := stack.OpenDatabaseWithFreezer(name, 0, 0, "AncientName", "", false)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %v", err)
		}
		defer chaindb.Close()

		cachePreimages := false
		triedb := MakeTrieDatabase(chaindb, cachePreimages, false, genesis.IsVerkle())
		defer triedb.Close()

		_, hash, err := core.SetupGenesisBlock(chaindb, triedb, genesis)
		if err != nil {
			return nil, fmt.Errorf("failed to write genesis block: %v", err)
		}
		log.Info("Successfully wrote genesis state", "database", name, "hash", hash)
	}

	err = setAccountManagerBackends(&nodeConfig, stack.AccountManager(), stack.KeyStoreDir())
	if err != nil {
		return nil, fmt.Errorf("failed to set account manager backends: %v", err)
	}

	if err := saveMiner(stack, privateKey); err != nil {
		return nil, fmt.Errorf("save miner err: %v", err)
	}

	// minerConfig := miner.DefaultConfig
	minerConfig := miner.Config{}
	minerConfig.Etherbase = etherbase
	// minerConfig.PendingFeeRecipient = etherbase
	// minerConfig.GasPrice = big.NewInt(1)
	// Create Ethereum Service
	config := &ethconfig.Config{
		NetworkId: 1337,
		Genesis:   genesis,
		SyncMode:  downloader.FullSync,
		Miner:     minerConfig,
	}

	ethservice := RegisterEthService(stack, config)

	if err := stack.Start(); err != nil {
		return nil, fmt.Errorf("can't start test node: %v", err)
	}

	err = ethservice.Start()
	if err != nil {
		return nil, fmt.Errorf("can't start mining, err: %v", err)
	}

	ethservice.SetSynced()

	miner := ethservice.Miner()
	miner.SetEtherbase(etherbase)
	miner.Start()

	time.Sleep(2 * time.Second)

	if !miner.Mining() {
		return nil, fmt.Errorf("miner not working")
	}

	return stack, nil
}

// RegisterEthService adds an Ethereum client to the stack.
// The second return value is the full node instance.
func RegisterEthService(stack *node.Node, cfg *ethconfig.Config) *eth.Ethereum {
	backend, err := eth.New(stack, cfg)
	if err != nil {
		panic(fmt.Errorf("failed to register the Ethereum service: %v", err))
	}
	stack.RegisterAPIs(tracers.APIs(backend.APIBackend))
	return backend
}

// MakeTrieDatabase constructs a trie database based on the configured scheme.
func MakeTrieDatabase(disk ethdb.Database, preimage bool, readOnly bool, isVerkle bool) *triedb.Database {
	config := &triedb.Config{
		Preimages: preimage,
		IsVerkle:  isVerkle,
	}

	if readOnly {
		config.PathDB = pathdb.ReadOnly
	} else {
		config.PathDB = pathdb.Defaults
	}
	return triedb.NewDatabase(disk, config)
}

func saveMiner(stack *node.Node, minerPrivKey *ecdsa.PrivateKey) error {
	var ks *keystore.KeyStore
	if keystores := stack.AccountManager().Backends(keystore.KeyStoreType); len(keystores) > 0 {
		ks = keystores[0].(*keystore.KeyStore)
	} else {
		return fmt.Errorf("no any keystores")
	}

	passphrase := "123"
	account, err := ks.ImportECDSA(minerPrivKey, passphrase)
	if err != nil {
		return err
	}

	return ks.Unlock(account, passphrase)
}

func generateTestGenesis(miner common.Address, alloc types.GenesisAlloc) *core.Genesis {
	// db := rawdb.NewMemoryDatabase()
	// config := params.AllEthashProtocolChanges
	genesis := core.DeveloperGenesisBlock(1000000000000000000, &miner)
	genesis.Alloc = alloc
	// genesis := &core.Genesis{
	// 	Config:    config,
	// 	Alloc:     alloc,
	// 	ExtraData: []byte("test genesis"),
	// 	Timestamp: 9000,
	// }
	return genesis
}

func setAccountManagerBackends(conf *node.Config, am *accounts.Manager, keydir string) error {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP
	if conf.UseLightweightKDF {
		scryptN = keystore.LightScryptN
		scryptP = keystore.LightScryptP
	}

	// Assemble the supported backends
	if len(conf.ExternalSigner) > 0 {
		log.Info("Using external signer", "url", conf.ExternalSigner)
		if extBackend, err := external.NewExternalBackend(conf.ExternalSigner); err == nil {
			am.AddBackend(extBackend)
			return nil
		} else {
			return fmt.Errorf("error connecting to external signer: %v", err)
		}
	}

	// For now, we're using EITHER external signer OR local signers.
	// If/when we implement some form of lockfile for USB and keystore wallets,
	// we can have both, but it's very confusing for the user to see the same
	// accounts in both externally and locally, plus very racey.
	am.AddBackend(keystore.NewKeyStore(keydir, scryptN, scryptP))
	if conf.USB {
		// Start a USB hub for Ledger hardware wallets
		if ledgerhub, err := usbwallet.NewLedgerHub(); err != nil {
			log.Warn(fmt.Sprintf("Failed to start Ledger hub, disabling: %v", err))
		} else {
			am.AddBackend(ledgerhub)
		}
		// Start a USB hub for Trezor hardware wallets (HID version)
		if trezorhub, err := usbwallet.NewTrezorHubWithHID(); err != nil {
			log.Warn(fmt.Sprintf("Failed to start HID Trezor hub, disabling: %v", err))
		} else {
			am.AddBackend(trezorhub)
		}
		// Start a USB hub for Trezor hardware wallets (WebUSB version)
		if trezorhub, err := usbwallet.NewTrezorHubWithWebUSB(); err != nil {
			log.Warn(fmt.Sprintf("Failed to start WebUSB Trezor hub, disabling: %v", err))
		} else {
			am.AddBackend(trezorhub)
		}
	}
	if len(conf.SmartCardDaemonPath) > 0 {
		// Start a smart card hub
		if schub, err := scwallet.NewHub(conf.SmartCardDaemonPath, scwallet.Scheme, keydir); err != nil {
			log.Warn(fmt.Sprintf("Failed to start smart card hub, disabling: %v", err))
		} else {
			am.AddBackend(schub)
		}
	}

	return nil
}
