package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	swapserviceQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the swapservice module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	swapserviceQueryCmd.AddCommand(client.GetCommands(
		GetCmdPoolStruct(storeKey, cdc),
		GetCmdPoolStructs(storeKey, cdc),
		GetCmdPoolIndex(storeKey, cdc),
		GetCmdSwapRecord(storeKey, cdc),
		GetCmdUnStakeRecord(storeKey, cdc),
		GetCmdTxOutArray(storeKey, cdc),
		GetCmdGetAdminConfig(storeKey, cdc),
	)...)
	return swapserviceQueryCmd
}

// GetCmdPoolStruct queries information about a domain
func GetCmdPoolStruct(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "poolstruct [pooldata]",
		Short: "Query poolstruct info of pooldata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			pooldata := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/poolstruct/%s", queryRoute, pooldata), nil)
			if err != nil {
				fmt.Printf("could not resolve poolstruct - %s \n", pooldata)
				return nil
			}

			var out types.PoolStruct
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdPoolStructs queries a list of all pool data
func GetCmdPoolStructs(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pools",
		Short: "pools",
		// Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/pools", queryRoute), nil)
			if err != nil {
				cmd.Println("could not get query pools", err)
				return nil
			}

			var out types.QueryResPoolStructs
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdPoolIndex query pool index
func GetCmdPoolIndex(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "poolindex",
		Short: "poolindex",
		// Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/poolindex", queryRoute), nil)
			if err != nil {
				cmd.Println("could not get query poolindex")
				return nil
			}

			var out types.PoolIndex
			cdc.MustUnmarshalJSON(res, &out)
			cmd.Println(out)
			return nil
		},
	}
}

// GetSwapRecord query a swap record
func GetCmdSwapRecord(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "swaprecord [requestTxHash]",
		Short: "swaprecord",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			requestTxHash := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/swaprecord/%s", queryRoute, requestTxHash), nil)
			if err != nil {
				cmd.Println("could not get query swaprecord")
				return nil
			}
			cmd.Println(string(res))
			return nil
		},
	}
}

// GetCmdUnStakeRecord query a swap record
func GetCmdUnStakeRecord(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unstakerecord [requestTxHash]",
		Short: "unstakerecord",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			requestTxHash := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/unstakerecord/%s", queryRoute, requestTxHash), nil)
			if err != nil {
				cmd.Println("could not get query unstake")
				return nil
			}
			cmd.Println(string(res))
			return nil
		},
	}
}

// GetCmdTxOutArray query txoutarray
func GetCmdTxOutArray(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "txout [height]",
		Short: "txout array",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			height := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/txoutarray/%s", queryRoute, height), nil)
			if err != nil {
				cmd.Println("could not get query txoutarray")
				return nil
			}
			cmd.Println(string(res))
			return nil
		},
	}
}

// GetCmdGetAdminConfig query a swap record
func GetCmdGetAdminConfig(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-admin-config [key]",
		Short: "admin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			key := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/adminconfig/%s", queryRoute, key), nil)
			if err != nil {
				cmd.Println("could not get query unstake")
				return nil
			}
			cmd.Println(string(res))
			return nil
		},
	}
}
