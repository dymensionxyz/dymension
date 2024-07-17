package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// NewCmdSubmitUpdateDenomMetadataProposal broadcasts a CreateMetadataProposal message.
func NewCmdSubmitUpdateDenomMetadataProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-denom-metadata-proposal denom_metadata.json [flags]",
		Short:   "proposal to update new denom metadata for a specific token",
		Example: `dymd tx gov submit-legacy-proposal update-denom-metadata-proposal denom_metadata.json`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, deposit, err := utils.ParseProposal(cmd)
			if err != nil {
				return err
			}

			path := args[0]

			var metadatas []banktypes.Metadata
			metadatas, err = utils.ParseJsonFromFile[banktypes.Metadata](path)
			if err != nil {
				return err
			}

			for _, metadata := range metadatas {
				err = metadata.Validate()
				if err != nil {
					return err
				}
			}

			content := types.NewUpdateDenomMetadataProposal(proposal.Title, proposal.Description, metadatas)
			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			txfCli, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf := txfCli.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "The proposal title")
	cmd.Flags().String(govcli.FlagDescription, "", "The proposal description")
	cmd.Flags().String(govcli.FlagDeposit, "", "The proposal deposit")

	return cmd
}
