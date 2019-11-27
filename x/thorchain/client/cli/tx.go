package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"gitlab.com/thorchain/thornode/common"

	appCmd "gitlab.com/thorchain/thornode/cmd"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	thorchainTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "thorchain transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	thorchainTxCmd.AddCommand(client.PostCommands(
		GetCmdSetAdminConfig(cdc),
		GetCmdSetTrustAccount(cdc),
		GetCmdEndPool(cdc),
		GetCmdSetVersion(cdc),
	)...)

	return thorchainTxCmd
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

			msg := types.NewMsgSetAdminConfig(common.Tx{}, key, args[1], cliCtx.GetFromAddress())
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
		Use:   "set-trust-account  [secp256k1] [ed25519] [validator_consensus_pub_key]",
		Short: "set trust account, the account use to sign this tx has to be whitelist first",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			secp256k1Key, err := common.NewPubKey(args[0])
			if nil != err {
				return fmt.Errorf("fail to parse secp256k1 pub key ,err:%w", err)
			}
			ed25519Key, err := common.NewPubKey(args[1])
			if nil != err {
				return fmt.Errorf("fail to parse ed25519 pub key ,err:%w", err)
			}
			pk := common.NewPubKeys(secp256k1Key, ed25519Key)
			validatorConsPubKey, err := sdk.GetConsPubKeyBech32(args[2])
			if err != nil {
				return errors.Wrap(err, "fail to parse validator consensus public key")
			}
			validatorConsPubKeyStr, err := sdk.Bech32ifyConsPub(validatorConsPubKey)
			if err != nil {
				return errors.Wrap(err, "fail to convert public key to string")
			}
			msg := types.NewMsgSetTrustAccount(pk, validatorConsPubKeyStr, cliCtx.GetFromAddress())
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
		Use:   "set-end-pool [asset] [requester_bnb_address] [request_txhash]",
		Short: "set end pool",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			asset, err := common.NewAsset(args[0])
			if nil != err {
				return errors.Wrap(err, "invalid asset")
			}
			requester, err := common.NewAddress(args[1])
			if nil != err {
				return errors.Wrap(err, "invalid requster bnb address")
			}
			txID, err := common.NewTxID(args[2])
			if nil != err {
				return errors.Wrap(err, "invalid tx hash")
			}

			tx := common.Tx{
				ID:          txID,
				FromAddress: requester,
			}

			msg := types.NewMsgEndPool(asset, tx, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
