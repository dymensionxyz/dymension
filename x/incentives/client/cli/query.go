package cli

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
)

// GetQueryCmd returns the query commands for this module.
func GetQueryCmd() *cobra.Command {
	// group incentives queries under a subcommand
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	qcGetter := types.NewQueryClient
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdGauges)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdToDistributeCoins)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdGaugeByID)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdActiveGauges)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdActiveGaugesPerDenom)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdUpcomingGauges)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdUpcomingGaugesPerDenom)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdRollappGauges)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdParams)

	return cmd
}

// GetCmdGauges returns all available gauges.
func GetCmdGauges() (*osmocli.QueryDescriptor, *types.GaugesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "gauges",
		Short: "Query all available gauges",
		Long:  "{{.Short}}",
	}, &types.GaugesRequest{}
}

// GetCmdRollappGauges returns all available rollapp gauges.
func GetCmdRollappGauges() (*osmocli.QueryDescriptor, *types.GaugesRequest) {
	return &osmocli.QueryDescriptor{
		QueryFnName: "RollappGauges",
		Use:         "rollapp-gauges",
		Short:       "Query all available rollapp gauges",
		Long:        "{{.Short}}",
	}, &types.GaugesRequest{}
}

// GetCmdParams returns the current parameters of the module.
func GetCmdParams() (*osmocli.QueryDescriptor, *types.ParamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "params",
		Short: "Query the current parameters of the module",
		Long:  "{{.Short}}",
	}, &types.ParamsRequest{}
}

// GetCmdToDistributeCoins returns coins that are going to be distributed.
func GetCmdToDistributeCoins() (*osmocli.QueryDescriptor, *types.ModuleToDistributeCoinsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "to-distribute-coins",
		Short: "Query coins that is going to be distributed",
		Long:  `{{.Short}}`,
	}, &types.ModuleToDistributeCoinsRequest{}
}

// GetCmdGaugeByID returns a gauge by ID.
func GetCmdGaugeByID() (*osmocli.QueryDescriptor, *types.GaugeByIDRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "gauge-by-id [id]",
		Short: "Query gauge by id.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} gauge-by-id 1
`,
	}, &types.GaugeByIDRequest{}
}

// GetCmdActiveGauges returns active gauges.
func GetCmdActiveGauges() (*osmocli.QueryDescriptor, *types.ActiveGaugesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "active-gauges",
		Short: "Query active gauges",
		Long:  `{{.Short}}`,
	}, &types.ActiveGaugesRequest{}
}

// GetCmdActiveGaugesPerDenom returns active gauges for a specified denom.
func GetCmdActiveGaugesPerDenom() (*osmocli.QueryDescriptor, *types.ActiveGaugesPerDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "active-gauges-per-den [den]denom [denom]",
		Short: "Query active gauges per denom",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} active-gauges-per-denom gamm/pool/1`,
	}, &types.ActiveGaugesPerDenomRequest{}
}

// GetCmdUpcomingGauges returns scheduled gauges.
func GetCmdUpcomingGauges() (*osmocli.QueryDescriptor, *types.UpcomingGaugesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "upcoming-gauges",
		Short: "Query upcoming gauges",
		Long:  `{{.Short}}`,
	}, &types.UpcomingGaugesRequest{}
}

// GetCmdUpcomingGaugesPerDenom returns scheduled gauges for specified denom..
func GetCmdUpcomingGaugesPerDenom() (*osmocli.QueryDescriptor, *types.UpcomingGaugesPerDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "upcoming-gauges-per-denom [denom]",
		Short: "Query scheduled gauges per denom",
		Long:  `{{.Short}}`,
	}, &types.UpcomingGaugesPerDenomRequest{}
}
