package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/stretchr/testify/require"
)

func setDymNameWithFunctionsAfter(ctx sdk.Context, dymName dymnstypes.DymName, t *testing.T, dk dymnskeeper.Keeper) {
	require.NoError(t, dk.SetDymName(ctx, dymName))
	require.NoError(t, dk.AfterDymNameOwnerChanged(ctx, dymName.Name))
	require.NoError(t, dk.AfterDymNameConfigChanged(ctx, dymName.Name))
}

func requireErrorContains(t *testing.T, err error, contains string) {
	require.Error(t, err)
	require.NotEmpty(t, contains, "mis-configured test")
	require.Contains(t, err.Error(), contains)
}

func requireErrorFContains(t *testing.T, f func() error, contains string) {
	requireErrorContains(t, f(), contains)
}

type dymNameBuilder struct {
	name       string
	owner      string
	controller string
	expireAt   int64
	configs    []dymnstypes.DymNameConfig
}

func newDN(name, owner string) *dymNameBuilder {
	return &dymNameBuilder{
		name:       name,
		owner:      owner,
		controller: owner,
		expireAt:   time.Now().Unix() + 10,
		configs:    nil,
	}
}

func (m *dymNameBuilder) exp(now time.Time, offset int64) *dymNameBuilder {
	m.expireAt = now.Unix() + offset
	return m
}

func (m *dymNameBuilder) cfgN(chainId, path, resolveTo string) *dymNameBuilder {
	m.configs = append(m.configs, dymnstypes.DymNameConfig{
		Type:    dymnstypes.DymNameConfigType_NAME,
		ChainId: chainId,
		Path:    path,
		Value:   resolveTo,
	})
	return m
}

func (m *dymNameBuilder) build() dymnstypes.DymName {
	return dymnstypes.DymName{
		Name:       m.name,
		Owner:      m.owner,
		Controller: m.controller,
		ExpireAt:   m.expireAt,
		Configs:    m.configs,
	}
}

func (m *dymNameBuilder) buildSlice() []dymnstypes.DymName {
	return []dymnstypes.DymName{m.build()}
}
