package cli

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/spf13/cobra"
)

// TODO: move to utils/cli package
/*func parseRecords(gaugesRaw, weightsRaw string) ([]types.DistrRecord, error) {
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
}*/

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
