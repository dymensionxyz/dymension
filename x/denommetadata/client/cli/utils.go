package cli

import (
	"fmt"
	"strconv"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/spf13/cobra"
)

func parseRecords(inputdenom, inputdecimals string) (banktypes.Metadata, error) {

	if len(inputdenom) < 2 || len(inputdenom) > 10 {
		return banktypes.Metadata{}, fmt.Errorf("the length of denom is not correct")
	}

	decimals, err := strconv.ParseUint(inputdecimals, 10, 64)
	if err != nil {
		return banktypes.Metadata{}, err
	}
	record := banktypes.Metadata{
		Description: ,
		DenomUnits:,
		Base:,
		Display:,
		Name:
		Symbol:
		Uri:
		UriHash:
	}
	return record, nil
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
