package types

import (
	"fmt"
	"net/url"

	errorsmod "cosmossdk.io/errors"
)

// constant for maximum string length of the SequencerMetadata fields
const (
	MaxMonikerLength      = 70
	MaxContactFieldLength = 140
	MaxDetailsLength      = 280
	MaxExtraDataLength    = 280
	maxURLLength          = 256
)

// UpdateSequencerMetadata updates the fields of a given metadata. An error is
// returned if the resulting metadata contains an invalid length.
// TODO: add more length checks
func (d SequencerMetadata) UpdateSequencerMetadata(d2 SequencerMetadata) (SequencerMetadata, error) {
	metadata := SequencerMetadata{
		Moniker:        d2.Moniker,
		Details:        d2.Details,
		P2PSeeds:       d2.P2PSeeds,
		Rpcs:           d2.Rpcs,
		EvmRpcs:        d2.EvmRpcs,
		RestApiUrls:    d2.RestApiUrls,
		ExplorerUrl:    d2.ExplorerUrl,
		GenesisUrls:    d2.GenesisUrls,
		ContactDetails: d2.ContactDetails,
		ExtraData:      d2.ExtraData,
		Snapshots:      d2.Snapshots,
		GasPrice:       d2.GasPrice,
	}

	return metadata.EnsureLength()
}

func (d SequencerMetadata) Validate(isEVM bool) error {
	_, err := d.EnsureLength()
	if err != nil {
		return err
	}

	if err = d.validateRPCs(); err != nil {
		return err
	}

	if isEVM {
		if err = d.ValidateEVMRPCs(); err != nil {
			return err
		}
	}

	if d.ContactDetails == nil {
		return nil
	}

	return d.ContactDetails.Validate()
}

func (cd ContactDetails) Validate() error {
	if err := validateURL(cd.Website); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, "invalid website URL")
	}

	if err := validateURL(cd.Telegram); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, "invalid telegram URL")
	}

	if err := validateURL(cd.X); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, "invalid x URL")
	}

	return nil
}

// ValidateRPCs validates the RPCs of a sequencer's metadata.
func (d SequencerMetadata) validateRPCs() error {
	if len(d.Rpcs) == 0 {
		return errorsmod.Wrap(ErrInvalidRequest, "rpcs cannot be empty")
	}

	for _, rpc := range d.Rpcs {
		if rpc == "" {
			return errorsmod.Wrap(ErrInvalidRequest, "rpc cannot be empty")
		}
		if err := validateURL(rpc); err != nil {
			return errorsmod.Wrap(ErrInvalidRequest, err.Error())
		}
	}

	return nil
}

// ValidateEVMRPCs validates the EVM RPCs of a sequencer's metadata.
// The EVM RPCs are not validated during ValidateBasic, as they are Rollapp-evm specific,
// so they will be validated in the handler.
func (d SequencerMetadata) ValidateEVMRPCs() error {
	if len(d.EvmRpcs) == 0 {
		return errorsmod.Wrap(ErrInvalidRequest, "evm rpcs cannot be empty")
	}

	for _, rpc := range d.Rpcs {
		if rpc == "" {
			return errorsmod.Wrap(ErrInvalidRequest, "evm rpc cannot be empty")
		}
		if err := validateURL(rpc); err != nil {
			return errorsmod.Wrap(ErrInvalidRequest, err.Error())
		}
	}

	return nil
}

func validateURL(urlStr string) error {
	if urlStr == "" {
		return nil
	}

	if len(urlStr) > maxURLLength {
		return fmt.Errorf("URL exceeds maximum length")
	}

	if _, err := url.ParseRequestURI(urlStr); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	return nil
}

// EnsureLength ensures the length of a sequencer's metadata.
func (d SequencerMetadata) EnsureLength() (SequencerMetadata, error) {
	if len(d.Moniker) > MaxMonikerLength {
		return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid moniker length; got: %d, max: %d", len(d.Moniker), MaxMonikerLength)
	}

	if len(d.Details) > MaxDetailsLength {
		return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid details length; got: %d, max: %d", len(d.Details), MaxDetailsLength)
	}

	if len(d.ExtraData) > MaxExtraDataLength {
		return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid extra data length; got: %d, max: %d", len(d.ExtraData), MaxExtraDataLength)
	}

	if cd := d.ContactDetails; cd != nil {
		if len(cd.Website) > MaxContactFieldLength {
			return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid website length; got: %d, max: %d", len(cd.Website), MaxContactFieldLength)
		}
		if len(cd.Telegram) > MaxContactFieldLength {
			return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid telegram length; got: %d, max: %d", len(cd.Telegram), MaxContactFieldLength)
		}
		if len(cd.X) > MaxContactFieldLength {
			return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid x length; got: %d, max: %d", len(cd.X), MaxContactFieldLength)
		}
	}

	return d, nil
}
