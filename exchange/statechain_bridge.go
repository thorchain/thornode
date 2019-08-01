package exchange

import (
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	st "github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// StatechainBridge will be used to forward requests to statechain
// move all logic related to forward request to statechain here thus we could test
// and add more error handling here
type StatechainBridge struct {
	clictx *context.CLIContext
}

// NewStatechainBridge create a new instance of StatechainBridge
func NewStatechainBridge(clictx *context.CLIContext) (*StatechainBridge, error) {
	if nil == clictx {
		return nil, errors.New("clictx is nil")
	}
	return &StatechainBridge{clictx: clictx}, nil
}

// Stake send the stake request to statechin
// for the owner address here , something I am not quite sure yet
// in this service,we might need to keep a map between the sender address of binance chain
// and the address in statechain, if it doesn't exist, we create it automatically
// thus we can keep a record of who stake what? and how much
func (b *StatechainBridge) Stake(name, ticker, r, token string, owner sdk.AccAddress, passphrase, memo string) (string, error) {
	if len(memo) > 0 {
		viper.Set(flags.FlagMemo, memo)
	}
	msg := st.NewMsgSetStakeData(name, ticker, r, token, owner)
	if err := msg.ValidateBasic(); nil != err {
		return "", errors.Wrap(err, "invalid message")
	}
	txBldr := auth.NewTxBuilderFromCLI().
		WithTxEncoder(utils.GetTxEncoder(b.clictx.Codec))

	res, err := completeAndBroadcastTxCLI(txBldr, *b.clictx, []sdk.Msg{msg}, passphrase)
	if nil != err {
		return "", errors.Wrap(err, "fail to send stake")
	}
	// success
	if res.Code == uint32(sdk.CodeOK) {
		return res.TxHash, nil
	}

	// somthing is wrong, let's find out and print out appropriate messages
	return "", errors.New(res.String())
}

// SendSwap send swap request to statechain
// first return parameter is txHash
func (b *StatechainBridge) SendSwap(source, target, amount, requester, destination string, owner sdk.AccAddress, passphrase, memo string) (string, error) {
	if len(memo) > 0 {
		viper.Set(flags.FlagMemo, memo)
	}
	msg := st.NewMsgSwap(source, target, amount, requester, destination, owner)
	if err := msg.ValidateBasic(); nil != err {
		return "", errors.Wrap(err, "invalid swap msg")
	}
	txBldr := auth.NewTxBuilderFromCLI().
		WithTxEncoder(utils.GetTxEncoder(b.clictx.Codec))
	res, err := completeAndBroadcastTxCLI(txBldr, *b.clictx, []sdk.Msg{msg}, passphrase)
	if nil != err {
		return "", errors.Wrap(err, "fail to send swap")
	}
	if res.Code == uint32(sdk.CodeOK) {
		return res.TxHash, nil
	}

	return "", errors.New(res.String())
}

// GetSwapTokenAmountFromHashWithRetry retry a few times before we give up
// TODO what do we do if we didn't get after many retry
func (b *StatechainBridge) GetSwapTokenAmountFromHashWithRetry(hash string) (string, error) {
	bf := backoff.NewExponentialBackOff()
	var lasterr error
	for {
		amount, err := b.GetSwapTokenAmountFromHash(hash)
		if nil != err {
			lasterr = err
			// given we need to retry it anyway , thus we don't write the error
			sleepTime := bf.NextBackOff()
			if sleepTime == backoff.Stop {
				break
			}
			time.Sleep(bf.NextBackOff())
			continue
		}
		return amount, nil
	}
	// if we get to this point, means we failed to get the result
	// need to consider what should we do here
	return "", errors.Wrap(lasterr, "fail after maximum retry")
}

// GetSwapTokenAmountFromHash we need to retrieve the tx detail based on the tx hash
// thus we could get the response from statechain
func (b *StatechainBridge) GetSwapTokenAmountFromHash(hash string) (string, error) {
	if len(hash) == 0 {
		return "", errors.New("hash is empty")
	}
	hashBuf, err := hex.DecodeString(hash)
	if nil != err {
		return "", errors.Wrapf(err, "fail to decode hash,%s should be hex encoded string")
	}
	rt, err := b.clictx.Client.Tx(hashBuf, true)
	if nil != err {
		return "", errors.Wrap(err, "fail to get tx detail based on tx hash")
	}
	data := rt.TxResult.GetData()
	if len(data) <= 0 {
		return "", errors.New("no data")
	}
	var swapResult struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &swapResult); nil != err {
		return "", errors.Wrap(err, "fail to unmarshal data")
	}
	return swapResult.Token, nil
}

func completeAndBroadcastTxCLI(txBldr authtypes.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg, passphrase string) (sdk.TxResponse, error) {
	txBldr, err := utils.PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return sdk.TxResponse{}, err
	}
	fromName := cliCtx.GetFromName()
	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, passphrase, msgs)
	if err != nil {
		return sdk.TxResponse{}, err
	}
	// broadcast to a Tendermint node
	return cliCtx.BroadcastTx(txBytes)
}
