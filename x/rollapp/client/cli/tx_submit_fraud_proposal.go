package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// NewCmdSubmitFraudProposal submits a fraud proposal
func NewCmdSubmitFraudProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-fraud-proposal <rollappID> <height> <propser_addr> <client_id>",
		Short: "submit a fraud proposal",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, deposit, err := utils.ParseProposal(cmd)
			if err != nil {
				return err
			}

			rollappID := args[0]
			height, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			proposerAddr := args[2]
			ibcClientID := args[3]

			content := types.NewSubmitFraudProposal(proposal.Title, proposal.Description, rollappID, height, proposerAddr, ibcClientID)
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
