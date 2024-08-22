package types

import (
	"time"

	"github.com/cometbft/cometbft/libs/math"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ics23 "github.com/cosmos/ics23/go"
)

type CanonicalClientParams struct {
	// Trust level is the fraction of the trusted validator set
	// that must sign over a new untrusted header before it is accepted.
	// For a rollapp should be 1/1.
	TrustLevel ibctm.Fraction
	// TrustingPeriod is the duration of the period since the
	// LatestTimestamp during which the submitted headers are valid for update.
	TrustingPeriod time.Duration
	// Unbonding period is the duration of the sequencer unbonding period.
	UnbondingPeriod time.Duration
	// MaxClockDrift defines how much new (untrusted) header's Time
	//  can drift into the future relative to our local clock.
	MaxClockDrift time.Duration
	// ProofSpecs defines the ICS-23 standard proof specifications used by
	// the light client. It is used configure a proof for either existence
	// or non-existence of a key value pair
	ProofSpecs                   []*ics23.ProofSpec
	AllowUpdateAfterExpiry       bool
	AllowUpdateAfterMisbehaviour bool
}

var ExpectedCanonicalClientParams = CanonicalClientParams{
	TrustLevel:      ibctm.NewFractionFromTm(math.Fraction{Numerator: 1, Denominator: 1}),
	TrustingPeriod:  time.Hour * 24 * 7 * 2,
	UnbondingPeriod: time.Hour * 24 * 7 * 3,
	MaxClockDrift:   time.Minute * 10,
	ProofSpecs: []*ics23.ProofSpec{ // the proofspecs for a SDK chain
		ics23.IavlSpec,
		ics23.TendermintSpec,
	},
	AllowUpdateAfterExpiry:       false,
	AllowUpdateAfterMisbehaviour: false,
}

func IsCanonicalClientParamsValid(clientState *ibctm.ClientState) bool {
	if clientState.TrustLevel != ExpectedCanonicalClientParams.TrustLevel {
		return false
	}
	if clientState.TrustingPeriod > ExpectedCanonicalClientParams.TrustingPeriod {
		return false
	}
	if clientState.UnbondingPeriod > ExpectedCanonicalClientParams.UnbondingPeriod {
		return false
	}
	if clientState.MaxClockDrift > ExpectedCanonicalClientParams.MaxClockDrift {
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
	return true
}
