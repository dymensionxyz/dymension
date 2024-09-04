package types

import (
	errorsmod "cosmossdk.io/errors"
)

func NewApp(name, rollappId, description, image, url string, order int32) App {
	return App{
		Name:        name,
		RollappId:   rollappId,
		Description: description,
		ImageUrl:    image,
		Url:         url,
		Order:       order,
	}
}

func (r App) GetAddedEvent() *EventAppAdded {
	return &EventAppAdded{
		App: &r,
	}
}

func (r App) GetUpdatedEvent() *EventAppUpdated {
	return &EventAppUpdated{
		App: &r,
	}
}

func (r App) GetRemovedEvent() *EventAppRemoved {
	return &EventAppRemoved{
		App: &r,
	}
}

func (r App) ValidateBasic() error {
	if len(r.Name) == 0 {
		return ErrInvalidAppName
	}

	if len(r.Name) > maxAppNameLength {
		return ErrInvalidAppName
	}

	if len(r.RollappId) == 0 {
		return ErrInvalidRollappID
	}

	if len(r.Description) > maxDescriptionLength {
		return ErrInvalidDescription
	}

	if err := validateURL(r.ImageUrl); err != nil {
		return errorsmod.Wrap(ErrInvalidAppImage, err.Error())
	}

	if err := validateURL(r.Url); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, err.Error())
	}

	return nil
}
