package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"

	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// NewCmdSubmitUpdateStreamDistributionProposal broadcasts a CreateStream message.
func NewCmdSubmitUpdateStreamDistributionProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-stream-distribution-proposal streamID gaugeIds weights [flags]",
		Short: "Submit an update to the distribution records of an existing stream",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, deposit, err := utils.ParseProposal(cmd)
			if err != nil {
				return err
			}
			streamID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			records, err := parseRecords(args[1], args[2])
			if err != nil {
				return err
			}

			content := types.NewUpdateStreamDistributionProposal(proposal.Title, proposal.Description, streamID, records)
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
