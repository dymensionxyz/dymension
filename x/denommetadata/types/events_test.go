package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	denomtypes "github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

func Test_fromBankDenomMetadata(t *testing.T) {
	type args struct {
		metadata types.Metadata
	}
	tests := []struct {
		name string
		args args
		want *denomtypes.DenomMetadata
	}{
		{
			name: "success",
			args: args{
				metadata: types.Metadata{
					Description: "New denom",
					DenomUnits: []*types.DenomUnit{
						{
							Denom:    "denom",
							Exponent: 18,
							Aliases:  []string{"alias"},
						},
					},
					Base:    "aden",
					Display: "den",
					Name:    "Denom",
					Symbol:  "DEN",
					URI:     "https://denom.com",
					URIHash: "hash",
				},
			},
			want: &denomtypes.DenomMetadata{
				Description: "New denom",
				DenomUnits: []*denomtypes.DenomUnit{
					{
						Denom:    "denom",
						Exponent: 18,
						Aliases:  []string{"alias"},
					},
				},
				Base:    "aden",
				Display: "den",
				Name:    "Denom",
				Symbol:  "DEN",
				URI:     "https://denom.com",
				URIHash: "hash",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := denomtypes.FromBankDenomMetadata(tt.args.metadata)
			require.Equal(t, tt.want, got)
		})
	}
}
