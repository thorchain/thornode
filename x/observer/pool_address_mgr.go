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

	"gitlab.com/thorchain/bepswap/thornode/common"
	"gitlab.com/thorchain/bepswap/thornode/x/metrics"
	"gitlab.com/thorchain/bepswap/thornode/x/statechain"
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/types"
)

type PoolAddressValidator interface {
	IsValidPoolAddress(addr string, chain common.Chain) (bool, common.ChainPoolInfo)
}

// PoolAddressManager it manage the pool address
type PoolAddressManager struct {
	cdc           *codec.Codec
	poolAddresses types.PoolAddresses // current pool addresses
	rwMutex       *sync.RWMutex
	logger        zerolog.Logger
	chainHost     string // statechain host
	errCounter    *prometheus.CounterVec
	m             *metrics.Metrics
	wg            *sync.WaitGroup
	stopChan      chan struct{}
}

// NewPoolAddressManager create a new instance of PoolAddressManager
func NewPoolAddressManager(chainHost string, m *metrics.Metrics) (*PoolAddressManager, error) {
	return &PoolAddressManager{
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
func (pam *PoolAddressManager) Start() error {
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
func (pam *PoolAddressManager) Stop() error {
	defer pam.logger.Info().Msg("pool address manager stopped")
	close(pam.stopChan)
	pam.wg.Wait()
	return nil
}

func (pam *PoolAddressManager) updatePoolAddresses() {
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

// IsValidPoolAddress check whether the given address is a pool addr
func (pam *PoolAddressManager) IsValidPoolAddress(addr string, chain common.Chain) (bool, common.ChainPoolInfo) {
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
func (pam *PoolAddressManager) getPoolAddresses() (types.PoolAddresses, error) {
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
