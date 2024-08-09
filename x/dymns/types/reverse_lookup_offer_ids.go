package types

import "sort"

// Distinct returns a new list of offer-ids with duplicates removed.
// Result will be sorted.
func (m ReverseLookupOfferIds) Distinct() (distinct ReverseLookupOfferIds) {
	return ReverseLookupOfferIds{
		OfferIds: StringList(m.OfferIds).Distinct(),
	}
}

// Combine merges the offer-ids from the current list and the other list.
// Result will be sorted distinct.
func (m ReverseLookupOfferIds) Combine(other ReverseLookupOfferIds) ReverseLookupOfferIds {
	return ReverseLookupOfferIds{
		OfferIds: StringList(m.OfferIds).Combine(other.OfferIds),
	}.Distinct()
}

// Exclude removes the offer-ids from the current list that are in the toBeExcluded list.
// Result will be sorted distinct.
func (m ReverseLookupOfferIds) Exclude(toBeExcluded ReverseLookupOfferIds) (afterExcluded ReverseLookupOfferIds) {
	return ReverseLookupOfferIds{
		OfferIds: StringList(m.OfferIds).Exclude(toBeExcluded.OfferIds),
	}.Distinct()
}

// Sort sorts the offer-ids in the list.
func (m ReverseLookupOfferIds) Sort() ReverseLookupOfferIds {
	sort.Strings(m.OfferIds)
	return m
}
