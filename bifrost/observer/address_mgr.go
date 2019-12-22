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

	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/common"
)

type AddressValidator interface {
	IsValidPoolAddress(addr string, chain common.Chain) (bool, common.ChainPoolInfo)
	IsValidAddress(addr string, chain common.Chain) bool
	AddPubKey(pk common.PubKey)
	RemovePubKey(pk common.PubKey)
}

// AddressManager it manage the pool address
type AddressManager struct {
	cdc        *codec.Codec
	pubkeys    []common.PubKey
	rwMutex    *sync.RWMutex
	logger     zerolog.Logger
	chainHost  string // statechain host
	errCounter *prometheus.CounterVec
	m          *metrics.Metrics
	wg         *sync.WaitGroup
	stopChan   chan struct{}
}

// NewAddressManager create a new instance of AddressManager
func NewAddressManager(chainHost string, m *metrics.Metrics) (*AddressManager, error) {
	return &AddressManager{
		cdc:        thorclient.MakeCodec(),
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
	pubkeys, err := pam.getPubkeys()
	if nil != err {
		return errors.Wrap(err, "fail to get pubkeys from thorchain")
	}
	pam.rwMutex.Lock()
	defer pam.rwMutex.Unlock()
	pam.pubkeys = pubkeys
	go pam.updatePubKeys()
	return nil
}

// Stop pool address manager
func (pam *AddressManager) Stop() error {
	defer pam.logger.Info().Msg("pool address manager stopped")
	close(pam.stopChan)
	pam.wg.Wait()
	return nil
}

func (pam *AddressManager) AddPubKey(pk common.PubKey) {
	pam.rwMutex.Lock()
	found := false
	for _, pubkey := range pam.pubkeys {
		if pk.Equals(pubkey) {
			break
		}
	}
	if !found {
		pam.pubkeys = append(pam.pubkeys, pk)
	}
	pam.rwMutex.Unlock()
}

func (pam *AddressManager) RemovePubKey(pk common.PubKey) {
	pam.rwMutex.Lock()
	for i, pubkey := range pam.pubkeys {
		if pk.Equals(pubkey) {
			pam.pubkeys[i] = pam.pubkeys[len(pam.pubkeys)-1]     // Copy last element to index i.
			pam.pubkeys[len(pam.pubkeys)-1] = common.EmptyPubKey // Erase last element (write zero value).
			pam.pubkeys = pam.pubkeys[:len(pam.pubkeys)-1]       // Truncate slice.
			break
		}
	}
	pam.rwMutex.Unlock()
}

func (pam *AddressManager) updatePubKeys() {
	pam.logger.Info().Msg("start to update pub keys")
	defer pam.logger.Info().Msg("stop to update pub keys")
	defer pam.wg.Done()
	for {
		select {
		case <-pam.stopChan:
			return
		case <-time.After(time.Minute):
			pubkeys, err := pam.getPubkeys()
			if nil != err {
				pam.logger.Error().Err(err).Msg("fail to get pubkeys from thorchain")
			}
			for _, pk := range pubkeys {
				pam.AddPubKey(pk)
			}
		}
	}
}

func matchAddress(addr string, chain common.Chain, key common.PubKey) (bool, common.ChainPoolInfo) {
	cpi, err := common.NewChainPoolInfo(chain, key)
	if err != nil {
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

	for _, pk := range pam.pubkeys {
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

	for _, pk := range pam.pubkeys {
		ok, cpi := matchAddress(addr, chain, pk)
		if ok {
			return ok, cpi
		}
	}
	return false, common.EmptyChainPoolInfo
}

// getPubkeys from thorchain
func (pam *AddressManager) getPubkeys() ([]common.PubKey, error) {
	uri := url.URL{
		Scheme: "http",
		Host:   pam.chainHost,
		Path:   "/thorchain/vaults/pubkeys",
	}
	resp, err := retryablehttp.Get(uri.String())
	if nil != err {
		return nil, errors.Wrap(err, "fail to get pubkeys from thorchain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			pam.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(err, "fail to get pubkeys from thorchain")
	}

	var pubs struct {
		Asgard    []common.PubKey `json:"asgard"`
		Yggdrasil []common.PubKey `json:"yggdrasil"`
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return nil, errors.Wrap(err, "fail to read response body")
	}
	if err := pam.cdc.UnmarshalJSON(buf, &pubs); nil != err {
		pam.errCounter.WithLabelValues("fail_unmarshal_pubkeys", "").Inc()
		return nil, errors.Wrap(err, "fail to unmarshal pubkeys")
	}

	pubkeys := append(pubs.Asgard, pubs.Yggdrasil...)
	return pubkeys, nil
}
