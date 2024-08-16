package types

import "slices"

// TODO: move this to a sdk-utils package

// StringList is a list of strings.
// Used to add some operations on the list.
type StringList []string

// Distinct returns a new list with duplicates removed.
// Result will be sorted.
func (m StringList) Distinct() (distinct StringList) {
	uniqueElements := make(map[string]bool)
	// Describe usage of Go Map: used to store unique elements, later result will be sorted
	defer func() {
		distinct.Sort()
	}()

	for _, ele := range m {
		uniqueElements[ele] = true
	}

	distinctElements := make([]string, 0, len(uniqueElements))
	for name := range uniqueElements {
		distinctElements = append(distinctElements, name)
	}

	distinct = distinctElements

	return
}

// Combine merges the elements from the current list and the other list.
// Result will be sorted distinct.
func (m StringList) Combine(other StringList) StringList {
	return append(m, other...).Distinct()
}

// Exclude removes the elements from the current list that are in the toBeExcluded list.
// Result will be sorted distinct.
func (m StringList) Exclude(toBeExcluded StringList) (afterExcluded StringList) {
	var excludedElements map[string]bool
	// Describe usage of Go Map: used to store unique elements, later result will be sorted
	defer func() {
		afterExcluded.Sort()
	}()

	if len(toBeExcluded) > 0 {
		excludedElements = make(map[string]bool)

		for _, element := range toBeExcluded {
			excludedElements[element] = true
		}

		filteredElements := make([]string, 0, len(m))
		for _, element := range m {
			if !excludedElements[element] {
				filteredElements = append(filteredElements, element)
			}
		}

		afterExcluded = StringList(filteredElements).Distinct()
	} else {
		afterExcluded = m
	}

	return
}

// Sort sorts the elements in the list.
func (m StringList) Sort() StringList {
	slices.Sort(m)
	return m
}
