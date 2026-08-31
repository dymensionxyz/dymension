package types_test

import (
	"testing"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func TestErrorsUseGerrcCategories(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		category error
	}{
		{name: "invalid gauge weight", err: types.ErrInvalidGaugeWeight, category: gerrc.ErrInvalidArgument},
		{name: "invalid distribution", err: types.ErrInvalidDistribution, category: gerrc.ErrInvalidArgument},
		{name: "invalid params", err: types.ErrInvalidParams, category: gerrc.ErrInvalidArgument},
		{name: "invalid genesis", err: types.ErrInvalidGenesis, category: gerrc.ErrInvalidArgument},
		{name: "invalid vote", err: types.ErrInvalidVote, category: gerrc.ErrInvalidArgument},
		{name: "invalid voter info", err: types.ErrInvalidVoterInfo, category: gerrc.ErrInvalidArgument},
		{name: "no endorsers", err: types.ErrNoEndorsers, category: gerrc.ErrFailedPrecondition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorIs(t, tt.err, tt.category)
		})
	}
}
