package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// NewCreateDenomMetadataCmd broadcasts a CreateMetadataProposal message.
func NewCmdSubmitCreateDenomMetadataProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-denometadata-proposal description denomunit-denom denomunit-exponent denomunit-alias base display name symbol uri urihash [flags]",
		Short: "proposal to create new denometa data for a specific token",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, deposit, err := parseProposal(cmd)
			if err != nil {
				return err
			}
			record, err := parseRecords(args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9])
			if err != nil {
				return err
			}

			content := types.NewCreateMetadataProposal(proposal.Title, proposal.Description, record)
			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "The proposal title")
	cmd.Flags().String(govcli.FlagDescription, "", "The proposal description")
	cmd.Flags().String(govcli.FlagDeposit, "", "The proposal deposit")

	return cmd
}
