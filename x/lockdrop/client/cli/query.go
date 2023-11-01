package cli

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/x/lockdrop/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		GetCmdDistrInfo(),
		osmocli.GetParams[*types.QueryParamsRequest](
			types.ModuleName, types.NewQueryClient),
	)

	return cmd
}

// GetCmdDistrInfo takes the pool id and returns the matching gauge ids and weights.
func GetCmdDistrInfo() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryDistrInfoRequest](
		"distr-info",
		"Query distribution info",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} distr-info
`, types.ModuleName, types.NewQueryClient)
}
