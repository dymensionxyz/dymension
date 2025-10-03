package types

import (
	"slices"
)

// MergeTopRollApps merges multiple slices of PumpPressure that maintain the same rollapp order.
// When rollapp IDs match, it sums the pressure values.
// Entries with zero pressure are filtered out.
// Contract: all input slices maintain the same relative order of rollapps (all are prefixes
// of the same original sorted list, e.g., top 5 and top 10 from the same source)
// O(k*n) where k is the number of slices and n is the average length.
func MergeTopRollApps(top ...[]PumpPressure) []PumpPressure {
	if len(top) == 0 {
		return []PumpPressure{}
	}
	if len(top) == 1 {
		return filterZeroPressure(top[0])
	}

	// Merge all slices sequentially
	result := top[0]
	for i := 1; i < len(top); i++ {
		result = mergeTwoTopRollApps(result, top[i])
	}
	return result
}

// mergeTwoTopRollApps merges two slices of PumpPressure that maintain the same rollapp order.
// O(n+m) solution based on modified merge algorithm.
func mergeTwoTopRollApps(lhs, rhs []PumpPressure) []PumpPressure {
	var (
		result = make([]PumpPressure, 0, len(lhs)+len(rhs))
		i      = 0 // first iterator
		j      = 0 // second iterator
	)

	for i < len(lhs) && j < len(rhs) {
		var pump PumpPressure
		switch {
		case lhs[i].RollappId == rhs[j].RollappId:
			pump = PumpPressure{
				RollappId: lhs[i].RollappId,
				Pressure:  lhs[i].Pressure.Add(rhs[j].Pressure),
			}
			i++
			j++
		case lhs[i].RollappId < rhs[j].RollappId:
			pump = lhs[i]
			i++
		case lhs[i].RollappId > rhs[j].RollappId:
			pump = rhs[j]
			j++
		}
		// Don't store pumps with 0 pressure
		if !pump.Pressure.IsZero() {
			result = append(result, pump)
		}
	}

	if i != len(lhs) {
		result = append(result, lhs[i:]...)
	}
	if j != len(rhs) {
		result = append(result, rhs[j:]...)
	}

	return slices.Clip(result)
}

// filterZeroPressure removes entries with zero pressure from the slice
func filterZeroPressure(pumps []PumpPressure) []PumpPressure {
	result := make([]PumpPressure, 0, len(pumps))
	for _, p := range pumps {
		if !p.Pressure.IsZero() {
			result = append(result, p)
		}
	}
	return slices.Clip(result)
}
