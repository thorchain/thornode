package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.com/thorchain/bepswap/thornode/common"

	appCmd "gitlab.com/thorchain/bepswap/thornode/cmd"
	"gitlab.com/thorchain/bepswap/thornode/x/swapservice/types"
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
		GetCmdSetAdminConfig(cdc),
		GetCmdSetTrustAccount(cdc),
		GetCmdEndPool(cdc),
		GetCmdSetVersion(cdc),
	)...)

	return swapserviceTxCmd
}

// GetCmdSetVersion command to set an admin config
func GetCmdSetVersion(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-version",
		Short: "update registered version",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			msg := types.NewMsgSetVersion(appCmd.Version, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSetAdminConfig command to set an admin config
func GetCmdSetAdminConfig(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-admin-config [key] [value]",
		Short: "set admin config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			key := types.GetAdminConfigKey(args[0])

			msg := types.NewMsgSetAdminConfig(key, args[1], cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSetTrustAccount command to add a trust account
func GetCmdSetTrustAccount(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-trust-account  [observer_address] [validator_consensus_pub_key]",
		Short: "set trust account, the account use to sign this tx has to be whitelist first",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			observer, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return errors.Wrap(err, "fail to parse observer address")
			}

			validatorConsPubKey, err := sdk.GetConsPubKeyBech32(args[1])
			if err != nil {
				return errors.Wrap(err, "fail to parse validator consensus public key")
			}
			validatorConsPubKeyStr, err := sdk.Bech32ifyConsPub(validatorConsPubKey)
			if err != nil {
				return errors.Wrap(err, "fail to convert public key to string")
			}
			trust := types.NewTrustAccount("", observer, validatorConsPubKeyStr)
			msg := types.NewMsgSetTrustAccount(trust, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdEndPool
func GetCmdEndPool(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-end-pool [ticker ][requester_bnb_address] [request_txhash]",
		Short: "set end pool",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ticker, err := common.NewTicker(args[0])
			if nil != err {
				return errors.Wrap(err, "invalid ticker")
			}
			requester, err := common.NewAddress(args[1])
			if nil != err {
				return errors.Wrap(err, "invalid requster bnb address")
			}
			txID, err := common.NewTxID(args[2])
			if nil != err {
				return errors.Wrap(err, "invalid tx hash")
			}

			msg := types.NewMsgEndPool(ticker, requester, txID, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
