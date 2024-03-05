package cli

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
)

// GetQueryCmd returns the query commands for this module.
func GetQueryCmd() *cobra.Command {
	// group streamer queries under a subcommand
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	qcGetter := types.NewQueryClient
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdDenomMetadataByID)
	return cmd
}

// GetCmdDenomMetadataByID returns a denonmetadata by ID.
func GetCmdDenomMetadataByID() (*osmocli.QueryDescriptor, *types.DenomMetadataByIDRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "denonmetadata-by-id [id]",
		Short: "Query denonmetadata by id.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} denonmetadata-by-id 1
`}, &types.DenomMetadataByIDRequest{}
}
