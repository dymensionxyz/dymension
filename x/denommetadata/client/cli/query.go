package cli

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
)

// GetQueryCmd returns the query commands for this module.
func GetQueryCmd() *cobra.Command {
	// group metadata queries under a subcommand
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

// GetCmdDenomMetadataByID returns a denonmetadata by ID.
func GetCmdAllDenomMetadata() (*osmocli.QueryDescriptor, *types.DenomMetadataByDisplayDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "denonmetadata",
		Short: "Query all denonmetadata.",
		Long: `{{.Short}}{{.ExampleHeader}}
`}, &types.DenomMetadataByDisplayDenomRequest{}
}

// GetCmdDenomMetadataByID returns a denonmetadata by ID.
func GetCmdDenomMetadataByDisplayDenom() (*osmocli.QueryDescriptor, *types.DenomMetadataByDisplayDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "denonmetadata-by-display [display]",
		Short: "Query denonmetadata by display.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} denonmetadata-by-display 1
`}, &types.DenomMetadataByDisplayDenomRequest{}
}

// GetCmdDenomMetadataByID returns a denonmetadata by ID.
func GetCmdDenomMetadataByBaseDenom() (*osmocli.QueryDescriptor, *types.DenomMetadataByBaseDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "denonmetadata-by-base [base]",
		Short: "Query denonmetadata by base.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} denonmetadata-by-base 1
`}, &types.DenomMetadataByBaseDenomRequest{}
}

// GetCmdDenomMetadataByID returns a denonmetadata by ID.
func GetCmdDenomMetadataBySymbolDenom() (*osmocli.QueryDescriptor, *types.DenomMetadataBySymbolDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "denonmetadata-by-symbol [symbol]",
		Short: "Query denonmetadata by symbol.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} denonmetadata-by-symbol 1
`}, &types.DenomMetadataBySymbolDenomRequest{}
}
