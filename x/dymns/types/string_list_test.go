package types

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stringList_Distinct(t *testing.T) {
	tests := []struct {
		name         string
		provided     StringList
		wantDistinct StringList
	}{
		{
			name:         "distinct",
			provided:     []string{"a", "b", "b", "a", "c", "d"},
			wantDistinct: []string{"a", "b", "c", "d"},
		},
		{
			name:         "distinct of single",
			provided:     []string{"a"},
			wantDistinct: []string{"a"},
		},
		{
			name:         "empty",
			provided:     []string{},
			wantDistinct: []string{},
		},
		{
			name:         "nil",
			provided:     nil,
			wantDistinct: []string{},
		},
		{
			name:         "result must be sorted",
			provided:     []string{"d", "c", "a", "b", "a", "b", "e"},
			wantDistinct: []string{"a", "b", "c", "d", "e"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distinct := tt.provided.Distinct()
			want := tt.wantDistinct

			sort.Strings(distinct)
			sort.Strings(want)

			require.Equal(t, want, distinct)
		})

		t.Run(tt.name, func(t *testing.T) {
			m := ReverseLookupDymNames{
				DymNames: tt.provided,
			}
			distinct := m.Distinct().DymNames
			want := []string(tt.wantDistinct)

			sort.Strings(distinct)
			sort.Strings(want)

			require.Equal(t, want, distinct)
		})

		t.Run(tt.name, func(t *testing.T) {
			m := ReverseLookupOfferIds{
				OfferIds: tt.provided,
			}
			distinct := m.Distinct().OfferIds
			want := []string(tt.wantDistinct)

			sort.Strings(distinct)
			sort.Strings(want)

			require.Equal(t, want, distinct)
		})
	}
}

func Test_stringList_Combine(t *testing.T) {
	tests := []struct {
		name         string
		provided     StringList
		others       StringList
		wantCombined StringList
	}{
		{
			name:         "combined",
			provided:     []string{"a", "b"},
			others:       []string{"c", "d"},
			wantCombined: []string{"a", "b", "c", "d"},
		},
		{
			name:         "combined, distinct",
			provided:     []string{"a", "b"},
			others:       []string{"b", "c", "d"},
			wantCombined: []string{"a", "b", "c", "d"},
		},
		{
			name:         "combined, distinct",
			provided:     []string{"a"},
			others:       []string{"a"},
			wantCombined: []string{"a"},
		},
		{
			name:         "combine empty with other",
			provided:     nil,
			others:       []string{"a"},
			wantCombined: []string{"a"},
		},
		{
			name:         "combine empty with other",
			provided:     []string{"a"},
			others:       nil,
			wantCombined: []string{"a"},
		},
		{
			name:         "combine empty with other",
			provided:     nil,
			others:       []string{"a", "b"},
			wantCombined: []string{"a", "b"},
		},
		{
			name:         "combine with other empty",
			provided:     []string{"a", "b"},
			others:       nil,
			wantCombined: []string{"a", "b"},
		},
		{
			name:         "distinct source",
			provided:     []string{"a", "b", "a"},
			others:       []string{"c", "c", "d"},
			wantCombined: []string{"a", "b", "c", "d"},
		},
		{
			name:         "both empty",
			provided:     []string{},
			others:       []string{},
			wantCombined: []string{},
		},
		{
			name:         "both nil",
			provided:     nil,
			others:       nil,
			wantCombined: []string{},
		},
		{
			name:         "result must be sorted",
			provided:     []string{"d", "c", "a", "b"},
			others:       []string{"a", "b", "e"},
			wantCombined: []string{"a", "b", "c", "d", "e"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			combined := tt.provided.Combine(tt.others)
			want := tt.wantCombined

			sort.Strings(combined)
			sort.Strings(want)

			require.Equal(t, want, combined)
		})

		t.Run(tt.name, func(t *testing.T) {
			m := ReverseLookupDymNames{
				DymNames: tt.provided,
			}
			other := ReverseLookupDymNames{
				DymNames: tt.others,
			}
			combined := m.Combine(other).DymNames
			want := []string(tt.wantCombined)

			sort.Strings(combined)
			sort.Strings(want)

			require.Equal(t, want, combined)
		})

		t.Run(tt.name, func(t *testing.T) {
			m := ReverseLookupOfferIds{
				OfferIds: tt.provided,
			}
			other := ReverseLookupOfferIds{
				OfferIds: tt.others,
			}
			combined := m.Combine(other).OfferIds
			want := []string(tt.wantCombined)

			sort.Strings(combined)
			sort.Strings(want)

			require.Equal(t, want, combined)
		})
	}
}

func Test_stringList_Exclude(t *testing.T) {
	tests := []struct {
		name         string
		provided     StringList
		toBeExcluded StringList
		want         StringList
	}{
		{
			name:         "exclude",
			provided:     []string{"a", "b", "c", "d"},
			toBeExcluded: []string{"b", "d"},
			want:         []string{"a", "c"},
		},
		{
			name:         "exclude all",
			provided:     []string{"a", "b", "c", "d"},
			toBeExcluded: []string{"d", "c", "b", "a"},
			want:         []string{},
		},
		{
			name:         "exclude none",
			provided:     []string{"a", "b", "c", "d"},
			toBeExcluded: []string{},
			want:         []string{"a", "b", "c", "d"},
		},
		{
			name:         "exclude nil",
			provided:     []string{"a", "b", "c", "d"},
			toBeExcluded: []string{},
			want:         []string{"a", "b", "c", "d"},
		},
		{
			name:         "none exclude",
			provided:     []string{},
			toBeExcluded: []string{"a", "b", "c", "d"},
			want:         []string{},
		},
		{
			name:         "nil exclude",
			provided:     nil,
			toBeExcluded: []string{"a", "b", "c", "d"},
			want:         []string{},
		},
		{
			name:         "distinct after exclude",
			provided:     []string{"a", "a", "b"},
			toBeExcluded: []string{"b", "d"},
			want:         []string{"a"},
		},
		{
			name:         "exclude partial",
			provided:     []string{"a", "b", "c"},
			toBeExcluded: []string{"b", "c", "d"},
			want:         []string{"a"},
		},
		{
			name:         "result must be sorted",
			provided:     []string{"d", "c", "a", "b", "a", "b", "e"},
			toBeExcluded: []string{"e", "f"},
			want:         []string{"a", "b", "c", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			combined := tt.provided.Exclude(tt.toBeExcluded)
			want := tt.want

			sort.Strings(combined)
			sort.Strings(want)

			require.Equal(t, want, combined)
		})

		t.Run(tt.name, func(t *testing.T) {
			m := ReverseLookupDymNames{
				DymNames: tt.provided,
			}
			other := ReverseLookupDymNames{
				DymNames: tt.toBeExcluded,
			}
			combined := m.Exclude(other).DymNames
			want := []string(tt.want)

			sort.Strings(combined)
			sort.Strings(want)

			require.Equal(t, want, combined)
		})

		t.Run(tt.name, func(t *testing.T) {
			m := ReverseLookupOfferIds{
				OfferIds: tt.provided,
			}
			other := ReverseLookupOfferIds{
				OfferIds: tt.toBeExcluded,
			}
			combined := m.Exclude(other).OfferIds
			want := []string(tt.want)

			sort.Strings(combined)
			sort.Strings(want)

			require.Equal(t, want, combined)
		})
	}
}

func Test_stringList_Sort(t *testing.T) {
	tests := []struct {
		name     string
		provided StringList
		want     StringList
	}{
		{
			name:     "can sort",
			provided: []string{"b", "a", "c"},
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "can sort single",
			provided: []string{"a"},
			want:     []string{"a"},
		},
		{
			name:     "sort will not try distinct",
			provided: []string{"b", "a", "c", "a"},
			want:     []string{"a", "a", "b", "c"},
		},
		{
			name:     "nil",
			provided: nil,
			want:     nil,
		},
		{
			name:     "empty",
			provided: []string{},
			want:     []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.provided.Sort())
		})

		t.Run(tt.name, func(t *testing.T) {
			m := ReverseLookupDymNames{
				DymNames: tt.provided,
			}
			require.Equal(t, []string(tt.want), m.Sort().DymNames)
		})

		t.Run(tt.name, func(t *testing.T) {
			m := ReverseLookupOfferIds{
				OfferIds: tt.provided,
			}
			require.Equal(t, []string(tt.want), m.Sort().OfferIds)
		})
	}
}
