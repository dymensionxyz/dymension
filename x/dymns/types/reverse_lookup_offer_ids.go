package types

import "sort"

// Distinct returns a new list of offer-ids with duplicates removed.
// Result will be sorted.
func (m ReverseLookupBuyOrderIds) Distinct() (distinct ReverseLookupBuyOrderIds) {
	return ReverseLookupBuyOrderIds{
		OrderIds: StringList(m.OrderIds).Distinct(),
	}
}

// Combine merges the offer-ids from the current list and the other list.
// Result will be sorted distinct.
func (m ReverseLookupBuyOrderIds) Combine(other ReverseLookupBuyOrderIds) ReverseLookupBuyOrderIds {
	return ReverseLookupBuyOrderIds{
		OrderIds: StringList(m.OrderIds).Combine(other.OrderIds),
	}.Distinct()
}

// Exclude removes the offer-ids from the current list that are in the toBeExcluded list.
// Result will be sorted distinct.
func (m ReverseLookupBuyOrderIds) Exclude(toBeExcluded ReverseLookupBuyOrderIds) (afterExcluded ReverseLookupBuyOrderIds) {
	return ReverseLookupBuyOrderIds{
		OrderIds: StringList(m.OrderIds).Exclude(toBeExcluded.OrderIds),
	}.Distinct()
}

// Sort sorts the offer-ids in the list.
func (m ReverseLookupBuyOrderIds) Sort() ReverseLookupBuyOrderIds {
	sort.Strings(m.OrderIds)
	return m
}
