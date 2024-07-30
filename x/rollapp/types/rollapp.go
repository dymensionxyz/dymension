package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	lastStateUpdateHeightSentinel = -1 // means value is absent
)

func NewRollapp(creator string, rollappId string, maxSequencers uint64, permissionedAddresses []string, transfersEnabled bool) Rollapp {
	ret := Rollapp{
		RollappId:             rollappId,
		Creator:               creator,
		MaxSequencers:         maxSequencers,
		PermissionedAddresses: permissionedAddresses,
		LastStateUpdateHeight: lastStateUpdateHeightSentinel,
	}
	ret.GenesisState.TransfersEnabled = transfersEnabled
	return ret
}

func (r Rollapp) LastStateUpdateHeightIsSet() bool {
	return r.LastStateUpdateHeight != lastStateUpdateHeightSentinel
}

func (r Rollapp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(r.Creator)
	if err != nil {
		return errorsmod.Wrap(err, ErrInvalidCreatorAddress.Error())
	}

	// validate rollappId
	_, err = NewChainID(r.RollappId)
	if err != nil {
		return err
	}

	if r.MaxSequencers > MaxAllowedSequencers {
		return errorsmod.Wrapf(ErrInvalidMaxSequencers, "max sequencers: %d, max sequencers allowed: %d", r.GetMaxSequencers(), MaxAllowedSequencers)
	}
	if uint64(len(r.PermissionedAddresses)) > r.GetMaxSequencers() {
		return errorsmod.Wrapf(ErrTooManyPermissionedAddresses, "permissioned addresses: %d, max sequencers: %d", len(r.PermissionedAddresses), r.GetMaxSequencers())
	}

	// verifies that there's no duplicate address in PermissionedAddresses
	// and addresses are in Bech32 format
	permissionedAddresses := r.GetPermissionedAddresses()
	if len(permissionedAddresses) > 0 {
		duplicateAddresses := make(map[string]bool)
		for _, item := range permissionedAddresses {
			// check if the item/element exist in the duplicateAddresses map
			_, exist := duplicateAddresses[item]
			if exist {
				return errorsmod.Wrapf(ErrPermissionedAddressesDuplicate, "address: %s", item)
			}
			// check Bech32 format
			if _, err := sdk.AccAddressFromBech32(item); err != nil {
				return errorsmod.Wrapf(ErrInvalidPermissionedAddress, "%s", err)
			}
			// mark as exist
			duplicateAddresses[item] = true
		}
	}

	return nil
}
