package types

import (
	"fmt"
	"net/url"

	errorsmod "cosmossdk.io/errors"
)

func NewApp(name, rollappId, description, image, url string) App {
	return App{
		Name:        name,
		RollappId:   rollappId,
		Description: description,
		Image:       image,
		Url:         url,
	}
}

const (
	maxDescriptionLength = 512
	maxURLLength         = 256
)

func (r App) GetCreatedEvent() *EventAppCreated {
	return &EventAppCreated{
		App: &r,
	}
}

func (r App) GetUpdatedEvent() *EventAppUpdated {
	return &EventAppUpdated{
		App: &r,
	}
}

func (r App) GetDeletedEvent() *EventAppDeleted {
	return &EventAppDeleted{
		App: &r,
	}
}

func (r App) ValidateBasic() error {
	if len(r.Name) == 0 {
		return ErrInvalidName
	}

	if len(r.RollappId) == 0 {
		return ErrInvalidRollappId
	}

	if len(r.Description) > maxDescriptionLength {
		return ErrInvalidDescription
	}

	if err := validateURL(r.Image); err != nil {
		return errorsmod.Wrap(ErrInvalidImage, err.Error())
	}

	if err := validateURL(r.Url); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, err.Error())
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

	if _, err := url.Parse(urlStr); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	return nil
}
