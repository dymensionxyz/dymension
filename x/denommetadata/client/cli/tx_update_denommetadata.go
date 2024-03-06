package cli

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// NewCreateDenomMetadataCmd broadcasts a CreateMetadataProposal message.
func NewCmdSubmitUpdateDenomMetadataProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-denometadata-proposal denommetadataID metadata.json [flags]",
		Short: "proposal to update new denom metadata for a specific token",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, deposit, err := parseProposal(cmd)
			if err != nil {
				return err
			}
			denomMetadataID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			path := args[0]

			fileContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			metadata := types.TokenMetadata{}
			err = json.Unmarshal([]byte(fileContent), &metadata)
			if err != nil {
				return err
			}
			err = metadata.Validate()
			if err != nil {
				return err
			}

			content := types.NewUpdateDenomMetadataProposal(proposal.Title, proposal.Description, denomMetadataID, metadata)
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
