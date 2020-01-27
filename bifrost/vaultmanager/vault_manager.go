package vaultmanager

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/types"
	"gitlab.com/thorchain/thornode/common"
)

// ThorchainClient interfae to thorchain client
type ThorchainClient interface {
	GetVaults() (types.Vaults, error)
}

// VaultManager is responsible for keeping up to date the mapping of pubKeys to chain addresses from thorchain
type VaultManager struct {
	asgardMutex *sync.RWMutex
	asgard      chainAddressPubKeyVaultMap

	yggdrasilMutex *sync.RWMutex
	yggdrasil      chainAddressPubKeyVaultMap

	logger          zerolog.Logger
	m               *metrics.Metrics
	thorchainClient ThorchainClient
	rawVaultsMutex  *sync.RWMutex
	rawVaults       types.Vaults
	wg              *sync.WaitGroup
	stopChan        chan struct{}
}

// chainAddressPubKeyVaultMap is the data structure for holding all processed pubkey to chain and address mapping
// chainAddressPubKeyVaultMap["BNB"]["tbnb1k5gnkdv0p3384ylylm39nke5tzc5l553xwxrf3"]["thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck"]
type chainAddressPubKeyVaultMap map[common.Chain]map[common.Address]common.PubKey

// NewVaultManager create a new vault manager with thorchain client
func NewVaultManager(thorchainClient ThorchainClient, m *metrics.Metrics) (*VaultManager, error) {
	return &VaultManager{
		logger:          log.With().Str("module", "VaultManager").Logger(),
		m:               m,
		thorchainClient: thorchainClient,
		rawVaultsMutex:  &sync.RWMutex{},
		asgardMutex:     &sync.RWMutex{},
		yggdrasilMutex:  &sync.RWMutex{},
		wg:              &sync.WaitGroup{},
		stopChan:        make(chan struct{}),
		asgard:          make(chainAddressPubKeyVaultMap),
		yggdrasil:       make(chainAddressPubKeyVaultMap),
	}, nil
}

// Start run a background work to keep all vaults up to date with thorchain
func (vaultManager *VaultManager) Start() error {
	vaultManager.logger.Info().Msg("starting vault manager")
	if err := vaultManager.fetchRawVaultsData(); err != nil {
		vaultManager.logger.Error().Err(err).Msg("failed to get vaults from thorchain")
	}
	vaultManager.wg.Add(1)
	go vaultManager.updateVaults()
	return nil
}

// Stop will kill the updateVaults worker
func (vaultManager *VaultManager) Stop() error {
	vaultManager.logger.Info().Msg("stopping vault manager")
	close(vaultManager.stopChan)
	vaultManager.wg.Wait()
	return nil
}

// GetPubKeys return all current pub keys in the vaults
func (vaultManager *VaultManager) GetPubKeys() []common.PubKey {
	var pubkeys []common.PubKey
	pubkeys = append(pubkeys, vaultManager.rawVaults.Asgard...)
	pubkeys = append(pubkeys, vaultManager.rawVaults.Yggdrasil...)
	return pubkeys
}

// GetYggdrasilPubKeys return yggdrail current pub keys in the vaults
func (vaultManager *VaultManager) GetYggdrasilPubKeys() []common.PubKey {
	return vaultManager.rawVaults.Yggdrasil
}

// GetAsgardPubKeys return asgard current pub keys in the vaults
func (vaultManager *VaultManager) GetAsgardPubKeys() []common.PubKey {
	return vaultManager.rawVaults.Asgard
}

// updateVaults will run as a worker and will update every minute the rawVaults and mappings
func (vaultManager *VaultManager) updateVaults() {
	vaultManager.logger.Info().Msg("starting updateVaults worker")
	defer vaultManager.logger.Info().Msg("stopped updateVaults worker")
	defer vaultManager.wg.Done()

	for {
		select {
		case <-vaultManager.stopChan:
			return
		case <-time.After(time.Minute):
			if err := vaultManager.fetchRawVaultsData(); err != nil {
				vaultManager.logger.Error().Err(err).Msg("failed to get vaults from thorchain")
			}
		}
	}
}

func (vaultManager *VaultManager) fetchRawVaultsData() error {
	rawVaults, err := vaultManager.thorchainClient.GetVaults()
	if err != nil {
		return err
	}

	// save rawVaults locally
	vaultManager.rawVaultsMutex.Lock()
	vaultManager.rawVaults = rawVaults
	vaultManager.rawVaultsMutex.Unlock()

	// process rawVaults into usable structure
	a := vaultManager.processRawAsgardVaults()
	vaultManager.asgardMutex.Lock()
	vaultManager.asgard = a
	vaultManager.asgardMutex.Unlock()

	y := vaultManager.processRawYggdrasilVaults()
	vaultManager.yggdrasilMutex.Lock()
	vaultManager.yggdrasil = y
	vaultManager.yggdrasilMutex.Unlock()

	return nil
}

// processRawAsgardVaults will process the raw Asgard vault into its chain/address mapping
func (vaultManager *VaultManager) processRawAsgardVaults() chainAddressPubKeyVaultMap {
	return vaultManager.processVaults(vaultManager.rawVaults.Asgard)
}

// processRawYggdrasilVaults will process the raw yaggdrasil vault into its chain/address mappings
func (vaultManager *VaultManager) processRawYggdrasilVaults() chainAddressPubKeyVaultMap {
	return vaultManager.processVaults(vaultManager.rawVaults.Yggdrasil)
}

// processVaults processes the given raw vault into its chain/address mapping
func (vaultManager *VaultManager) processVaults(vault []common.PubKey) chainAddressPubKeyVaultMap {
	var mapper = make(chainAddressPubKeyVaultMap)
	for _, pk := range vault {
		// Get BNB address
		bnbAddress, err := pk.GetAddress(common.BNBChain)
		if err != nil {
			log.Print(err.Error()) // TODO Review error handling.
		}

		// initialise inner map
		if mapper[common.BNBChain] == nil {
			mapper[common.BNBChain] = make(map[common.Address]common.PubKey)
		}
		mapper[common.BNBChain][bnbAddress] = pk

		// Get BTC address
		btcAddress, err := pk.GetAddress(common.BTCChain)
		if err != nil {
			log.Print(err.Error()) // TODO Review error handling.
		}

		// initialise inner map
		if mapper[common.BTCChain] == nil {
			mapper[common.BTCChain] = make(map[common.Address]common.PubKey)
		}
		mapper[common.BTCChain][btcAddress] = pk

		// TODO Add ETH support
	}
	return mapper
}

func (vaultManager *VaultManager) Get(chain common.Chain, address common.Address, vault chainAddressPubKeyVaultMap) common.PubKey {
	return vault[chain][address]
}

// HasKey determinate whether the given key is in the vault manager
func (vaultManager *VaultManager) HasKey(pk common.PubKey) bool {
	for _, item := range vaultManager.rawVaults.Yggdrasil {
		if item.Equals(pk) {
			return true
		}
	}
	for _, item := range vaultManager.rawVaults.Asgard {
		if item.Equals(pk) {
			return true
		}
	}
	return false
}
