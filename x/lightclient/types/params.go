package types

import (
	"time"

	"github.com/cometbft/cometbft/libs/math"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ics23 "github.com/cosmos/ics23/go"
)

// ExpectedCanonicalClientParams defines the expected parameters for a canonical IBC Tendermint client state
// The ChainID is not included as that varies for each rollapp
// The LatestHeight is not included as there is no condition on when a client can be registered as canonical
var ExpectedCanonicalClientParams = ibctm.ClientState{
	// Trust level is the fraction of the trusted validator set
	// that must sign over a new untrusted header before it is accepted.
	// For a rollapp should be 1/1.
	TrustLevel: ibctm.NewFractionFromTm(math.Fraction{Numerator: 1, Denominator: 1}),
	// TrustingPeriod is the duration of the period since the
	// LatestTimestamp during which the submitted headers are valid for update.
	TrustingPeriod: time.Hour * 24 * 7 * 2,
	// Unbonding period is the duration of the sequencer unbonding period.
	UnbondingPeriod: time.Hour * 24 * 7 * 3,
	// MaxClockDrift defines how much new (untrusted) header's Time
	// can drift into the future relative to our local clock.
	MaxClockDrift: time.Minute * 10,
	// Frozen Height should be zero (default) as frozen clients cannot be canonical
	// as they cannot receive state updates
	FrozenHeight: ibcclienttypes.ZeroHeight(),
	// ProofSpecs defines the ICS-23 standard proof specifications used by
	// the light client. It is used configure a proof for either existence
	// or non-existence of a key value pair
	ProofSpecs: []*ics23.ProofSpec{ // the proofspecs for a SDK chain
		ics23.IavlSpec,
		ics23.TendermintSpec,
	},
	AllowUpdateAfterExpiry:       false,
	AllowUpdateAfterMisbehaviour: false,
	// For chains using Cosmos-SDK's default x/upgrade module, the upgrade path is as follows
	UpgradePath: []string{"upgrade", "upgradedIBCState"},
}

// IsCanonicalClientParamsValid checks if the given IBC tendermint client state has the expected canonical client parameters
func IsCanonicalClientParamsValid(clientState *ibctm.ClientState) bool {
	if clientState.TrustLevel != ExpectedCanonicalClientParams.TrustLevel {
		return false
	}
	if clientState.TrustingPeriod != ExpectedCanonicalClientParams.TrustingPeriod {
		return false
	}
	if clientState.UnbondingPeriod != ExpectedCanonicalClientParams.UnbondingPeriod {
		return false
	}
	if clientState.MaxClockDrift != ExpectedCanonicalClientParams.MaxClockDrift {
		return false
	}
	if clientState.FrozenHeight != ExpectedCanonicalClientParams.FrozenHeight {
		return false
	}
	if clientState.AllowUpdateAfterExpiry != ExpectedCanonicalClientParams.AllowUpdateAfterExpiry {
		return false
	}
	if clientState.AllowUpdateAfterMisbehaviour != ExpectedCanonicalClientParams.AllowUpdateAfterMisbehaviour {
		return false
	}
	for i, proofSpec := range clientState.ProofSpecs {
		if proofSpec != ExpectedCanonicalClientParams.ProofSpecs[i] {
			return false
		}
	}
	for i, path := range clientState.UpgradePath {
		if path != ExpectedCanonicalClientParams.UpgradePath[i] {
			return false
		}
	}
	return true
}
