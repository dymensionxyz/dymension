package types

import errorsmod "cosmossdk.io/errors"

func (bds BlockDescriptors) Validate() error {
	for _, bd := range bds.BD {
		if err := bd.Validate(); err != nil {
			return errorsmod.Wrap(err, "block descriptor validate")
		}
	}
	return nil
}

func (bd BlockDescriptor) Validate() error {
	if bd.Timestamp.IsZero() {
		return ErrInvalidBlockDescriptorTimestamp
	}
	return nil
}
