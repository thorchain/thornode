package exchange

import (
	"sync"

	"github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/client/websocket"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"bep2-swap-node/config"
)

type Service struct {
	cfg    config.Settings
	logger zerolog.Logger
	ws     *Wallets
	wg     *sync.WaitGroup
	quit   chan struct{}
}

// NewService create a new instance of service which will talk to the exchange
func NewService(cfg config.Settings, ws *Wallets, logger zerolog.Logger) (*Service, error) {
	return &Service{
		cfg:    cfg,
		ws:     ws,
		wg:     &sync.WaitGroup{},
		quit:   make(chan struct{}),
		logger: logger.With().Str("module", "service").Logger(),
	}, nil
}

func (s *Service) Start() error {
	s.logger.Info().Msg("start")
	for _, symbol := range s.cfg.Pools {
		s.logger.Info().Msgf("start to process %s", symbol)
		w, err := s.ws.GetWallet(symbol)
		if nil != err {
			return errors.Wrap(err, "fail to get wallet")
		}
		s.logger.Info().Msgf("wallet:%s", w)
		if err := s.startProcess(w); nil != err {
			return errors.Wrapf(err, "fail to start processing,%s", symbol)
		}
	}
	return nil
}
func (s *Service) startProcess(wallet *Bep2Wallet) error {
	keyManager, err := keys.NewPrivateKeyManager(wallet.PrivateKey)
	if nil != err {
		return errors.Wrap(err, "fail to create private key manager")
	}
	c, err := client.NewDexClient(s.cfg.DexBaseUrl, types.TestNetwork, keyManager)
	if nil != err {
		return errors.Wrap(err, "fail to create dex client")
	}

	if err := c.SubscribeAccountEvent(c.GetKeyManager().GetAddr().String(), s.quit, s.receiveAccountEvent, s.onAccountEventError, func() {
		s.logger.Info().Msg("close account event subscription")
	}); nil != err {
		return errors.Wrap(err, "fail to subscribe account event")
	}
	if err := c.SubscribeOrderEvent(c.GetKeyManager().GetAddr().String(), s.quit, s.receiveOrder, s.onAccountEventError, func() {
		s.logger.Info().Msg("close order event subscription")
	}); nil != err {
		return errors.Wrap(err, "fail to subscribe order event")
	}
	return nil
}
func (s *Service) receiveOrder(events []*websocket.OrderEvent) {
	for _, e := range events {
		s.logger.Info().Msgf("order event:%v \n", e)
	}
}
func (s *Service) receiveAccountEvent(e *websocket.AccountEvent) {
	/*

		TODO:

		We need to pull the "memo" out of the transaction, so we can decide what to do with the tx

		To do this, we will use the tx event information to get the actual transaction
		It may be, that we get the block number ae.EventBlock and get the block, and iterate over the transactions in it

		Once we have obtained the transaction memo, we will use a switch statement to handle the various memo's we support
		More on that later...

	*/
	s.logger.Info().Msgf("event-type:%s , event-time:%d \n", e.EventType, e.EventTime)
	for _, item := range e.Balances {
		s.logger.Info().Msgf("asset:%s free: %d frozen: %d locked %d \n", item.Asset, item.Free, item.Frozen, item.Locked)
	}
}
func (s *Service) onAccountEventError(e error) {
	s.logger.Err(e).Msg("account event error")
}
func (s *Service) Stop() error {
	close(s.quit)
	return nil
}
