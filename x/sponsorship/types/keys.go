package types

import (
	"cosmossdk.io/collections"
)

// Module name and store keys.
const (
	// ModuleName defines the module name
	ModuleName = "sponsorship"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	RouterKey = ModuleName
)

const (
	ParamsByte                  = iota // Module params: Params
	DistributionByte                   // Current distribution: Distribution
	DelegatorValidatorPowerByte        // Delegator voting power by the validator: math.Int
	VoteByte                           // User's vote: Vote
	RAEndorsementsByte                 // RA endorsement: Endorsement
	RAGaugeIDIndexByte                 // RA endorsement by RA gauge ID index: Endorsement
	ClaimBlacklistIndexByte            // Claim blacklist: sdk.AccAddress
)

func ParamsPrefix() collections.Prefix {
	return collections.NewPrefix(ParamsByte)
}

func DistributionPrefix() collections.Prefix {
	return collections.NewPrefix(DistributionByte)
}

func DelegatorValidatorPrefix() collections.Prefix {
	return collections.NewPrefix(DelegatorValidatorPowerByte)
}

func VotePrefix() collections.Prefix {
	return collections.NewPrefix(VoteByte)
}

func RAEndorsementsPrefix() collections.Prefix {
	return collections.NewPrefix(RAEndorsementsByte)
}

func RAGaugeIDIndexPrefix() collections.Prefix {
	return collections.NewPrefix(RAGaugeIDIndexByte)
}

func ClaimBlacklistPrefix() collections.Prefix {
	return collections.NewPrefix(ClaimBlacklistIndexByte)
}
