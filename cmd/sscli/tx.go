package main

import (
	"encoding/hex"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	app "github.com/jpthor/cosmos-swap"
	"github.com/jpthor/cosmos-swap/exchange"
)

// getTxDetailCmd
func getTxDetailCmd() *cobra.Command {
	exCmd := &cobra.Command{
		Use:   "gettx [txhash]",
		Short: "get tx detail based on tx hash",
		Long:  "gettx",
		Args:  cobra.ExactArgs(1),
		RunE:  getTxDetail,
	}
	return flags.PostCommands(exCmd)[0]
}

// getCliContext create a clicontext that can be used to send request to state chain
func getCliContext() *context.CLIContext {
	clictx := context.NewCLIContext().
		WithCodec(app.MakeCodec()).
		WithTrustNode(true).
		WithSimulation(false).
		WithGenerateOnly(false).
		WithBroadcastMode(flags.BroadcastSync)
	clictx.SkipConfirm = true
	return &clictx
}

func getTxDetail(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide txhash")
	}
	txHash := args[0]
	clictx := getCliContext()
	hashBuf, err := hex.DecodeString(txHash)
	if nil != err {
		return errors.Wrapf(err, "fail to decode hash,%s should be hex encoded string", txHash)
	}
	rt, err := clictx.Client.Tx(hashBuf, true)
	if nil != err {
		return errors.Wrap(err, "fail to get tx detail based on tx hash")
	}
	result, err := json.MarshalIndent(rt, "", "	")
	if nil != err {
		return errors.Wrap(err, "fail to marshal result to json")
	}
	cmd.Println(string(result))
	cdc := app.MakeCodec()
	var stdTx auth.StdTx
	if err := cdc.UnmarshalBinaryLengthPrefixed(rt.Tx, &stdTx); nil != err {
		return err
	}
	cmd.Println("tx:")
	txStr, err := json.MarshalIndent(stdTx, "", "	")
	if nil != err {
		return errors.Wrap(err, "fail to marshal tx to json")
	}
	cmd.Println(string(txStr))
	return nil
}

// TODO we need to remove it later for now let me keep it for test
func getTestSwap() *cobra.Command {
	testnetCmd := &cobra.Command{
		Use:   "testnet",
		Short: "subcommand used for test,will be removed later",
	}

	testSwapCmd := &cobra.Command{
		Use:   "swap [requestTxHash] [source] [target] [amount] [requester] [destination]",
		Short: "swap coins",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clictx := getCliContext()
			sb, err := exchange.NewStatechainBridge(clictx)
			if nil != err {
				return errors.Wrap(err, "fail to create statechain bridge")
			}
			txHash, err := sb.SendSwap(args[0], args[1], args[2], args[3], args[4], args[5], clictx.GetFromAddress(), "welcome@1", "")
			if nil != err {
				return errors.Wrap(err, "fail to send tx to statechain")
			}
			cmd.Println("txHash:", txHash)
			result, err := sb.GetSwapTokenAmountFromHashWithRetry(txHash)
			if nil != err {
				return errors.Wrapf(err, "fail to get tx detail for %s", txHash)
			}
			cmd.Printf("we should pay %s tokens", result)

			return nil
		},
	}
	testnetCmd.AddCommand(flags.PostCommands(testSwapCmd)...)
	return testnetCmd
}
