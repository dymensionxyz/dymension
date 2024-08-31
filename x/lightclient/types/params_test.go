package types_test

import (
	"testing"

	"github.com/cometbft/cometbft/libs/math"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ics23 "github.com/cosmos/ics23/go"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

func TestIsCanonicalClientParamsValid(t *testing.T) {
	t.Skip("disabled temporarily - need to bring back")
	testCases := []struct {
		name        string
		clientState func() ibctm.ClientState
		valid       bool
	}{
		{
			"valid client state",
			func() ibctm.ClientState {
				return types.ExpectedCanonicalClientParams
			},
			true,
		},
		{
			"invalid trust level",
			func() ibctm.ClientState {
				clientState := types.ExpectedCanonicalClientParams
				clientState.TrustLevel = ibctm.NewFractionFromTm(math.Fraction{Numerator: 1, Denominator: 2})
				return clientState
			},
			false,
		},
		{
			"invalid trusting period",
			func() ibctm.ClientState {
				clientState := types.ExpectedCanonicalClientParams
				clientState.TrustingPeriod = clientState.TrustingPeriod + 1
				return clientState
			},
			false,
		},
		{
			"invalid unbonding period",
			func() ibctm.ClientState {
				clientState := types.ExpectedCanonicalClientParams
				clientState.UnbondingPeriod = clientState.UnbondingPeriod + 1
				return clientState
			},
			false,
		},
		{
			"invalid max clock drift",
			func() ibctm.ClientState {
				clientState := types.ExpectedCanonicalClientParams
				clientState.MaxClockDrift = clientState.MaxClockDrift + 1
				return clientState
			},
			false,
		},
		{
			"invalid frozen height",
			func() ibctm.ClientState {
				clientState := types.ExpectedCanonicalClientParams
				clientState.FrozenHeight = ibcclienttypes.NewHeight(1, 1)
				return clientState
			},
			false,
		},
		{
			"invalid allow update after expiry",
			func() ibctm.ClientState {
				clientState := types.ExpectedCanonicalClientParams
				clientState.AllowUpdateAfterExpiry = true
				return clientState
			},
			false,
		},
		{
			"invalid allow update after misbehaviour",
			func() ibctm.ClientState {
				clientState := types.ExpectedCanonicalClientParams
				clientState.AllowUpdateAfterMisbehaviour = true
				return clientState
			},
			false,
		},
		{
			"invalid proof specs",
			func() ibctm.ClientState {
				clientState := types.ExpectedCanonicalClientParams
				clientState.ProofSpecs = []*ics23.ProofSpec{ics23.SmtSpec}
				return clientState
			},
			false,
		},
		{
			"invalid upgrade path",
			func() ibctm.ClientState {
				clientState := types.ExpectedCanonicalClientParams
				clientState.UpgradePath = []string{"custom", "upgrade"}
				return clientState
			},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clientState := tc.clientState()
			valid := types.IsCanonicalClientParamsValid(&clientState)
			if valid != tc.valid {
				t.Errorf("expected valid: %v, got: %v", tc.valid, valid)
			}
		})
	}
}
