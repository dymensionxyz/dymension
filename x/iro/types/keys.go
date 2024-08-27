package types

import fmt "fmt"

const (
	// ModuleName defines the module name
	ModuleName = "iro"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

var (
	// KeySeparator defines the separator for keys
	KeySeparator = "/"

	// PlanKeyPrefix is the prefix to retrieve all Plans by their ID
	PlanKeyPrefix = []byte{0x1} // prefix/planId

	// PlansByRollappKeyPrefix is the prefix to retrieve all Plans by rollapp ID
	PlansByRollappKeyPrefix = []byte{0x2} // prefix/rollappId

	// LastPlanIdKey is the key to retrieve the last plan ID
	LastPlanIdKey = []byte{0x3} // lastPlanId

	// ParamsKey is the key to retrieve the module parameters
	ParamsKey = []byte{0x4} // params
)

/* --------------------- specific plan ID keys -------------------- */
func PlanKey(planId string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", PlanKeyPrefix, KeySeparator, planId))
}

/* ------------------------- multiple plans keys ------------------------ */
func PlansByRollappKey(rollappId string) []byte {
	rollappIdBytes := []byte(rollappId)
	return []byte(fmt.Sprintf("%s%s%s", PlansByRollappKeyPrefix, KeySeparator, rollappIdBytes))
}
