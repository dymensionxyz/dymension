package types

import (
	"errors"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		VoterInfos: make([]VoterInfo, 0),
	}
}

func (g GenesisState) Validate() error {
	err := g.Params.Validate()
	if err != nil {
		return errors.Join(ErrInvalidGenesis, err)
	}

	for _, i := range g.VoterInfos {
		err = i.Validate()
		if err != nil {
			return errors.Join(ErrInvalidGenesis, err)
		}
	}

	return nil
}

func (v VoterInfo) Validate() error {
	_, err := sdk.AccAddressFromBech32(v.Voter)
	if err != nil {
		return ErrInvalidVoterInfo.Wrapf(
			"voter '%s' must be a valid bech32 address: %s",
			v.Voter, err.Error(),
		)
	}

	err = v.Vote.Validate()
	if err != nil {
		return ErrInvalidVoterInfo.Wrapf(err.Error())
	}

	// Validate validators
	total := math.ZeroInt()
	validators := make(map[string]struct{}, len(v.Validators)) // this map helps check for duplicates
	for _, val := range v.Validators {
		if _, ok := validators[val.Validator]; ok {
			return ErrInvalidVoterInfo.Wrapf("duplicated validators: %s", val.Validator)
		}
		validators[val.Validator] = struct{}{}
		total = total.Add(val.Power)
	}
	if total.GT(v.Vote.VotingPower) {
		return ErrInvalidVoterInfo.Wrapf("voting power mismatch: vote voting power %s is less than total validator power %s", v.Vote.VotingPower, total)
	}

	return nil
}
