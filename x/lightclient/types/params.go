package types

import (
	"bytes"
	"time"

	"github.com/cometbft/cometbft/libs/math"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
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
	ProofSpecs:                   commitmenttypes.GetSDKSpecs(),
	AllowUpdateAfterExpiry:       false,
	AllowUpdateAfterMisbehaviour: false,
	// For chains using Cosmos-SDK's default x/upgrade module, the upgrade path is as follows
	UpgradePath: []string{"upgrade", "upgradedIBCState"},
}

// IsCanonicalClientParamsValid checks if the given IBC tendermint client state has the expected canonical client parameters
func IsCanonicalClientParamsValid(clientState *ibctm.ClientState) bool {
	return true
	// TODO: coordinate with Rollapp params and relayer defaults
	/*
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
			if !EqualICS23ProofSpecs(*proofSpec, *ExpectedCanonicalClientParams.ProofSpecs[i]) {
				return false
			}
		}
		for i, path := range clientState.UpgradePath {
			if path != ExpectedCanonicalClientParams.UpgradePath[i] {
				return false
			}
		}
		return true

	*/
}

func EqualICS23ProofSpecs(proofSpecs1, proofSpecs2 ics23.ProofSpec) bool {
	if proofSpecs1.MaxDepth != proofSpecs2.MaxDepth {
		return false
	}
	if proofSpecs1.MinDepth != proofSpecs2.MinDepth {
		return false
	}
	if proofSpecs1.PrehashKeyBeforeComparison != proofSpecs2.PrehashKeyBeforeComparison {
		return false
	}
	if proofSpecs1.LeafSpec.Hash != proofSpecs2.LeafSpec.Hash {
		return false
	}
	if proofSpecs1.LeafSpec.PrehashKey != proofSpecs2.LeafSpec.PrehashKey {
		return false
	}
	if proofSpecs1.LeafSpec.PrehashValue != proofSpecs2.LeafSpec.PrehashValue {
		return false
	}
	if proofSpecs1.LeafSpec.Length != proofSpecs2.LeafSpec.Length {
		return false
	}
	if !bytes.Equal(proofSpecs1.LeafSpec.Prefix, proofSpecs2.LeafSpec.Prefix) {
		return false
	}
	if len(proofSpecs1.InnerSpec.ChildOrder) != len(proofSpecs2.InnerSpec.ChildOrder) {
		return false
	}
	for i, childOrder := range proofSpecs1.InnerSpec.ChildOrder {
		if childOrder != proofSpecs2.InnerSpec.ChildOrder[i] {
			return false
		}
	}
	if proofSpecs1.InnerSpec.ChildSize != proofSpecs2.InnerSpec.ChildSize {
		return false
	}
	if proofSpecs1.InnerSpec.MinPrefixLength != proofSpecs2.InnerSpec.MinPrefixLength {
		return false
	}
	if proofSpecs1.InnerSpec.MaxPrefixLength != proofSpecs2.InnerSpec.MaxPrefixLength {
		return false
	}
	if !bytes.Equal(proofSpecs1.InnerSpec.EmptyChild, proofSpecs2.InnerSpec.EmptyChild) {
		return false
	}
	if proofSpecs1.InnerSpec.Hash != proofSpecs2.InnerSpec.Hash {
		return false
	}

	return true
}
