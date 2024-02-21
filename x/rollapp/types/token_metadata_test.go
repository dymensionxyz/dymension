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
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"urax", uint32(0), []string{"microrax"}},
					{"mrax", uint32(3), []string{"millirax"}},
					{"rax", uint32(6), nil},
				},
				Base:    "urax",
				Display: "rax",
			},
			false,
		},
		{
			"duplicate exponent", // this testcase is not in x/bank/types/metadata_test.go
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"urax", uint32(0), nil},
					{"urax2", uint32(0), nil},
					{"rax", uint32(6), nil},
				},
				Base:    "urax",
				Display: "rax",
			},
			true,
		},
		{
			"base coin is display coin",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"rax", uint32(0), []string{"rax"}},
				},
				Base:    "rax",
				Display: "rax",
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
				Name:   "RollApp RAX",
				Symbol: "",
			},
			true,
		},
		{
			"invalid base denom",
			TokenMetadata{
				Name:   "RollApp RAX",
				Symbol: "rax",
				Base:   "",
			},
			true,
		},
		{
			"invalid display denom",
			TokenMetadata{
				Name:    "RollApp RAX",
				Symbol:  "rax",
				Base:    "urax",
				Display: "",
			},
			true,
		},
		{
			"duplicate denom unit",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"urax", uint32(0), []string{"microrax"}},
					{"urax", uint32(1), []string{"microrax"}},
				},
				Base:    "urax",
				Display: "rax",
			},
			true,
		},
		{
			"invalid denom unit",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"", uint32(0), []string{"microrax"}},
				},
				Base:    "urax",
				Display: "rax",
			},
			true,
		},
		{
			"invalid denom unit alias",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"urax", uint32(0), []string{""}},
				},
				Base:    "urax",
				Display: "rax",
			},
			true,
		},
		{
			"duplicate denom unit alias",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"urax", uint32(0), []string{"microrax", "microrax"}},
				},
				Base:    "urax",
				Display: "rax",
			},
			true,
		},
		{
			"no base denom unit",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"mrax", uint32(3), []string{"millirax"}},
					{"rax", uint32(6), nil},
				},
				Base:    "urax",
				Display: "rax",
			},
			true,
		},
		{
			"base denom exponent not zero",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"urax", uint32(1), []string{"microrax"}},
					{"mrax", uint32(3), []string{"millirax"}},
					{"rax", uint32(6), nil},
				},
				Base:    "urax",
				Display: "rax",
			},
			true,
		},
		{
			"invalid denom unit",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"urax", uint32(0), []string{"microrax"}},
					{"", uint32(3), []string{"millirax"}},
				},
				Base:    "urax",
				Display: "urax",
			},
			true,
		},
		{
			"no display denom unit",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"urax", uint32(0), []string{"microrax"}},
				},
				Base:    "urax",
				Display: "rax",
			},
			true,
		},
		{
			"denom units not sorted",
			TokenMetadata{
				Name:        "RollApp RAX",
				Symbol:      "rax",
				Description: "The native staking token of RollApp XYZ",
				DenomUnits: []*DenomUnit{
					{"urax", uint32(0), []string{"microrax"}},
					{"rax", uint32(6), nil},
					{"mrax", uint32(3), []string{"millirax"}},
				},
				Base:    "urax",
				Display: "rax",
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
					Description: "The native staking token of RollApp XYZ",
					DenomUnits: []*DenomUnit{
						{"urax", uint32(0), []string{"microrax"}}, // The default exponent value 0 is omitted in the json
						{"mrax", uint32(3), []string{"millirax"}},
						{"rax", uint32(6), nil},
					},
					Base:    "urax",
					Display: "rax",
				},
			},
			`[{"description":"The native staking token of RollApp XYZ","denom_units":[{"denom":"urax","aliases":["microrax"]},{"denom":"mrax","exponent":3,"aliases":["millirax"]},{"denom":"rax","exponent":6}],"base":"urax","display":"rax"}]`,
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
