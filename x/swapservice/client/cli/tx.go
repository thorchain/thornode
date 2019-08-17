package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	common "gitlab.com/thorchain/bepswap/common"
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
		GetCmdUnstake(cdc),
		GetCmdSetTxIn(cdc),
		GetCmdSetAdminConfig(cdc),
	)...)

	return swapserviceTxCmd
}

// GetCmdSetPoolData is the CLI command for sending a SetPoolData transaction
func GetCmdSetPoolData(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-pool [ticker] [status]",
		Short: "Set pool data",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ticker, err := common.NewTicker(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetPoolData(ticker, types.GetPoolStatus(args[1]), cliCtx.GetFromAddress())
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
		Use:   "set-stake  [ticker] [runes] [tokens] [stakerAddress] [requestTxHash]",
		Short: "Stake coins into a pool",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ticker, err := common.NewTicker(args[0])
			if err != nil {
				return err
			}

			runeAmt, err := common.NewAmount(args[1])
			if err != nil {
				return err
			}

			tokenAmt, err := common.NewAmount(args[2])
			if err != nil {
				return err
			}

			bnbAddr, err := common.NewBnbAddress(args[3])
			if err != nil {
				return err
			}

			txID, err := common.NewTxID(args[4])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetStakeData(ticker, runeAmt, tokenAmt, bnbAddr, txID, cliCtx.GetFromAddress())
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
		Use:   "set-swap [requestTxHash] [source] [target] [amount] [requester] [destination] [target_price]",
		Short: "Swap coins",
		Args:  cobra.MinimumNArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			source, err := common.NewTicker(args[1])
			if err != nil {
				return err
			}

			target, err := common.NewTicker(args[2])
			if err != nil {
				return err
			}

			txID, err := common.NewTxID(args[0])
			if err != nil {
				return err
			}

			amt, err := common.NewAmount(args[3])
			if err != nil {
				return err
			}

			requester, err := common.NewBnbAddress(args[4])
			if err != nil {
				return err
			}
			destination := common.NoBnbAddress
			if len(args) > 5 {
				destination, err = common.NewBnbAddress(args[5])
				if err != nil {
					return err
				}
			}
			if destination.IsEmpty() {
				destination = requester
			}
			price := common.ZeroAmount
			if len(args) > 6 {
				price, err = common.NewAmount(args[6])
				if err != nil {
					return err
				}
			}

			msg := types.NewMsgSwap(txID, source, target, amt, requester, destination, price, cliCtx.GetFromAddress())
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
		Use:   "unstake [address] [ticker] [requestTxHash] [withdraw basis points]",
		Short: "Withdraw coins",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			bnbAddr, err := common.NewBnbAddress(args[0])
			if err != nil {
				return err
			}

			ticker, err := common.NewTicker(args[1])
			if err != nil {
				return err
			}

			txID, err := common.NewTxID(args[2])
			if err != nil {
				return err
			}
			withdrawBasisPoints := common.ZeroAmount
			if len(args) > 3 {
				withdrawBasisPoints, err = common.NewAmount(args[3])
				if err != nil {
					return err
				}
			}
			if !withdrawBasisPoints.GreaterThen(0) {
				withdrawBasisPoints = common.NewAmountFromFloat(types.MaxWithdrawBasisPoints)
			}
			msg := types.NewMsgSetUnStake(bnbAddr, withdrawBasisPoints, ticker, txID, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSetTxIn command to send MsgSetTxIn Message from command line
func GetCmdSetTxIn(cdc *codec.Codec) *cobra.Command {
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
			stateCoins, err := types.FromSdkCoins(coins)
			if nil != err {
				return err
			}
			txID, err := common.NewTxID(args[0])
			if err != nil {
				return err
			}

			bnbAddr, err := common.NewBnbAddress(args[3])
			if err != nil {
				return err
			}

			tx := types.NewTxIn(txID, stateCoins, args[2], bnbAddr)
			msg := types.NewMsgSetTxIn([]types.TxIn{tx}, cliCtx.GetFromAddress())
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
		Use:   "set-adming-config [key] [value] [bnbAddress]",
		Short: "set admin config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			bnb, err := common.NewBnbAddress(args[2])
			if err != nil {
				return err
			}

			key := types.GetAdminConfigKey(args[0])

			msg := types.NewMsgSetAdminConfig(key, args[1], bnb, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
