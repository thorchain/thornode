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
	)...)

	return swapserviceTxCmd
}

// GetCmdSetPoolData is the CLI command for sending a SetPoolData transaction
func GetCmdSetPoolData(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-pool [token name] [ticker] [poolAddress] [status]",
		Short: "Set pool data",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			msg := types.NewMsgSetPoolData(args[0], args[1], args[2], types.GetPoolStatus(args[3]), cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
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
			msg := types.NewMsgSetStakeData(args[0], args[1], args[2], args[3], args[4], args[5], cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
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
			msg := types.NewMsgSwapComplete(args[0], args[1], cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
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
			msg := types.NewMsgSwap(args[0], args[1], args[2], args[3], args[4], args[5], args[6], cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
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
			msg := types.NewMsgSetUnStake(args[0], args[1], args[2], args[3], args[4], cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
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
			msg := types.NewMsgUnStakeComplete(args[0], args[1], cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
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
		Use:   "set-txhash [requestTxHash] ",
		Short: "mark a txhash Complete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			msg := types.NewMsgSetTxHash(args[0], cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
