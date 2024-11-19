package invar

import sdk "github.com/cosmos/cosmos-sdk/types"

type Func = func(sdk.Context) (error, bool)

type NamedFunc[K any] struct {
	Name string
	Func func(K) Func
}

type NamedFuncsList[K any] []NamedFunc[K]

func (l NamedFuncsList[K]) RegisterInvariants(module string, ir sdk.InvariantRegistry, keeper K) {
	for _, invar := range l {
		ir.RegisterRoute(module, invar.Name, func(ctx sdk.Context) (string, bool) {
			err, broken := invar.Func(keeper)(ctx)
			return err.Error(), broken
		})
	}
}

func (l NamedFuncsList[K]) All(keeper K) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, invar := range l {
			err, broken := invar.Func(keeper)(ctx)
			if broken {
				return err.Error(), broken
			}
		}
		return "", false
	}
}
