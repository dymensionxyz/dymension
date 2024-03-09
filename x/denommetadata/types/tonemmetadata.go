package types

import (
	"fmt"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func NewTokenMetadata(description string, denomstring string, exponent uint32, denomalias string, base string, display string, name string, symbol string, uri string, urihash string) *TokenMetadata {

	du := &DenomUnit{
		Denom:    denomstring,
		Exponent: exponent,
		Aliases:  []string{denomalias},
	}
	dus := []*DenomUnit{du}
	tokenMetadata := &TokenMetadata{
		Description: description,
		DenomUnits:  dus,
		Base:        base,
		Display:     display,
		Name:        name,
		Symbol:      symbol,
		URI:         uri,
		URIHash:     urihash,
	}

	return tokenMetadata
}

// Validate performs a basic validation of the coin metadata fields.
// Inherits from x/bank metadata and following same spec of x/bank/types/metadata.go
func (m *TokenMetadata) Validate() error {
	if m == nil {
		return fmt.Errorf("token metadata cannot be nil")
	}

	bankMetadata := m.ConvertToBankMetadata()
	return bankMetadata.Validate()
}

// ConvertToBankMetadata converts TokenMetadata to Metadata of x/bank/types
func (m *TokenMetadata) ConvertToBankMetadata() banktypes.Metadata {
	var denomUnits []*banktypes.DenomUnit

	for _, denomUnit := range m.DenomUnits {
		denomUnits = append(denomUnits, &banktypes.DenomUnit{
			Denom:    denomUnit.Denom,
			Exponent: denomUnit.Exponent,
			Aliases:  denomUnit.Aliases,
		})
	}

	return banktypes.Metadata{
		Description: m.Description,
		DenomUnits:  denomUnits,
		Base:        m.Base,
		Display:     m.Display,
		Name:        m.Name,
		Symbol:      m.Symbol,
		URI:         m.URI,
		URIHash:     m.URIHash,
	}
}
