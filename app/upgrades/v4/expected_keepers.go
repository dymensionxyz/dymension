package v4

import paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

type ParamsKeeper interface {
	Subspace(s string) paramstypes.Subspace
	GetSubspaces() []paramstypes.Subspace
}
