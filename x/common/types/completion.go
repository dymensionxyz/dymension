package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m CompletionHookCall) ValidateBasic() error {
	if m.Name == "" {
		return fmt.Errorf("hook name is empty")
	}
	return nil
}

type WasNotDelayedKey struct{}

func WithWasNotDelayed(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(WasNotDelayedKey{}, true)
}

func WasNotDelayed(ctx sdk.Context) bool {
	val, ok := ctx.Value(WasNotDelayedKey{}).(bool)
	return ok && val
}
