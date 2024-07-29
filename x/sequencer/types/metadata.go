package types

import errorsmod "cosmossdk.io/errors"

// constant for maximum string length of the SequencerMetadata fields
const (
	MaxMonikerLength      = 70
	MaxContactFieldLength = 140
	MaxDetailsLength      = 280
	MaxExtraDataLength    = 280
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
		RestApiUrl:     d2.RestApiUrl,
		ExplorerUrl:    d2.ExplorerUrl,
		GenesisUrls:    d2.GenesisUrls,
		ContactDetails: d2.ContactDetails,
		ExtraData:      d2.ExtraData,
		Snapshots:      d2.Snapshots,
		GasPrice:       d2.GasPrice,
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
