package types

import fmt "fmt"

func (m CompletionHook) ValidateBasic() error {
	if m.HookName == "" {
		return fmt.Errorf("hook name is empty")
	}
	return nil
}
