package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type ctxKeySkipRollappMiddlewares struct{}

// SkipRollappMiddlewareContext returns a context which can be passed
// to dymension ibc middlewares related to rollapp logic to skip the middleware.
func SkipRollappMiddlewareContext(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(ctxKeySkipRollappMiddlewares{}, true)
}

// SkipRollappMiddleware returns if the context contains the SkipRollappMiddleware directive
func SkipRollappMiddleware(ctx sdk.Context) bool {
	val, ok := ctx.Value(ctxKeySkipRollappMiddlewares{}).(bool)
	return ok && val
}
