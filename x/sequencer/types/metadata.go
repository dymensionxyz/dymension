package types

import errorsmod "cosmossdk.io/errors"

// constant for maximum string length of the SequencerMetadata fields
const (
	MaxMonikerLength         = 70
	MaxIdentityLength        = 3000
	MaxWebsiteLength         = 140
	MaxDetailsLength         = 280
	MaxSecurityContactLength = 140
	MaxExtraDataLength       = 280
)

// DoNotModifyDesc constant is used in flags to indicate that the metadata field should not be updated
const DoNotModifyDesc = "[do-not-modify]"

// UpdateSequencerMetadata updates the fields of a given metadata. An error is
// returned if the resulting metadata contains an invalid length.
func (d SequencerMetadata) UpdateSequencerMetadata(d2 SequencerMetadata) (SequencerMetadata, error) {
	metadata := SequencerMetadata{
		P2PSeed:     d2.P2PSeed,
		Rpcs:        d2.Rpcs,
		EvmRpcs:     d2.EvmRpcs,
		RestApiUrls: d2.RestApiUrls,
		ExplorerUrl: d2.ExplorerUrl,
	}

	if d.Moniker != DoNotModifyDesc {
		metadata.Moniker = d2.Moniker
	}

	if d.Identity != DoNotModifyDesc {
		metadata.Identity = d2.Identity
	}

	if d.Website != DoNotModifyDesc {
		metadata.Website = d2.Website
	}

	if d.Details != DoNotModifyDesc {
		metadata.Details = d2.Details
	}

	if d.SecurityContact != DoNotModifyDesc {
		metadata.SecurityContact = d2.SecurityContact
	}

	if string(d.ExtraData) != DoNotModifyDesc {
		metadata.ExtraData = d2.ExtraData
	}

	return metadata.EnsureLength()
}

// EnsureLength ensures the length of a sequencer's metadata.
func (d SequencerMetadata) EnsureLength() (SequencerMetadata, error) {
	if len(d.Moniker) > MaxMonikerLength {
		return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid moniker length; got: %d, max: %d", len(d.Moniker), MaxMonikerLength)
	}

	if len(d.Identity) > MaxIdentityLength {
		return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid identity length; got: %d, max: %d", len(d.Identity), MaxIdentityLength)
	}

	if len(d.Website) > MaxWebsiteLength {
		return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid website length; got: %d, max: %d", len(d.Website), MaxWebsiteLength)
	}

	if len(d.Details) > MaxDetailsLength {
		return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid details length; got: %d, max: %d", len(d.Details), MaxDetailsLength)
	}

	if len(d.SecurityContact) > MaxSecurityContactLength {
		return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid security contact length; got: %d, max: %d", len(d.SecurityContact), MaxSecurityContactLength)
	}

	if len(d.ExtraData) > MaxExtraDataLength {
		return d, errorsmod.Wrapf(ErrInvalidRequest, "invalid extra data length; got: %d, max: %d", len(d.ExtraData), MaxExtraDataLength)
	}

	return d, nil
}
