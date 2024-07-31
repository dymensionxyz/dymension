package types

import (
	"testing"

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
			SubName: "aaaa",
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
			SubName: "aaaa",
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
				SubName: "aaaa",
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
				SubName: "aaaa",
				Name:    "aa",
			},
		}, afterDistinct, "must be sorted after distinct")
	})
}
