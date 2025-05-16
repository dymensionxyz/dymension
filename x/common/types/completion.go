package types

import fmt "fmt"

func (m CompletionHookCall) ValidateBasic() error {
	if m.Name == "" {
		return fmt.Errorf("hook name is empty")
	}
	return nil
}
