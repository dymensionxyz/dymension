package keeper_test

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var dymNsModuleAccAddr = authtypes.NewModuleAddress(dymnstypes.ModuleName)

func setDymNameWithFunctionsAfter(ctx sdk.Context, dymName dymnstypes.DymName, t *testing.T, dk dymnskeeper.Keeper) {
	require.NoError(t, dk.SetDymName(ctx, dymName))
	require.NoError(t, dk.AfterDymNameOwnerChanged(ctx, dymName.Name))
	require.NoError(t, dk.AfterDymNameConfigChanged(ctx, dymName.Name))
}

func requireDymNameList(dymNames []dymnstypes.DymName, wantNames []string, t *testing.T, msgAndArgs ...any) {
	var gotNames []string
	for _, dymName := range dymNames {
		gotNames = append(gotNames, dymName.Name)
	}

	sort.Strings(gotNames)
	sort.Strings(wantNames)

	if len(wantNames) == 0 {
		wantNames = nil
	}

	require.Equal(t, wantNames, gotNames, msgAndArgs...)
}

func registerRollApp(
	t *testing.T,
	ctx sdk.Context,
	rk rollappkeeper.Keeper, dk dymnskeeper.Keeper,
	rollAppID, bech32, alias string,
) {
	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId:    rollAppID,
		Owner:        testAddr(0).bech32(),
		Bech32Prefix: bech32,
	})
	if alias != "" {
		err := dk.SetAliasForRollAppId(ctx, rollAppID, alias)
		require.NoError(t, err)

		a, found := dk.GetAliasByRollAppId(ctx, rollAppID)
		require.True(t, found)
		require.Equal(t, alias, a)
	}
}

// ta stands for test-address, a simple wrapper for generating account for testing purpose.
// Usage is short, memorable, easy to type.
// The generated address is predictable, deterministic, supports output multiple formats.
type ta struct {
	bz []byte
}

// testAddr creates a general 20-bytes address from seed.
func testAddr(no uint64) ta {
	bz1 := sdk.Uint64ToBigEndian(no)
	bz2 := make([]byte, 20)
	copy(bz2, bz1)
	return ta{bz: bz2}
}

// testICAddr creates a 32-bytes address of Interchain Account from seed.
func testICAddr(no uint64) ta {
	bz1 := sdk.Uint64ToBigEndian(no)
	bz2 := make([]byte, 32)
	copy(bz2, bz1)
	return ta{bz: bz2}
}

func (a ta) bytes() []byte {
	return a.bz
}

func (a ta) bech32() string {
	return a.bech32C(params.AccountAddressPrefix)
}

func (a ta) bech32Valoper() string {
	return a.bech32C(params.AccountAddressPrefix + "valoper")
}

func (a ta) bech32C(customHrp string) string {
	return sdk.MustBech32ifyAddressBytes(customHrp, a.bz)
}

func (a ta) fallback() dymnstypes.FallbackAddress {
	return a.bz
}

func (a ta) hexStr() string {
	return dymnsutils.GetHexAddressFromBytes(a.bz)
}

func (a ta) checksumHex() string {
	if len(a.bz) != 20 {
		panic("invalid call")
	}
	return common.BytesToAddress(a.bz).Hex()
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

func (m *dymNameBuilder) cfgN(chainId, subName, resolveTo string) *dymNameBuilder {
	m.configs = append(m.configs, dymnstypes.DymNameConfig{
		Type:    dymnstypes.DymNameConfigType_DCT_NAME,
		ChainId: chainId,
		Path:    subName,
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
