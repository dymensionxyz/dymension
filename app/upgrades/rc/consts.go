package rc

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	UpgradeName = "rc-TBD"
)

var (
	DYMTokenMetata = banktypes.Metadata{
		Description: "Denom metadata for DYM (udym)",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "udym",
				Exponent: 0,
				Aliases:  []string{},
			},
			{
				Denom:    "DYM",
				Exponent: 18,
				Aliases:  []string{},
			},
		},
		Base:    "udym",
		Display: "DYM",
		Name:    "DYM",
		Symbol:  "DYM",
		URI:     "",
		URIHash: "",
	}
)
