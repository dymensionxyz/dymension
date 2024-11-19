package invar

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// return bool should be if the invariant is broken. If true, error should have meaningful debug info
type Func = func(sdk.Context) (error, bool)

type NamedFunc[K any] struct {
	Name string
	Func func(K) Func
}

func (nf NamedFunc[K]) Exec(ctx sdk.Context, module string, keeper K) (string, bool) {
	err, broken := nf.Func(keeper)(ctx)
	if err == nil {
		return "", broken
	}
	return sdk.FormatInvariant(module, nf.Name, err.Error()), broken
}

type NamedFuncsList[K any] []NamedFunc[K]

func (l NamedFuncsList[K]) RegisterInvariants(module string, ir sdk.InvariantRegistry, keeper K) {
	for _, invar := range l {
		ir.RegisterRoute(module, invar.Name, func(ctx sdk.Context) (string, bool) {
			return invar.Exec(ctx, module, keeper)
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
