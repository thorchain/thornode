package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	appCmd "gitlab.com/thorchain/bepswap/thornode/cmd"
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/types"
)

type ver struct {
	Version int `json:"version"`
}

func (v ver) String() string {
	return fmt.Sprintf("%d", v.Version)
}

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	thorchainQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the thorchain module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	thorchainQueryCmd.AddCommand(client.GetCommands(
		GetCmdGetVersion(storeKey, cdc),
		GetCmdPool(storeKey, cdc),
		GetCmdPools(storeKey, cdc),
		GetCmdStakerPool(storeKey, cdc),
		GetCmdPoolStaker(storeKey, cdc),
		GetCmdPoolIndex(storeKey, cdc),
		GetCmdSwapRecord(storeKey, cdc),
		GetCmdUnStakeRecord(storeKey, cdc),
		GetCmdTxOutArray(storeKey, cdc),
		GetCmdGetAdminConfig(storeKey, cdc),
	)...)
	return thorchainQueryCmd
}

// GetCmdGetVersion queries current version
func GetCmdGetVersion(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Gets the statechain version",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			out := ver{appCmd.Version}
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdPool queries information about a domain
func GetCmdPool(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pool [pooldata]",
		Short: "Query pool info of pooldata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			pooldata := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/pool/%s", queryRoute, pooldata), nil)
			if err != nil {
				fmt.Printf("could not resolve pool - %s \n", pooldata)
				return nil
			}

			var out types.Pool
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdStakerPool queries staker pool
func GetCmdStakerPool(queryRoute string, cdc *codec.Codec) *cobra.Command {
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
} // GetCmdPoolStaker queries pool staker
func GetCmdPoolStaker(queryRoute string, cdc *codec.Codec) *cobra.Command {
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

// GetCmdPools queries a list of all pool data
func GetCmdPools(queryRoute string, cdc *codec.Codec) *cobra.Command {
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

			var out types.QueryResPools
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
