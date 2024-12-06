package types

import (
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type UnexpectedFinishedGaugeError struct {
	GaugeId uint64
}

func (e UnexpectedFinishedGaugeError) Error() string {
	return gerrc.ErrInternal.Wrapf("gauge already finished: id: %d", e.GaugeId).Error()
}
