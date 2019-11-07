package observer

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/thornode/bifrost/metrics"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/statechain"
	"gitlab.com/thorchain/bepswap/thornode/common"
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/types"
)

type AddressValidator interface {
	IsValidPoolAddress(addr string, chain common.Chain) (bool, common.ChainPoolInfo)
	IsValidAddress(addr string, chain common.Chain) bool
	AddPubKey(pk common.PubKeys)
	RemovePubKey(pk common.PubKeys)
}

// AddressManager it manage the pool address
type AddressManager struct {
	cdc           *codec.Codec
	addresses     []common.PubKeys
	poolAddresses types.PoolAddresses // current pool addresses
	rwMutex       *sync.RWMutex
	logger        zerolog.Logger
	chainHost     string // statechain host
	errCounter    *prometheus.CounterVec
	m             *metrics.Metrics
	wg            *sync.WaitGroup
	stopChan      chan struct{}
}

// NewAddressManager create a new instance of AddressManager
func NewAddressManager(chainHost string, m *metrics.Metrics) (*AddressManager, error) {
	return &AddressManager{
		cdc:        statechain.MakeCodec(),
		logger:     log.With().Str("module", "statechain_bridge").Logger(),
		chainHost:  chainHost,
		errCounter: m.GetCounterVec(metrics.PoolAddressManagerError),
		m:          m,
		wg:         &sync.WaitGroup{},
		stopChan:   make(chan struct{}),
		rwMutex:    &sync.RWMutex{},
	}, nil
}

// Start to poll poll addresses from statechain
func (pam *AddressManager) Start() error {
	pam.wg.Add(1)
	pa, err := pam.getPoolAddresses()
	if nil != err {
		return errors.Wrap(err, "fail to get pool addresses from statechain")
	}
	pam.rwMutex.Lock()
	defer pam.rwMutex.Unlock()
	pam.poolAddresses = pa
	currentAddr, err := pa.Current.GetByChain(common.BNBChain).GetAddress()
	if nil != err {
		return err
	}
	pam.logger.Info().Str("addr", currentAddr.String()).Msg("current pool address")
	go pam.updatePoolAddresses()
	return nil
}

// Stop pool address manager
func (pam *AddressManager) Stop() error {
	defer pam.logger.Info().Msg("pool address manager stopped")
	close(pam.stopChan)
	pam.wg.Wait()
	return nil
}

func (pam *AddressManager) AddPubKey(pk common.PubKeys) {
	pam.rwMutex.Lock()
	found := false
	for _, pubkey := range pam.addresses {
		if pk.Equals(pubkey) {
			break
		}
	}
	if !found {
		pam.addresses = append(pam.addresses, pk)
	}
	pam.rwMutex.Unlock()
}

func (pam *AddressManager) RemovePubKey(pk common.PubKeys) {
	pam.rwMutex.Lock()
	for i, pubkey := range pam.addresses {
		if pk.Equals(pubkey) {
			pam.addresses[i] = pam.addresses[len(pam.addresses)-1] // Copy last element to index i.
			pam.addresses[len(pam.addresses)-1] = common.PubKeys{} // Erase last element (write zero value).
			pam.addresses = pam.addresses[:len(pam.addresses)-1]   // Truncate slice.
			break
		}
	}
	pam.rwMutex.Unlock()
}

func (pam *AddressManager) updatePoolAddresses() {
	pam.logger.Info().Msg("start to update pool addresses")
	defer pam.logger.Info().Msg("stop to update pool addresses")
	defer pam.wg.Done()
	for {
		select {
		case <-pam.stopChan:
			return
		case <-time.After(time.Minute):
			pa, err := pam.getPoolAddresses()
			if nil != err {
				pam.logger.Error().Err(err).Msg("fail to get pool address from statechain")
			}
			pam.rwMutex.Lock()
			pam.poolAddresses = pa
			pam.rwMutex.Unlock()
		}
	}
}

func matchAddress(addr string, chain common.Chain, key common.PubKey) (bool, common.ChainPoolInfo) {
	cpi, err := common.NewChainPoolInfo(chain, key)
	if nil != err {
		return false, common.EmptyChainPoolInfo
	}
	if strings.EqualFold(cpi.PoolAddress.String(), addr) {
		return true, cpi
	}
	return false, common.EmptyChainPoolInfo
}

// IsValidAddress check whether the given address is a monitored address
func (pam *AddressManager) IsValidAddress(addr string, chain common.Chain) bool {
	pam.rwMutex.RLock()
	defer pam.rwMutex.RUnlock()

	for _, pk := range pam.addresses {
		pkAddr, _ := pk.GetAddress(chain)
		address, _ := common.NewAddress(addr)
		if address.Equals(pkAddr) && !pkAddr.IsEmpty() && !address.IsEmpty() {
			return true
		}
	}

	return false
}

// IsValidPoolAddress check whether the given address is a pool addr
func (pam *AddressManager) IsValidPoolAddress(addr string, chain common.Chain) (bool, common.ChainPoolInfo) {
	pam.rwMutex.RLock()
	defer pam.rwMutex.RUnlock()
	pa := pam.poolAddresses
	bnbChainCurrent := pa.Current.GetByChain(common.BNBChain)
	if nil == bnbChainCurrent {
		return false, common.EmptyChainPoolInfo
	}

	matchCurrent, cpi := matchAddress(addr, chain, bnbChainCurrent.PubKey)
	if matchCurrent {
		return matchCurrent, cpi
	}
	bnbChainPrevious := pa.Previous.GetByChain(common.BNBChain)
	if nil != bnbChainPrevious {
		matchPrevious, cpi := matchAddress(addr, chain, bnbChainPrevious.PubKey)
		if matchPrevious {
			return matchPrevious, cpi
		}
	}
	bnbChainNext := pa.Previous.GetByChain(common.BNBChain)
	if nil != bnbChainNext {
		matchNext, cpi := matchAddress(addr, chain, bnbChainNext.PubKey)
		if matchNext {
			return matchNext, cpi
		}
	}
	pam.logger.Debug().Str("previous", pa.Previous.String()).
		Str("current", pa.Current.String()).
		Str("next", pa.Next.String()).
		Str("addr", addr).Msg("doesn't match")
	return false, common.EmptyChainPoolInfo
}

// getPoolAddresses from statechain
func (pam *AddressManager) getPoolAddresses() (types.PoolAddresses, error) {
	uri := url.URL{
		Scheme: "http",
		Host:   pam.chainHost,
		Path:   "/thorchain/pooladdresses",
	}
	resp, err := retryablehttp.Get(uri.String())
	if nil != err {
		return types.EmptyPoolAddresses, errors.Wrap(err, "fail to get pool addresses from statechain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			pam.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return types.EmptyPoolAddresses, errors.Wrap(err, "fail to get pool addresses from statechain")
	}
	var pa types.PoolAddresses
	buf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return types.EmptyPoolAddresses, errors.Wrap(err, "fail to read response body")
	}
	if err := pam.cdc.UnmarshalJSON(buf, &pa); nil != err {
		pam.errCounter.WithLabelValues("fail_unmarshal_pool_address", "").Inc()
		return types.EmptyPoolAddresses, errors.Wrap(err, "fail to unmarshal pool address")
	}
	return pa, nil
}
