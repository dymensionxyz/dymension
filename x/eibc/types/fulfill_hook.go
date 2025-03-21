package types

import (
	fmt "fmt"

	_ "github.com/cosmos/gogoproto/gogoproto"
)

func (m FulfillHook) ValidateBasic() error {
	if m.HookName == "" {
		return fmt.Errorf("hook name is empty")
	}
	return nil
}

func (m ForwardHook) ValidateBasic() error {

}
