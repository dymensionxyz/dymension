package types

import (
	"fmt"
	"net/url"

	errorsmod "cosmossdk.io/errors"
	"github.com/cockroachdb/errors"
)

// constant for maximum string length of the SequencerMetadata fields
const (
	MaxMonikerLength      = 70
	MaxContactFieldLength = 140
	MaxDetailsLength      = 280
	MaxExtraDataLength    = 280
	maxURLLength          = 256
)

func (d SequencerMetadata) Validate() error {
	_, err := d.EnsureLength()
	if err != nil {
		return err
	}

	if err = validateURLs(d.Rpcs); err != nil {
		return errorsmod.Wrap(err, "invalid rpcs URLs")
	}

	if err = validateURLs(d.RestApiUrls); err != nil {
		return errorsmod.Wrap(err, "invalid rest api URLs")
	}

	if d.ContactDetails == nil {
		return nil
	}

	return d.ContactDetails.Validate()
}

func (cd ContactDetails) Validate() error {
	if cd.Website != "" {
		if err := validateURL(cd.Website); err != nil {
			return errorsmod.Wrap(ErrInvalidURL, "invalid website URL")
		}
	}

	if cd.Telegram != "" {
		if err := validateURL(cd.Telegram); err != nil {
			return errorsmod.Wrap(ErrInvalidURL, "invalid telegram URL")
		}
	}

	if cd.X != "" {
		if err := validateURL(cd.X); err != nil {
			return errorsmod.Wrap(ErrInvalidURL, "invalid x URL")
		}
	}

	return nil
}

// ValidateURLs validates the URLs of a sequencer's metadata.
func validateURLs(urls []string) error {
	if len(urls) == 0 {
		return errorsmod.Wrap(ErrInvalidRequest, "urls cannot be empty")
	}

	for _, u := range urls {
		if err := validateURL(u); err != nil {
			return errorsmod.Wrap(ErrInvalidURL, err.Error())
		}
	}

	return nil
}

func validateURL(urlStr string) error {
	if urlStr == "" {
		return errorsmod.Wrap(ErrInvalidRequest, "url cannot be empty")
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
		return d, errors.Newf("invalid moniker length; got: %d, max: %d", len(d.Moniker), MaxMonikerLength)
	}

	if len(d.Details) > MaxDetailsLength {
		return d, errors.Newf("invalid details length; got: %d, max: %d", len(d.Details), MaxDetailsLength)
	}

	if len(d.ExtraData) > MaxExtraDataLength {
		return d, errors.Newf("invalid extra data length; got: %d, max: %d", len(d.ExtraData), MaxExtraDataLength)
	}

	if cd := d.ContactDetails; cd != nil {
		if len(cd.Website) > MaxContactFieldLength {
			return d, errors.Newf("invalid website length; got: %d, max: %d", len(cd.Website), MaxContactFieldLength)
		}
		if len(cd.Telegram) > MaxContactFieldLength {
			return d, errors.Newf("invalid telegram length; got: %d, max: %d", len(cd.Telegram), MaxContactFieldLength)
		}
		if len(cd.X) > MaxContactFieldLength {
			return d, errors.Newf("invalid x length; got: %d, max: %d", len(cd.X), MaxContactFieldLength)
		}
	}

	return d, nil
}
