package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/jpthor/test/x/swapservice/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	swapserviceQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the swapservice module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       utils.ValidateCmd,
	}
	swapserviceQueryCmd.AddCommand(client.GetCommands(
		GetCmdResolvePoolData(storeKey, cdc),
		GetCmdPoolStruct(storeKey, cdc),
		GetCmdPoolDatas(storeKey, cdc),
	)...)
	return swapserviceQueryCmd
}

// GetCmdResolvePoolData queries information about a pooldata
func GetCmdResolvePoolData(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "resolve [pooldata]",
		Short: "resolve pooldata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			pooldata := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/resolve/%s", queryRoute, pooldata), nil)
			if err != nil {
				fmt.Printf("could not resolve pooldata - %s \n", pooldata)
				return nil
			}

			var out types.QueryResResolve
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
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

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/poolstruct/%s", queryRoute, pooldata), nil)
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

// GetCmdPoolDatas queries a list of all pooldatas
func GetCmdPoolDatas(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pooldatas",
		Short: "pooldatas",
		// Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/pooldatas", queryRoute), nil)
			if err != nil {
				fmt.Printf("could not get query pooldatas\n")
				return nil
			}

			var out types.QueryResPoolDatas
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
