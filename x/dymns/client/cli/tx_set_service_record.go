package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/spf13/cobra"
)

// NewSetServiceRecordTxCmd returns the CLI command for setting a typed
// service/endpoint record on a Dym-Name.
func NewSetServiceRecordTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-service-record [Dym-Name] [service-key] [?value]",
		Short: "Set a typed service/endpoint record on a Dym-Name. Empty value removes the record.",
		Example: fmt.Sprintf(
			"$ %s tx %s set-service-record my-name mcp https://mcp.example.com --%s hub-user",
			version.AppName, dymnstypes.ModuleName, flags.FlagFrom,
		),
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			serviceKey := args[1]
			var value string
			if len(args) > 2 {
				value = args[2]
			}

			controller := clientCtx.GetFromAddress().String()
			if controller == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			msg := &dymnstypes.MsgSetServiceRecord{
				Name:       name,
				Controller: controller,
				ServiceKey: serviceKey,
				Value:      value,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
