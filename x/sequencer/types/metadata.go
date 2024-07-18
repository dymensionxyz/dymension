package types

import errorsmod "cosmossdk.io/errors"

// constant for maximum string length of the SequencerMetadata fields
const (
	MaxMonikerLength      = 70
	MaxContactFieldLength = 140
	MaxDetailsLength      = 280
	MaxExtraDataLength    = 280
)

// DoNotModifyDesc constant is used in flags to indicate that the metadata field should not be updated
const DoNotModifyDesc = "[do-not-modify]"

// UpdateSequencerMetadata updates the fields of a given metadata. An error is
// returned if the resulting metadata contains an invalid length.
func (d SequencerMetadata) UpdateSequencerMetadata(d2 SequencerMetadata) (SequencerMetadata, error) {
	metadata := SequencerMetadata{
		Moniker:        d2.Moniker,
		Details:        d2.Details,
		P2PSeeds:       d2.P2PSeeds,
		Rpcs:           d2.Rpcs,
		EvmRpcs:        d2.EvmRpcs,
		RestApiUrl:     d2.RestApiUrl,
		ExplorerUrl:    d2.ExplorerUrl,
		GenesisUrls:    d2.GenesisUrls,
		ContactDetails: d2.ContactDetails,
		ExtraData:      d2.ExtraData,
		Snapshots:      d2.Snapshots,
		GasPrice:       d2.GasPrice,
	}

	if d.Moniker != DoNotModifyDesc {
		metadata.Moniker = d2.Moniker
	}

	if d.Details != DoNotModifyDesc {
		metadata.Details = d2.Details
	}

	if string(d.ExtraData) != DoNotModifyDesc {
		metadata.ExtraData = d2.ExtraData
	}

	if d.ContactDetails != nil {
		metadata.ContactDetails = &ContactDetails{}

		if d.ContactDetails.Website != DoNotModifyDesc {
			metadata.ContactDetails.Website = d2.ContactDetails.Website
		}
		if d.ContactDetails.Telegram != DoNotModifyDesc {
			metadata.ContactDetails.Telegram = d2.ContactDetails.Telegram
		}
		if d.ContactDetails.X != DoNotModifyDesc {
			metadata.ContactDetails.X = d2.ContactDetails.X
		}
	}

	return metadata.EnsureLength()
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
