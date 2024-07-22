package types

import fmt "fmt"

type UnexpectedFinishedGaugeError struct {
	GaugeId uint64
}

func (e UnexpectedFinishedGaugeError) Error() string {
	return fmt.Sprintf("gauge with ID (%d) is already finished", e.GaugeId)
}
