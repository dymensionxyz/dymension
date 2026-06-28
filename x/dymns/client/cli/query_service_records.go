package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

// CmdQueryServiceRecords is the CLI command for querying the typed
// service/endpoint records of a Dym-Name.
func CmdQueryServiceRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service-records [Dym-Name]",
		Aliases: []string{"services"},
		Short:   "Get the typed service/endpoint records of a Dym-Name",
		Example: fmt.Sprintf("%s q %s service-records myname", version.AppName, dymnstypes.ModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !dymnsutils.IsValidDymName(name) {
				return fmt.Errorf("input is not a valid Dym-Name: %s", name)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := dymnstypes.NewQueryClient(clientCtx)

			res, err := queryClient.DymNameServices(cmd.Context(), &dymnstypes.QueryDymNameServicesRequest{
				Name: name,
			})
			if err != nil {
				return fmt.Errorf("failed to fetch service records of '%s': %w", name, err)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
