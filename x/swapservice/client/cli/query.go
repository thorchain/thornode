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
		GetCmdPoolDatas(storeKey, cdc),
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

// GetCmdPoolDatas queries a list of all pooldatas
func GetCmdPoolDatas(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pooldatas",
		Short: "pooldatas",
		// Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/pooldatas", queryRoute), nil)
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
