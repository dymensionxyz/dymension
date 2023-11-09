package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	"github.com/dymensionxyz/dymension/x/lockdrop/types"
)

func parseRecords(args []string) ([]types.DistrRecord, error) {
	gaugeIds, err := osmoutils.ParseUint64SliceFromString(args[0], ",")
	if err != nil {
		return nil, err
	}

	weights, err := osmoutils.ParseSdkIntFromString(args[1], ",")
	if err != nil {
		return nil, err
	}

	if len(gaugeIds) != len(weights) {
		return nil, fmt.Errorf("the length of gauge ids and weights not matched")
	}

	if len(gaugeIds) == 0 {
		return nil, fmt.Errorf("records is empty")
	}

	var records []types.DistrRecord
	for i, gaugeId := range gaugeIds {
		records = append(records, types.DistrRecord{
			GaugeId: gaugeId,
			Weight:  weights[i],
		})
	}
	return records, nil
}

func parseProposal(cmd *cobra.Command) (osmoutils.Proposal, sdk.Coins, error) {
	proposal, err := osmoutils.ParseProposalFlags(cmd.Flags())
	if err != nil {
		return osmoutils.Proposal{}, nil, fmt.Errorf("failed to parse proposal: %w", err)
	}

	deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
	if err != nil {
		return osmoutils.Proposal{}, nil, err
	}
	return *proposal, deposit, nil
}

func NewCmdSubmitUpdateLockdropProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-lockdrop [gaugeIds] [weights]",
		Args:    cobra.ExactArgs(2),
		Short:   "Submit an update to the records for pool incentives",
		Example: "update-lockdrop 1,2 40,60 ",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			from := clientCtx.GetFromAddress()

			proposal, deposit, err := parseProposal(cmd)
			if err != nil {
				return err
			}
			records, err := parseRecords(args)
			if err != nil {
				return err
			}
			content := types.NewUpdateLockdropProposal(proposal.Title, proposal.Description, records)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "The proposal title")
	cmd.Flags().String(govcli.FlagDescription, "", "The proposal description")
	cmd.Flags().String(govcli.FlagDeposit, "", "The proposal deposit")
	cmd.Flags().String(govcli.FlagProposal, "", "Proposal file path (if this path is given, other proposal flags are ignored)")

	return cmd
}

func NewCmdSubmitReplaceLockdropProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "replace-lockdrop [gaugeIds] [weights]",
		Args:    cobra.ExactArgs(2),
		Short:   "Submit a full replacement to the records for pool incentives",
		Example: "replace-lockdrop 1,2 40,60 ",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			from := clientCtx.GetFromAddress()

			proposal, deposit, err := parseProposal(cmd)
			if err != nil {
				return err
			}
			records, err := parseRecords(args)
			if err != nil {
				return err
			}
			content := types.NewReplaceLockdropProposal(proposal.Title, proposal.Description, records)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "The proposal title")
	cmd.Flags().String(govcli.FlagDescription, "", "The proposal description")
	cmd.Flags().String(govcli.FlagDeposit, "", "The proposal deposit")
	cmd.Flags().String(govcli.FlagProposal, "", "Proposal file path (if this path is given, other proposal flags are ignored)")

	return cmd
}
