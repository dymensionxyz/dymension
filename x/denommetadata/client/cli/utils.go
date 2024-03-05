package cli

import (
	"fmt"
	"strconv"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/spf13/cobra"
)

func parseRecords(description, denomstring, denomexponent, denomalias, base, display, name, symbol, uri, urihash string) (*types.TokenMetadata, error) {

	exponent, err := strconv.Atoi(denomexponent)
	if err != nil {
		return &types.TokenMetadata{}, err
	}
	record := types.NewTokenMetadata(description, denomstring, uint32(exponent), denomalias, base, display, name, symbol, uri, urihash)
	err = record.Validate()
	if err != nil {
		return &types.TokenMetadata{}, err
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
