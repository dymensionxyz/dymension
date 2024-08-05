package types

import (
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ReverseResolvedDymNameAddress is a struct that contains the reverse-resolved Dym-Name-Address components.
type ReverseResolvedDymNameAddress struct {
	SubName        string
	Name           string
	ChainIdOrAlias string
}

// ReverseResolvedDymNameAddresses is a list of ReverseResolvedDymNameAddress.
// Used to add some operations on the list.
type ReverseResolvedDymNameAddresses []ReverseResolvedDymNameAddress

// String returns the string representation of the ReverseResolvedDymNameAddress.
// It returns the string in the format of "subname.name@chainIdOrAlias".
func (m ReverseResolvedDymNameAddress) String() string {
	var sb strings.Builder
	if m.SubName != "" {
		sb.WriteString(m.SubName)
		sb.WriteString(".")
	}
	sb.WriteString(m.Name)
	sb.WriteString("@")
	sb.WriteString(m.ChainIdOrAlias)
	return sb.String()
}

// Sort sorts the ReverseResolvedDymNameAddress in the list.
func (m ReverseResolvedDymNameAddresses) Sort() {
	if len(m) > 0 {
		sort.Slice(m, func(i, j int) bool {
			addr1 := m[i].String()
			addr2 := m[j].String()

			if len(addr1) < len(addr2) {
				return true
			}

			if len(addr1) > len(addr2) {
				return false
			}

			return strings.Compare(addr1, addr2) < 0
		})
	}
}

// Distinct returns a new list of ReverseResolvedDymNameAddress with duplicates removed.
func (m ReverseResolvedDymNameAddresses) Distinct() (distinct ReverseResolvedDymNameAddresses) {
	if len(m) < 1 {
		return m
	}

	unique := make(map[string]ReverseResolvedDymNameAddress)
	// Describe usage of Go Map: used to store unique addresses, later result will be sorted
	defer func() {
		distinct.Sort()
	}()

	for _, addr := range m {
		unique[addr.String()] = addr
	}

	for _, addr := range unique {
		distinct = append(distinct, addr)
	}

	return
}

func (m ReverseResolvedDymNameAddresses) AppendConfigs(ctx sdk.Context, dymName DymName, configs []DymNameConfig, filter func(ReverseResolvedDymNameAddress) bool) ReverseResolvedDymNameAddresses {
	result := m[:]

	for _, config := range configs {
		record := ReverseResolvedDymNameAddress{
			SubName:        config.Path,
			Name:           dymName.Name,
			ChainIdOrAlias: config.ChainId,
		}
		if config.ChainId == "" {
			record.ChainIdOrAlias = ctx.ChainID()
		}

		if filter != nil && !filter(record) {
			continue
		}

		result = append(result, record)
	}

	return result
}
