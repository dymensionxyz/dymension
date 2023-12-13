package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
)

// TestMetadataValidate tests Validate method of TokenMetadata.
// Testcases are copied from x/bank/types/metadata_test.go
func TestMetadataValidate(t *testing.T) {
	testCases := []struct {
		name     string
		metadata TokenMetadata
		expErr   bool
	}{
		{
			"non-empty coins",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
					{"matom", uint32(3), []string{"milliatom"}},
					{"atom", uint32(6), nil},
				},
				Base:    "uatom",
				Display: "atom",
			},
			false,
		},
		{
			"duplicate exponent", // this testcase is not in x/bank/types/metadata_test.go
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"uatom", uint32(0), nil},
					{"uatom2", uint32(0), nil},
					{"atom", uint32(6), nil},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"base coin is display coin",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"atom", uint32(0), []string{"ATOM"}},
				},
				Base:    "atom",
				Display: "atom",
			},
			false,
		},
		{"empty metadata", TokenMetadata{}, true},
		{
			"blank name",
			TokenMetadata{
				Name: "",
			},
			true,
		},
		{
			"blank symbol",
			TokenMetadata{
				Name:   "Cosmos Hub Atom",
				Symbol: "",
			},
			true,
		},
		{
			"invalid base denom",
			TokenMetadata{
				Name:   "Cosmos Hub Atom",
				Symbol: "ATOM",
				Base:   "",
			},
			true,
		},
		{
			"invalid display denom",
			TokenMetadata{
				Name:    "Cosmos Hub Atom",
				Symbol:  "ATOM",
				Base:    "uatom",
				Display: "",
			},
			true,
		},
		{
			"duplicate denom unit",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
					{"uatom", uint32(1), []string{"microatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"invalid denom unit",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"", uint32(0), []string{"microatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"invalid denom unit alias",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"uatom", uint32(0), []string{""}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"duplicate denom unit alias",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"uatom", uint32(0), []string{"microatom", "microatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"no base denom unit",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"matom", uint32(3), []string{"milliatom"}},
					{"atom", uint32(6), nil},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"base denom exponent not zero",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"uatom", uint32(1), []string{"microatom"}},
					{"matom", uint32(3), []string{"milliatom"}},
					{"atom", uint32(6), nil},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"invalid denom unit",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
					{"", uint32(3), []string{"milliatom"}},
				},
				Base:    "uatom",
				Display: "uatom",
			},
			true,
		},
		{
			"no display denom unit",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"denom units not sorted",
			TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
					{"atom", uint32(6), nil},
					{"matom", uint32(3), []string{"milliatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.metadata.Validate()

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMarshalJSONMetaData tests MarshalJSON for TokenMetadata.
// Testcases are copied from x/bank/types/metadata_test.go
func TestMarshalJSONMetaData(t *testing.T) {
	cdc := codec.NewLegacyAmino()

	testCases := []struct {
		name      string
		input     []TokenMetadata
		strOutput string
	}{
		{"nil metadata", nil, `null`},
		{"empty metadata", []TokenMetadata{}, `[]`},
		{
			"non-empty coins",
			[]TokenMetadata{
				{
					Description: "The native staking token of the Cosmos Hub.",
					DenomUnits: []*DenomUnit{
						{"uatom", uint32(0), []string{"microatom"}}, // The default exponent value 0 is omitted in the json
						{"matom", uint32(3), []string{"milliatom"}},
						{"atom", uint32(6), nil},
					},
					Base:    "uatom",
					Display: "atom",
				},
			},
			`[{"description":"The native staking token of the Cosmos Hub.","denom_units":[{"denom":"uatom","aliases":["microatom"]},{"denom":"matom","exponent":3,"aliases":["milliatom"]},{"denom":"atom","exponent":6}],"base":"uatom","display":"atom"}]`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			bz, err := cdc.MarshalJSON(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.strOutput, string(bz))

			var newMetadata []TokenMetadata
			require.NoError(t, cdc.UnmarshalJSON(bz, &newMetadata))

			if len(tc.input) == 0 {
				require.Nil(t, newMetadata)
			} else {
				require.Equal(t, tc.input, newMetadata)
			}
		})
	}
}
