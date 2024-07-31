package types

import (
	"sort"
	"strings"
)

type ReverseResolvedDymNameAddress struct {
	SubName        string
	Name           string
	ChainIdOrAlias string
}

type ReverseResolvedDymNameAddresses []ReverseResolvedDymNameAddress

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
