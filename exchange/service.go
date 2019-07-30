package exchange

import (
	"fmt"
	"os"
	"sync"

	"github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/client/websocket"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"

	"github.com/jpthor/cosmos-swap/config"
	st "github.com/jpthor/cosmos-swap/x/swapservice/types"
)

type Service struct {
	cfg    config.Settings
	logger zerolog.Logger
	ws     *Wallets
	wg     *sync.WaitGroup
	dc     client.DexClient
	clictx *context.CLIContext
	quit   chan struct{}
}

// NewService create a new instance of service which will talk to the exchange
func NewService(clictx *context.CLIContext, cfg config.Settings, ws *Wallets, logger zerolog.Logger) (*Service, error) {
	if nil == clictx {
		return nil, errors.New("invalid clictx")
	}
	return &Service{
		cfg:    cfg,
		ws:     ws,
		wg:     &sync.WaitGroup{},
		clictx: clictx,
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
	s.dc = c
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

// for the owner address here , something I am not quite sure yet
// in this service,we might need to keep a map between the sender address of binance chain
// and the address in statechain, if it doesn't exist, we create it automatically
// thus we can keep a record of who stake what? and how much
func (s *Service) Stake(name, ticker, r, token string, owner sdk.AccAddress, passphrase string) error {
	msg := st.NewMsgSetStakeData(name, ticker, r, token, owner)
	if err := msg.ValidateBasic(); nil != err {
		return errors.Wrap(err, "invalid message")
	}
	txBldr := auth.NewTxBuilderFromCLI().
		WithTxEncoder(utils.GetTxEncoder(s.clictx.Codec))

	res, err := completeAndBroadcastTxCLI(txBldr, *s.clictx, []sdk.Msg{msg}, passphrase)
	if nil != err {
		return errors.Wrap(err, "fail to send stake")
	}
	// success
	if res.Code == uint32(codes.OK) {
		return nil
	}
	// somthing is wrong, let's find out and print out appropriate messages
	return errors.New(res.String())
}

func (s *Service) SendSwap(source, target, amount, requester, destination string, owner sdk.AccAddress, passphrase string) error {
	// TODO let's add more validations
	msg := st.NewMsgSwap(source, target, amount, requester, destination, owner)
	if err := msg.ValidateBasic(); nil != err {
		return errors.Wrap(err, "invalid swap msg")
	}
	txBldr := auth.NewTxBuilderFromCLI().
		WithTxEncoder(utils.GetTxEncoder(s.clictx.Codec))
	res, err := completeAndBroadcastTxCLI(txBldr, *s.clictx, []sdk.Msg{msg}, passphrase)
	if nil != err {
		return errors.Wrap(err, "fail to send swap")
	}

	if res.Code == uint32(codes.OK) {
		return nil
	}
	return errors.New(res.String())
}

func completeAndBroadcastTxCLI(txBldr authtypes.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg, passphrase string) (sdk.TxResponse, error) {
	txBldr, err := utils.PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	fromName := cliCtx.GetFromName()

	if txBldr.SimulateAndExecute() || cliCtx.Simulate {
		txBldr, err = utils.EnrichWithGas(txBldr, cliCtx, msgs)
		if err != nil {
			return sdk.TxResponse{}, err
		}

		gasEst := utils.GasEstimateResponse{GasEstimate: txBldr.Gas()}
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", gasEst.String())
	}

	//if !cliCtx.SkipConfirm {
	//	stdSignMsg, err := txBldr.BuildSignMsg(msgs)
	//	if err != nil {
	//		return err
	//	}
	//
	//	var json []byte
	//	if viper.GetBool(flags.FlagIndentResponse) {
	//		json, err = cliCtx.Codec.MarshalJSONIndent(stdSignMsg, "", "  ")
	//		if err != nil {
	//			panic(err)
	//		}
	//	} else {
	//		json = cliCtx.Codec.MustMarshalJSON(stdSignMsg)
	//	}
	//
	//	_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", json)
	//
	//	buf := bufio.NewReader(os.Stdin)
	//	ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf)
	//	if err != nil || !ok {
	//		_, _ = fmt.Fprintf(os.Stderr, "%s\n", "cancelled transaction")
	//		return err
	//	}
	//}
	//passphrase, err := ckeys.GetPassphrase(fromName)
	//if err != nil {
	//	return sdk.TxResponse{}, err
	//}

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, passphrase, msgs)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	// broadcast to a Tendermint node
	return cliCtx.BroadcastTx(txBytes)

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
