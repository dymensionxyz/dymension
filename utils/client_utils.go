package utils

import (
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/spf13/cobra"
)

func ParseProposal(cmd *cobra.Command) (osmoutils.Proposal, sdk.Coins, error) {
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

// ParseJsonFromFile parses a json file into a slice of type T
func ParseJsonFromFile[T any](path string) ([]T, error) {
	var result []T

	// #nosec G304
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
