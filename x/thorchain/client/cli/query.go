package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"gitlab.com/thorchain/thornode/constants"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type ver struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
}

func (v ver) String() string {
	return v.Version
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
	)...)
	return thorchainQueryCmd
}

// GetCmdGetVersion queries current version
func GetCmdGetVersion(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Gets the thorchain version and build information",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			cliCtx.OutputFormat = "json"

			out := ver{
				Version:   constants.SWVersion.String(),
				GitCommit: constants.GitCommit,
				BuildTime: constants.BuildTime,
			}
			return cliCtx.PrintOutput(out)
		},
	}
}
