package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	swapserviceTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "swapservice transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	swapserviceTxCmd.AddCommand(client.PostCommands(
		GetCmdSetPoolData(cdc),
		GetCmdSetStakeData(cdc),
		GetCmdSwap(cdc),
		GetCmdSwapComplete(cdc),
		GetCmdUnstake(cdc),
		GetCmdUnStakeComplete(cdc),
		GetCmdSetTxHash(cdc),
		GetCmdSetAdminConfig(cdc),
	)...)

	return swapserviceTxCmd
}

// GetCmdSetPoolData is the CLI command for sending a SetPoolData transaction
func GetCmdSetPoolData(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-pool [ticker] [poolAddress] [status]",
		Short: "Set pool data",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			ticker, err := types.NewTicker(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetPoolData(ticker, args[1], types.GetPoolStatus(args[3]), cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSetStakeData is the CLI command for sending a SetStakeData transaction
func GetCmdSetStakeData(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-stake  [name] [ticker] [runes] [tokens] [stakerAddress] [requestTxHash]",
		Short: "Stake coins into a pool",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ticker, err := types.NewTicker(args[1])
			if err != nil {
				return err
			}

			txID, err := types.NewTxID(args[5])
			if err != nil {
				return err
			}

			runeAmt, err := types.NewAmount(args[2])
			if err != nil {
				return err
			}

			tokenAmt, err := types.NewAmount(args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetStakeData(args[0], ticker, runeAmt, tokenAmt, args[4], txID, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSwapComplete command to send MsgSwapComplete Message
func GetCmdSwapComplete(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-swap-complete [requestTxHash] [payTxHash]",
		Short: "Swap Complete",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			request, err := types.NewTxID(args[0])
			if err != nil {
				return err
			}

			pay, err := types.NewTxID(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgSwapComplete(request, pay, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSwap is the CLI command for swapping tokens
func GetCmdSwap(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-swap [requestTxHash] [source] [target] [amount] [requester] [destination] [tradeSlipLimit]",
		Short: "Swap coins",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			source, err := types.NewTicker(args[1])
			if err != nil {
				return err
			}

			target, err := types.NewTicker(args[2])
			if err != nil {
				return err
			}

			txID, err := types.NewTxID(args[0])
			if err != nil {
				return err
			}

			amt, err := types.NewAmount(args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgSwap(txID, source, target, amt, args[4], args[5], args[6], cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdUnstake command to unstake coins
func GetCmdUnstake(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unstake [name] [address] [percentage] [ticker] [requestTxHash]",
		Short: "Withdraw coins",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ticker, err := types.NewTicker(args[3])
			if err != nil {
				return err
			}

			txID, err := types.NewTxID(args[4])
			if err != nil {
				return err
			}

			percentage, err := types.NewAmount(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetUnStake(args[0], args[1], percentage, ticker, txID, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdUnStakeComplete command to send MsgUnStakeComplete Message
func GetCmdUnStakeComplete(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-unstake-complete [requestTxHash] [payTxHash]",
		Short: "unstake Complete",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			request, err := types.NewTxID(args[0])
			if err != nil {
				return err
			}

			pay, err := types.NewTxID(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgUnStakeComplete(request, pay, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSetTxHash command to send MsgSetTxHash Message from command line
func GetCmdSetTxHash(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-txhash [requestTxHash] [coins] [memo] [sender]",
		Short: "add a tx hash",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			txID, err := types.NewTxID(args[0])
			if err != nil {
				return err
			}

			tx := types.NewTxHash(txID, coins, args[2], args[3])
			msg := types.NewMsgSetTxHash([]types.TxHash{tx}, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSetAdminConfig command to set an admin config
func GetCmdSetAdminConfig(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-adming-config [key] [value]",
		Short: "set admin config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			msg := types.NewMsgSetAdminConfig(args[0], args[1], cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
