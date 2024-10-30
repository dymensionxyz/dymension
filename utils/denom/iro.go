package denom

import (
	"fmt"
	"strings"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func IRODenom(rollappID string) string {
	return fmt.Sprintf("%s%s", types.IROTokenPrefix, rollappID)
}

func RollappIDFromIRODenom(denom string) (string, bool) {
	return strings.CutPrefix(denom, types.IROTokenPrefix)
}
