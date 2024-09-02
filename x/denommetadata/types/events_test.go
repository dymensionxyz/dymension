package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func Test_fromBankDenomMetadata(t *testing.T) {
	type args struct {
		metadata types.Metadata
	}
	tests := []struct {
		name string
		args args
		want *DenomMetadata
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
			want: &DenomMetadata{
				Description: "New denom",
				DenomUnits: []*DenomUnit{
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
			got := fromBankDenomMetadata(tt.args.metadata)
			require.Equal(t, tt.want, got)
		})
	}
}
