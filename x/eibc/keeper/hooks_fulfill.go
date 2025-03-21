package keeper

import (
	"fmt"

	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/gogo/protobuf/proto"
)

const (
	HookNameForward = "forward"
)

// assumed already passed validate basic
func validateFulfillHook(info types.FulfillHook) error {
	switch info.HookName {
	case HookNameForward:
		return validForward(info.HookData)
	default:
		return fmt.Errorf("invalid hook name: %s", info.HookName)
	}
}

func validForward(data []byte) error {
	var d types.ForwardHook
	err := proto.Unmarshal(data, &d)
	if err != nil {
		return fmt.Errorf("unmarshal forward hook metadata: %w", err)
	}
	if err := d.ValidateBasic(); err != nil {
		return fmt.Errorf("validate forward hook metadata: %w", err)
	}
	return nil
}
