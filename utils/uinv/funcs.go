package uinv

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var ErrBroken = errors.New("invariant broken")

func NewErr(err error) error {
	return errors.Join(ErrBroken, err)
}

// return bool should be if the invariant is broken. If true, error should have meaningful debug info
type Func = func(sdk.Context) error

type NamedFunc[K any] struct {
	Name string
	Func func(K) Func
}

func (nf NamedFunc[K]) Exec(ctx sdk.Context, module string, keeper K) (string, bool) {
	err := nf.Func(keeper)(ctx)
	broken := errorsmod.IsOf(err, ErrBroken)
	var msg string
	if err != nil {
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
