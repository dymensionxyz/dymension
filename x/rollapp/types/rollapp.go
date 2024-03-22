package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewRollapp(creator string, rollappId string, maxSequencers uint64, permissionedAddresses []string,
	metadatas []*TokenMetadata, genesisAccounts *RollappGenesisState,
) Rollapp {
	return Rollapp{
		RollappId:             rollappId,
		Creator:               creator,
		MaxSequencers:         maxSequencers,
		PermissionedAddresses: permissionedAddresses,
		GenesisState:          genesisAccounts,
		TokenMetadata:         metadatas,
	}
}

func (r Rollapp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(r.Creator)
	if err != nil {
		return errorsmod.Wrap(err, ErrInvalidCreatorAddress.Error())
	}

	// validate rollappId
	_, _, err = GetValidEIP155ChainId(r.RollappId)
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

	// verifies that token metadata, if any, must be valid
	if len(r.TokenMetadata) > 0 {
		for _, metadata := range r.TokenMetadata {
			if err := metadata.Validate(); err != nil {
				return errorsmod.Wrapf(ErrInvalidTokenMetadata, "%s: %v", metadata.Base, err)
			}
		}
	}

	// genesisAccounts address validation
	if r.GenesisState != nil {
		for _, acc := range r.GenesisState.GenesisAccounts {
			_, err := sdk.AccAddressFromBech32(acc.Address)
			if err != nil {
				return errorsmod.Wrapf(err, "invalid genesis account address (%s)", acc.Address)
			}
		}
	}

	return nil
}
