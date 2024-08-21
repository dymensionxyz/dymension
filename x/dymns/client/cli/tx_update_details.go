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

const (
	flagClearConfigs = "clear-configs"
)

// NewUpdateDetailsTxCmd is the CLI command for updating the details of a Dym-Name.
func NewUpdateDetailsTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("update-details [Dym-Name] --%s <new_contacts> [--%s]", flagContact, flagClearConfigs),
		Short: "Configure resolve Dym-Name address. 2nd arg if empty means to remove the configuration.",
		Example: fmt.Sprintf(
			"$ %s tx %s update-details myname --%s contact@example.com --%s hub-user [--%s]",
			version.AppName, dymnstypes.ModuleName, flagContact, flags.FlagFrom, flagClearConfigs,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dymName := args[0]

			controller := clientCtx.GetFromAddress().String()

			if controller == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			contact, err := cmd.Flags().GetString(flagContact)
			if err != nil {
				return err
			}
			clearConfigs, err := cmd.Flags().GetBool(flagClearConfigs)
			if err != nil {
				return err
			}

			msg := &dymnstypes.MsgUpdateDetails{
				Name:         dymName,
				Controller:   controller,
				Contact:      contact,
				ClearConfigs: clearConfigs,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(flagContact, dymnstypes.DoNotModifyDesc, "New contact details for the Dym-Name")
	cmd.Flags().Bool(flagClearConfigs, false, "Clear all the current resolution configurations for the Dym-Name")

	return cmd
}
