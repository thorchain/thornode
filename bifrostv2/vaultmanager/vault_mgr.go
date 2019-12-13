package vaultmanager

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/bifrostv2/thorclient"
	"gitlab.com/thorchain/thornode/bifrostv2/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

// VaultManager is responsible for keeping up to date the mapping of pubKeys to chain addresses from thorchain
type VaultManager struct {
	asgardMutex *sync.RWMutex
	asgard      chainAddressPubKeyVaultMap

	yggdrasilMutex *sync.RWMutex
	yggdrasil      chainAddressPubKeyVaultMap

	logger         zerolog.Logger
	m              *metrics.Metrics
	thorClient     *thorclient.Client
	rawVaultsMutex *sync.RWMutex
	rawVaults      types.Vaults
	wg             *sync.WaitGroup
	stopChan       chan struct{}
}

// chainAddressPubKeyVaultMap is the data structure for holding all processed pubkey to chain and address mapping
// chainAddressPubKeyVaultMap["BNB"]["tbnb1k5gnkdv0p3384ylylm39nke5tzc5l553xwxrf3"]["thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck"]
type chainAddressPubKeyVaultMap map[common.Chain]map[common.Address]common.PubKey

func NewVaultManager(thorClient *thorclient.Client, m *metrics.Metrics) (*VaultManager, error) {
	return &VaultManager{
		logger:         log.With().Str("module", "VaultManager").Logger(),
		m:              m,
		thorClient:     thorClient,
		rawVaultsMutex: &sync.RWMutex{},
		asgardMutex:    &sync.RWMutex{},
		yggdrasilMutex: &sync.RWMutex{},
		wg:             &sync.WaitGroup{},
		stopChan:       make(chan struct{}),
		asgard:         make(chainAddressPubKeyVaultMap),
		yggdrasil:      make(chainAddressPubKeyVaultMap),
	}, nil
}

// Start run a background work to keep all vaults up to date with thorchain
func (vaultMgr *VaultManager) Start() error {
	vaultMgr.logger.Info().Msg("starting vault manager")
	if err := vaultMgr.fetchRawVaultsData(); err != nil {
		vaultMgr.logger.Error().Err(err).Msg("failed to get vaults from thorchain")
	}
	vaultMgr.wg.Add(1)
	go vaultMgr.updateVaults()
	return nil
}

// Stop will kill the updateVaults worker
func (vaultMgr *VaultManager) Stop() error {
	vaultMgr.logger.Info().Msg("stopping vault manager")
	close(vaultMgr.stopChan)
	vaultMgr.wg.Wait()
	return nil
}

// updateVaults will run as a worker and will update every minute the rawVaults and mappings
func (vaultMgr *VaultManager) updateVaults() {
	vaultMgr.logger.Info().Msg("starting updateVaults worker")
	defer vaultMgr.logger.Info().Msg("stopped updateVaults worker")
	defer vaultMgr.wg.Done()

	for {
		select {
		case <-vaultMgr.stopChan:
			return
		case <-time.After(time.Minute):
			if err := vaultMgr.fetchRawVaultsData(); err != nil {
				vaultMgr.logger.Error().Err(err).Msg("failed to get vaults from thorchain")
			}
		}
	}
}

func (vaultMgr *VaultManager) fetchRawVaultsData() error {
	rawVaults, err := vaultMgr.thorClient.GetVaults()
	if err != nil {
		return err
	}

	// save rawVaults locally
	vaultMgr.rawVaultsMutex.Lock()
	vaultMgr.rawVaults = rawVaults
	vaultMgr.rawVaultsMutex.Unlock()

	// process rawVaults into usable structure
	a := vaultMgr.processRawAsgardVaults()
	vaultMgr.asgardMutex.Lock()
	vaultMgr.asgard = a
	vaultMgr.asgardMutex.Unlock()

	y := vaultMgr.processRawYggdrasilVaults()
	vaultMgr.yggdrasilMutex.Lock()
	vaultMgr.yggdrasil = y
	vaultMgr.yggdrasilMutex.Unlock()

	return nil
}

// processRawAsgardVaults will process the raw Asgard vault into its chain/address mapping
func (vaultMgr *VaultManager) processRawAsgardVaults() chainAddressPubKeyVaultMap {
	return vaultMgr.processVaults(vaultMgr.rawVaults.Asgard)
}

// processRawYggdrasilVaults will process the raw yaggdrasil vault into its chain/address mappings
func (vaultMgr *VaultManager) processRawYggdrasilVaults() chainAddressPubKeyVaultMap {
	return vaultMgr.processVaults(vaultMgr.rawVaults.Yggdrasil)
}

// processVaults processes the given raw vault into its chain/address mapping
func (vaultMgr *VaultManager) processVaults(vault []common.PubKey) chainAddressPubKeyVaultMap {
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

func (vaultMgr *VaultManager) get(chain common.Chain, address common.Address, vault chainAddressPubKeyVaultMap) common.PubKey {
	return vault[chain][address]
}
