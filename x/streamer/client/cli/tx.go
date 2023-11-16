package cli

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/x/streamer/types"

	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

// TODO: move to utils/cli package
func parseRecords(gaugesRaw, weightsRaw string) ([]types.DistrRecord, error) {
	gaugeIds, err := osmoutils.ParseUint64SliceFromString(gaugesRaw, ",")
	if err != nil {
		return nil, err
	}

	weights, err := osmoutils.ParseSdkIntFromString(weightsRaw, ",")
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

// NewCreateStreamCmd broadcasts a CreateStream message.
func NewCmdSubmitCreateStreamProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-stream-proposal gaugeIds weights reward [flags]",
		Short: "proposal to create a stream of incentives rewards over a period of time",
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
			records, err := parseRecords(args[0], args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			var startTime time.Time
			timeStr, err := cmd.Flags().GetString(FlagStartTime)
			if err != nil {
				return err
			}
			if timeStr == "" { // empty start time
				startTime = time.Unix(0, 0)
			} else if timeUnix, err := strconv.ParseInt(timeStr, 10, 64); err == nil { // unix time
				startTime = time.Unix(timeUnix, 0)
			} else if timeRFC, err := time.Parse(time.RFC3339, timeStr); err == nil { // RFC time
				startTime = timeRFC
			} else { // invalid input
				return errors.New("invalid start time format")
			}

			epochIdentifier, err := cmd.Flags().GetString(FlagEpochIdentifier)
			if err != nil {
				return err
			}

			epochs, err := cmd.Flags().GetUint64(FlagEpochs)
			if err != nil {
				return err
			}

			content := types.NewCreateStreamProposal(proposal.Title, proposal.Description, coins, records, startTime, epochIdentifier, epochs)
			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "The proposal title")
	cmd.Flags().String(govcli.FlagDescription, "", "The proposal description")
	cmd.Flags().String(govcli.FlagDeposit, "", "The proposal deposit")

	cmd.Flags().AddFlagSet(FlagSetCreateStream())
	return cmd
}
