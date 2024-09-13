package types

import banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

func NewEventDenomMetadataCreated(metadata banktypes.Metadata) *EventDenomMetadataCreated {
	return &EventDenomMetadataCreated{Metadata: FromBankDenomMetadata(metadata)}
}

func NewEventDenomMetadataUpdated(metadata banktypes.Metadata) *EventDenomMetadataUpdated {
	return &EventDenomMetadataUpdated{Metadata: FromBankDenomMetadata(metadata)}
}

func FromBankDenomMetadata(metadata banktypes.Metadata) *DenomMetadata {
	return &DenomMetadata{
		Description: metadata.Description,
		DenomUnits:  fromBankDenomUnit(metadata.DenomUnits),
		Base:        metadata.Base,
		Display:     metadata.Display,
		Name:        metadata.Name,
		Symbol:      metadata.Symbol,
		URI:         metadata.URI,
		URIHash:     metadata.URIHash,
	}
}

func fromBankDenomUnit(units []*banktypes.DenomUnit) (denomUnits []*DenomUnit) {
	for _, u := range units {
		denomUnits = append(denomUnits, (*DenomUnit)(u))
	}
	return
}
