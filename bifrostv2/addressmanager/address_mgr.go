package addressmanager

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type AddressValidator interface {
	IsValidPoolAddress(addr string, chain common.Chain) (bool, common.ChainPoolInfo)
	IsValidAddress(addr string, chain common.Chain) bool
	AddPubKey(pk common.PubKeys)
	RemovePubKey(pk common.PubKeys)
}

// AddressManager manages the pool address
type AddressManager struct {
	poolAddresses types.PoolAddresses // current pool addresses
	rwMutex       *sync.RWMutex
	logger        zerolog.Logger
	chainHost     string // thorNode
	errCounter    *prometheus.CounterVec
	m             *metrics.Metrics
	wg            *sync.WaitGroup
	stopChan      chan struct{}
	closeOnce     sync.Once
}

// NewAddressManager create a new instance of AddressManager
func NewAddressManager(chainHost string, m *metrics.Metrics) (*AddressManager, error) {

	return &AddressManager{
		logger:     log.With().Str("module", "AddressManager").Logger(),
		chainHost:  chainHost,
		errCounter: m.GetCounterVec(metrics.PoolAddressManagerError),
		m:          m,
		wg:         &sync.WaitGroup{},
		stopChan:   make(chan struct{}),
		rwMutex:    &sync.RWMutex{},
	}, nil
}

// Start to poll addresses from thorchain
func (addrMr *AddressManager) Start() error {
	pa, err := addrMr.getPoolAddresses()
	if nil != err {
		return errors.Wrap(err, "fail to get pool addresses from thorchain")
	}
	addrMr.rwMutex.Lock()
	defer addrMr.rwMutex.Unlock()
	addrMr.poolAddresses = pa
	addrMr.wg.Add(1)
	go addrMr.updatePoolAddresses()
	return nil
}

// Stop pool address manager
func (addrMr *AddressManager) Stop() error {
	defer addrMr.logger.Info().Msg("stopped address manager")
	addrMr.closeOnce.Do(func() {
		close(addrMr.stopChan)
	})

	addrMr.logger.Debug().Msg("Waiting for all gorountines to close.")
	addrMr.wg.Wait()
	return nil
}

func (addrMr *AddressManager) updatePoolAddresses() {
	addrMr.logger.Info().Msg("start updatePoolAddresses")
	defer addrMr.logger.Info().Msg("stop updatePoolAddresses")
	defer addrMr.wg.Done()
	for {
		select {
		case <-addrMr.stopChan:
			return
		case <-time.After(time.Minute):
			pa, err := addrMr.getPoolAddresses()
			if nil != err {
				addrMr.logger.Error().Err(err).Msg("fail to get pool address from thorchain")
			}
			addrMr.rwMutex.Lock()
			addrMr.poolAddresses = pa
			addrMr.rwMutex.Unlock()
		}
	}
}

// func matchAddress(addr string, chain common.Chain, key common.PubKey) (bool, common.ChainPoolInfo) {
// 	cpi, err := common.NewChainPoolInfo(chain, key)
// 	if nil != err {
// 		return false, common.EmptyChainPoolInfo
// 	}
// 	if strings.EqualFold(cpi.PoolAddress.String(), addr) {
// 		return true, cpi
// 	}
// 	return false, common.EmptyChainPoolInfo
// }

// TODO Clean up, make cross chain, and write tests.
// IsValidPoolAddress check whether the given address is a pool addr
// func (addrMr *AddressManager) IsValidPoolAddress(addr string, chain common.Chain) (bool, common.ChainPoolInfo) {
// 	addrMr.rwMutex.RLock()
// 	defer addrMr.rwMutex.RUnlock()
// 	pa := addrMr.poolAddresses
// 	bnbChainCurrent := pa.Current.GetByChain(common.BNBChain)
// 	if nil == bnbChainCurrent {
// 		return false, common.EmptyChainPoolInfo
// 	}
//
// 	matchCurrent, cpi := matchAddress(addr, chain, bnbChainCurrent.PubKey)
// 	if matchCurrent {
// 		return matchCurrent, cpi
// 	}
// 	bnbChainPrevious := pa.Previous.GetByChain(common.BNBChain)
// 	if nil != bnbChainPrevious {
// 		matchPrevious, cpi := matchAddress(addr, chain, bnbChainPrevious.PubKey)
// 		if matchPrevious {
// 			return matchPrevious, cpi
// 		}
// 	}
// 	bnbChainNext := pa.Previous.GetByChain(common.BNBChain)
// 	if nil != bnbChainNext {
// 		matchNext, cpi := matchAddress(addr, chain, bnbChainNext.PubKey)
// 		if matchNext {
// 			return matchNext, cpi
// 		}
// 	}
// 	addrMr.logger.Debug().Str("previous", pa.Previous.String()).
// 		Str("current", pa.Current.String()).
// 		Str("next", pa.Next.String()).
// 		Str("addr", addr).Msg("doesn't match")
// 	return false, common.EmptyChainPoolInfo
// }

// getPoolAddresses from thorchain
func (addrMr *AddressManager) getPoolAddresses() (types.PoolAddresses, error) {
	var uri *url.URL
	var err error

	if strings.Contains(addrMr.chainHost, "http") {
		uri, err = url.Parse(addrMr.chainHost)
		if err != nil {
			return types.EmptyPoolAddresses, errors.Wrap(err, "error parsing chain_host")
		}
	} else {
		uri = &url.URL{
			Scheme: "http",
			Host:   addrMr.chainHost,
		}
	}
	uri.Path = "/thorchain/pooladdresses"

	resp, err := retryablehttp.Get(uri.String())
	if nil != err {
		return types.EmptyPoolAddresses, errors.Wrap(err, "fail to get pool addresses from thorchain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			addrMr.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return types.EmptyPoolAddresses, errors.Wrap(err, "fail to get pool addresses from thorchain")
	}
	var pa types.PoolAddresses
	buf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return types.EmptyPoolAddresses, errors.Wrap(err, "fail to read response body")
	}
	if err := json.Unmarshal(buf, &pa); err != nil {
		addrMr.errCounter.WithLabelValues("fail_unmarshal_pool_address", "").Inc()
		return types.EmptyPoolAddresses, errors.Wrap(err, "fail to unmarshal pool address")
	}
	return pa, nil
}
