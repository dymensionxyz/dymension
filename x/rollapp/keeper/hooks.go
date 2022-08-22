package keeper

import (
	"errors"
	fmt "fmt"
	"runtime"
	"runtime/debug"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements RollappHooks interface in a recoverable manner
// RollappHooks event hooks for rollapp object (noalias)
type RollappRecoverableHooks interface {
	BeforeUpdateStateRecoverable(ctx sdk.Context, seqAddr string, rollappId string) (err error) // Must be called when a rollapp's state changes
}

var _ RollappRecoverableHooks = Keeper{}

// BeforeUpdateState - call hook if registered
// incase of panic, drop the state machine change and return the error.
func (k Keeper) BeforeUpdateStateRecoverable(ctx sdk.Context, seqAddr string, rollappId string) (err error) {
	if k.GetHooks() != nil {
		return panicCatchingHook(
			ctx,
			func(ctx sdk.Context, arg ...any) {
				k.hooks.BeforeUpdateState(ctx, arg[0].(string), arg[1].(string))
			},
			seqAddr, rollappId)
	}
	return nil
}

//--------------------------------------------------------------------------------------------------------
// panicCatchingHook lets us run the hook function hookFn, but if theres an error or panic
// drop the state machine change, log the error and return it.
// If there is no error, proceeds as normal (but with some slowdown due to SDK store weirdness)
func panicCatchingHook(
	ctx sdk.Context,
	hookFn func(ctx sdk.Context, arg ...any),
	args ...any,
) (err error) {
	defer func() {
		if recovErr := recover(); recovErr != nil {
			err = PrintPanicRecoveryError(ctx, recovErr)
		}
	}()
	cacheCtx, write := ctx.CacheContext()
	hookFn(cacheCtx, args...)
	if err != nil {
		ctx.Logger().Error(err.Error())
	} else {
		// no error, write the output of f
		write()
	}
	return err
}

// PrintPanicRecoveryError error logs the recoveryError, along with the stacktrace, if it can be parsed.
// If not emits them to stdout.
func PrintPanicRecoveryError(ctx sdk.Context, recoveryError interface{}) error {
	err := errors.New("panic occurred during execution")
	errStackTrace := string(debug.Stack())
	switch e := recoveryError.(type) {
	case string:
		ctx.Logger().Error("Recovering from (string) panic: " + e)
	case runtime.Error:
		err = e
		ctx.Logger().Error("recovered (runtime.Error) panic: " + e.Error())
	case error:
		err = e
		ctx.Logger().Error("recovered (error) panic: " + e.Error())
	default:
		ctx.Logger().Error("recovered (default) panic. Could not capture logs in ctx, see stdout")
		fmt.Println("Recovering from panic ", recoveryError)
		debug.PrintStack()
		return err
	}
	ctx.Logger().Error("stack trace: " + errStackTrace)

	return err
}
