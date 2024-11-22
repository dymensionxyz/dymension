package uinv

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var ErrBroken = gerrc.ErrInternal.Wrap("invariant broken")

// Wrap an error to mark as invariant breaking. If the error is nil, it will return nil.
func Breaking(err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(ErrBroken, err)
}

// If any error that your function returns is invariant breaking, use this function to wrap
// it to reduce verbosity.
func AnyErrorIsBreaking(f Func) Func {
	return func(ctx sdk.Context) error {
		return Breaking(f(ctx))
	}
}

// Should return an ErrBroken if invariant is broken. Other errors are logged.
type Func = func(sdk.Context) error

type NamedFunc[K any] struct {
	Name string
	Func func(K) Func
}

func (nf NamedFunc[K]) Exec(ctx sdk.Context, module string, keeper K) (string, bool) {
	err := nf.Func(keeper)(ctx)
	broken := false
	var msg string
	if err != nil {
		broken = errorsmod.IsOf(err, ErrBroken)
		if !broken {
			ctx.Logger().Error("Invariant function error but not breaking.", "module", module, "name", nf.Name, "error", err)
			// Note that if it is broken the SDK wil take care of logging the error somewhere else
		}
		msg = sdk.FormatInvariant(module, nf.Name, err.Error())
	}
	return msg, broken
}

type NamedFuncsList[K any] []NamedFunc[K]

func (l NamedFuncsList[K]) RegisterInvariants(module string, ir sdk.InvariantRegistry, keeper K) {
	for _, f := range l {
		ir.RegisterRoute(module, f.Name, func(ctx sdk.Context) (string, bool) {
			return f.Exec(ctx, module, keeper)
		})
	}
}

// Should be called in a function func AllInvariants(k Keeper) sdk.Invariant within your own module namespace.
func (l NamedFuncsList[K]) All(module string, keeper K) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, invar := range l {
			s, stop := invar.Exec(ctx, module, keeper)
			if stop {
				return s, stop
			}
		}
		return "", false
	}
}
