package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestReverseResolvedDymNameAddress_StringFormat(t *testing.T) {
	tests := []struct {
		name           string
		subName        string
		dymName        string
		chainIdOrAlias string
		want           string
	}{
		{
			name:           "normal case",
			subName:        "",
			dymName:        "a",
			chainIdOrAlias: "b",
			want:           "a@b",
		},
		{
			name:           "normal case with sub-name",
			subName:        "c",
			dymName:        "a",
			chainIdOrAlias: "b",
			want:           "c.a@b",
		},
		{
			name:           "normal case with multi-sub-name",
			subName:        "c.d",
			dymName:        "a",
			chainIdOrAlias: "b",
			want:           "c.d.a@b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := ReverseResolvedDymNameAddress{
				SubName:        tt.subName,
				Name:           tt.dymName,
				ChainIdOrAlias: tt.chainIdOrAlias,
			}
			require.Equal(t, tt.want, m.String())
		})
	}
}

func TestReverseResolvedDymNameAddresses_Sort(t *testing.T) {
	t.Run("allow passing empty", func(t *testing.T) {
		var m ReverseResolvedDymNameAddresses
		m.Sort()
		require.Empty(t, m)
	})

	input := ReverseResolvedDymNameAddresses{
		{
			SubName: "geek",
			Name:    "aa",
		},
		{
			SubName: "a",
			Name:    "b",
		},
		{
			SubName: "a",
			Name:    "a",
		},
		{
			SubName: "a",
			Name:    "z",
		},
		{
			SubName: "a",
			Name:    "zz",
		},
	}

	input.Sort()

	output := input

	require.Equal(t, ReverseResolvedDymNameAddresses{
		{
			SubName: "a",
			Name:    "a",
		},
		{
			SubName: "a",
			Name:    "b",
		},
		{
			SubName: "a",
			Name:    "z",
		},
		{
			SubName: "a",
			Name:    "zz",
		},
		{
			SubName: "geek",
			Name:    "aa",
		},
	}, output, "first by length, then by nature comparison")
}

func TestReverseResolvedDymNameAddresses_Distinct(t *testing.T) {
	tests := []struct {
		name string
		m    ReverseResolvedDymNameAddresses
		want ReverseResolvedDymNameAddresses
	}{
		{
			name: "distinct",
			m:    []ReverseResolvedDymNameAddress{{Name: "a"}, {Name: "b"}, {Name: "a"}},
			want: []ReverseResolvedDymNameAddress{{Name: "a"}, {Name: "b"}},
		},
		{
			name: "distinct",
			m:    []ReverseResolvedDymNameAddress{{Name: "a"}, {Name: "a"}},
			want: []ReverseResolvedDymNameAddress{{Name: "a"}},
		},
		{
			name: "already distinct",
			m:    []ReverseResolvedDymNameAddress{{Name: "a"}},
			want: []ReverseResolvedDymNameAddress{{Name: "a"}},
		},
		{
			name: "input is empty",
			m:    []ReverseResolvedDymNameAddress{},
			want: []ReverseResolvedDymNameAddress{},
		},
		{
			name: "input is nil",
			m:    nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.Distinct()
			tt.want.Sort()
			got.Sort()
			require.Equal(t, tt.want, got)
		})
	}

	t.Run("sorted after distinct", func(t *testing.T) {
		original := ReverseResolvedDymNameAddresses{
			{
				SubName: "geek",
				Name:    "aa",
			},
			{
				SubName: "a",
				Name:    "b",
			},
			{
				SubName: "a",
				Name:    "a",
			},
			{
				SubName: "a",
				Name:    "z",
			},
			{
				SubName: "a",
				Name:    "zz",
			},
		}

		duplicated := append(original, original...)
		duplicated = append(duplicated, original...)

		afterDistinct := duplicated.Distinct()

		require.Equal(t, ReverseResolvedDymNameAddresses{
			{
				SubName: "a",
				Name:    "a",
			},
			{
				SubName: "a",
				Name:    "b",
			},
			{
				SubName: "a",
				Name:    "z",
			},
			{
				SubName: "a",
				Name:    "zz",
			},
			{
				SubName: "geek",
				Name:    "aa",
			},
		}, afterDistinct, "must be sorted after distinct")
	})
}

func TestReverseResolvedDymNameAddresses_AppendConfigs(t *testing.T) {
	tests := []struct {
		name    string
		ctx     sdk.Context
		input   ReverseResolvedDymNameAddresses
		dymName DymName
		configs []DymNameConfig
		filter  func(ReverseResolvedDymNameAddress) bool
		want    ReverseResolvedDymNameAddresses
	}{
		{
			name: "can append without filter",
			ctx:  sdk.Context{}.WithChainID("dymension_1100-1"),
			input: ReverseResolvedDymNameAddresses{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "c",
				},
			},
			dymName: DymName{
				Name: "my-name",
			},
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "c1",
					Path:    "sub1",
					Value:   "v1",
				},
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "c2",
					Path:    "sub2",
					Value:   "v2",
				},
			},
			filter: func(address ReverseResolvedDymNameAddress) bool {
				return true
			},
			want: ReverseResolvedDymNameAddresses{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "c",
				},
				{
					SubName:        "sub1",
					Name:           "my-name",
					ChainIdOrAlias: "c1",
				},
				{
					SubName:        "sub2",
					Name:           "my-name",
					ChainIdOrAlias: "c2",
				},
			},
		},
		{
			name:  "accept nil input list",
			ctx:   sdk.Context{}.WithChainID("dymension_1100-1"),
			input: nil,
			dymName: DymName{
				Name: "my-name",
			},
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "c1",
					Path:    "sub1",
					Value:   "v1",
				},
			},
			filter: func(address ReverseResolvedDymNameAddress) bool {
				return true
			},
			want: ReverseResolvedDymNameAddresses{
				{
					SubName:        "sub1",
					Name:           "my-name",
					ChainIdOrAlias: "c1",
				},
			},
		},
		{
			name:  "accept nil input list and nil config list",
			ctx:   sdk.Context{}.WithChainID("dymension_1100-1"),
			input: nil,
			dymName: DymName{
				Name: "my-name",
			},
			configs: nil,
			filter: func(address ReverseResolvedDymNameAddress) bool {
				return true
			},
			want: nil,
		},
		{
			name: "when filter is nil, take all",
			ctx:  sdk.Context{},
			input: ReverseResolvedDymNameAddresses{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "c",
				},
			},
			dymName: DymName{
				Name: "my-name",
			},
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "c1",
					Path:    "sub1",
					Value:   "v1",
				},
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "c2",
					Path:    "sub2",
					Value:   "v2",
				},
			},
			filter: nil,
			want: ReverseResolvedDymNameAddresses{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "c",
				},
				{
					SubName:        "sub1",
					Name:           "my-name",
					ChainIdOrAlias: "c1",
				},
				{
					SubName:        "sub2",
					Name:           "my-name",
					ChainIdOrAlias: "c2",
				},
			},
		},
		{
			name: "use chain-id from context for configs that has empty chain-id",
			ctx:  sdk.Context{}.WithChainID("dymension_1100-1"),
			input: ReverseResolvedDymNameAddresses{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "c",
				},
			},
			dymName: DymName{
				Name: "my-name",
			},
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "sub1",
					Value:   "v1",
				},
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "c2",
					Path:    "sub2",
					Value:   "v2",
				},
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "sub3",
					Value:   "v3",
				},
			},
			filter: func(address ReverseResolvedDymNameAddress) bool {
				return true
			},
			want: ReverseResolvedDymNameAddresses{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "c",
				},
				{
					SubName:        "sub1",
					Name:           "my-name",
					ChainIdOrAlias: "dymension_1100-1",
				},
				{
					SubName:        "sub2",
					Name:           "my-name",
					ChainIdOrAlias: "c2",
				},
				{
					SubName:        "sub3",
					Name:           "my-name",
					ChainIdOrAlias: "dymension_1100-1",
				},
			},
		},
		{
			name: "filter should work",
			ctx:  sdk.Context{}.WithChainID("dymension_1100-1"),
			input: ReverseResolvedDymNameAddresses{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "dymension_1100-1",
				},
			},
			dymName: DymName{
				Name: "my-name",
			},
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "sub1",
					Value:   "v1",
				},
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "c2",
					Path:    "sub2",
					Value:   "v2",
				},
				{
					Type:    DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "sub3",
					Value:   "v3",
				},
			},
			filter: func(address ReverseResolvedDymNameAddress) bool {
				return address.ChainIdOrAlias == "dymension_1100-1"
			},
			want: ReverseResolvedDymNameAddresses{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "dymension_1100-1",
				},
				{
					SubName:        "sub1",
					Name:           "my-name",
					ChainIdOrAlias: "dymension_1100-1",
				},
				{
					SubName:        "sub3",
					Name:           "my-name",
					ChainIdOrAlias: "dymension_1100-1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.AppendConfigs(tt.ctx, tt.dymName, tt.configs, tt.filter)
			require.Equal(t, tt.want, got)
		})
	}
}
