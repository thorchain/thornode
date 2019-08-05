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
		GetCmdStakerPoolStruct(storeKey, cdc),
		GetCmdPoolStakerStruct(storeKey, cdc),
		GetCmdPoolIndex(storeKey, cdc),
		GetCmdSwapRecord(storeKey, cdc),
		GetCmdUnStakeRecord(storeKey, cdc),
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

// GetCmdStakerPoolStruct queries staker pool
func GetCmdStakerPoolStruct(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "stakerpool [stakeraddress]",
		Short: "Query staker pool info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			stakerAddress := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/stakerpools/%s", queryRoute, stakerAddress), nil)
			if err != nil {
				cmd.Printf("could not resolve stakerpool - %s \n", stakerAddress)
				return nil
			}

			var out types.StakerPool
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
} // GetCmdPoolStakerStruct queries pool staker
func GetCmdPoolStakerStruct(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "poolstaker [ticker]",
		Short: "Query pool staker info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			ticker := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/poolstakers/%s", queryRoute, ticker), nil)
			if err != nil {
				cmd.Printf("could not resolve poolstaker - %s \n", ticker)
				return nil
			}

			var out types.PoolStaker
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
				cmd.Println("could not get query pooldatas")
				return nil
			}

			var out types.QueryResPoolDatas
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
