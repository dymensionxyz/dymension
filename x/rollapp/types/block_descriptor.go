package types

import errorsmod "cosmossdk.io/errors"

func (bds BlockDescriptors) Validate() error {
	for _, bd := range bds.BD {
		if err := bd.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (bd BlockDescriptor) Validate() error {
	if bd.Timestamp.IsZero() {
		return errorsmod.Wrapf(ErrInvalidBlockDescriptorTimestamp, "timestamp is empty for block descriptor at height %d", bd.Height)
	}
	return nil
}
