package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(NewSetCanonicalClientTxCmd())

	return cmd
}

func NewSetCanonicalClientTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-canonical-client [client-id]",
		Short:   "Try and set the canonical client for a rollapp",
		Example: "dymd tx lightclient set-canonical-client <client-id>",
		Long:    `Try and set the canonical client for a rollapp.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientId := args[0]

			msg := &types.MsgSetCanonicalClient{
				Signer:   clientCtx.GetFromAddress().String(),
				ClientId: clientId,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
