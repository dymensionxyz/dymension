package keeper_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_UpdateResolveAddress(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now).WithChainID(chainId)

		return dk, ctx
	}

	ownerAcc := testAddr(1)
	controllerAcc := testAddr(2)
	anotherAcc := testAddr(3)

	const recordName = "my-name"

	params.SetAddressPrefixes()

	type testSuite struct {
		t   *testing.T
		dk  dymnskeeper.Keeper
		ctx sdk.Context
	}

	nts := func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) testSuite {
		return testSuite{
			t:   t,
			dk:  dk,
			ctx: ctx,
		}
	}

	requireConfiguredAddressMappedDymNames := func(ts testSuite, bech32Addr string, names ...string) {
		dymNames, err := ts.dk.GetDymNamesContainsConfiguredAddress(ts.ctx, bech32Addr)
		require.NoError(ts.t, err)
		require.Len(ts.t, dymNames, len(names))
		sort.Strings(names)
		sort.Slice(dymNames, func(i, j int) bool {
			return dymNames[i].Name < dymNames[j].Name
		})
		for i, name := range names {
			require.Equal(ts.t, name, dymNames[i].Name)
		}
	}

	requireConfiguredAddressMappedNoDymName := func(ts testSuite, bech32Addr string) {
		requireConfiguredAddressMappedDymNames(ts, bech32Addr)
	}

	require0xMappedDymNames := func(ts testSuite, bech32Addr string, names ...string) {
		_, bz, err := bech32.DecodeAndConvert(bech32Addr)
		require.NoError(ts.t, err)

		dymNames, err := ts.dk.GetDymNamesContainsHexAddress(ts.ctx, bz)
		require.NoError(ts.t, err)
		require.Len(ts.t, dymNames, len(names))
		sort.Strings(names)
		sort.Slice(dymNames, func(i, j int) bool {
			return dymNames[i].Name < dymNames[j].Name
		})
		for i, name := range names {
			require.Equal(ts.t, name, dymNames[i].Name)
		}
	}

	require0xMappedNoDymName := func(ts testSuite, bech32Addr string) {
		require0xMappedDymNames(ts, bech32Addr)
	}

	tests := []struct {
		name               string
		dymName            *dymnstypes.DymName
		msg                *dymnstypes.MsgUpdateResolveAddress
		preTestFunc        func(ts testSuite)
		wantErr            bool
		wantErrContains    string
		wantDymName        *dymnstypes.DymName
		wantMinGasConsumed sdk.Gas
		postTestFunc       func(ts testSuite)
	}{
		{
			name: "fail - reject if message not pass validate basic",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			wantErr:         true,
			wantErrContains: gerrc.ErrInvalidArgument.Error(),
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name:    "fail - Dym-Name does not exists",
			dymName: nil,
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: fmt.Sprintf("Dym-Name: %s: not found", recordName),
			wantDymName:     nil,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "fail - reject if Dym-Name expired",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() - 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "Dym-Name is already expired",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() - 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "fail - reject if sender is neither owner nor controller",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: anotherAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "permission denied",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "fail - reject if sender is owner but not controller",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: ownerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "please use controller account",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "fail - reject if config is not valid",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "0x1",
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "config is invalid:",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "fail - reject if config is not valid. Only accept lowercase",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				SubName:    "SUB", // upper-case is not accepted
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "config is invalid:",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - can update",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - add new record if not exists",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - override record if exists",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedDymNames(ts, controllerAcc.bech32(), recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - remove record if new resolve to empty, single-config",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "a",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs:    nil,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - remove record if new resolve to empty, single-config, not match any",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "non-exists",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - remove record if new resolve to empty, multi-config, first",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedDymNames(ts, controllerAcc.bech32(), recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "a",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedDymNames(ts, controllerAcc.bech32(), recordName)
			},
		},
		{
			name: "pass - remove record if new resolve to empty, multi-configs, last",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedDymNames(ts, controllerAcc.bech32(), recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - remove record if new resolve to empty, multi-config, not any of existing",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedDymNames(ts, controllerAcc.bech32(), recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "non-exists",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedDymNames(ts, controllerAcc.bech32(), recordName)
			},
		},
		{
			name: "pass - expiry not changed",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 99,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 99,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - chain-id automatically removed from record if is host chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    chainId,
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // empty
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - chain-id automatically removed from record if is host chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // originally empty
						Path:    "a",
						Value:   controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    chainId,
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // empty
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - chain-id recorded if is NOT host chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "blumbus_100-1",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - do not override record with different chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "blumbus_100-1",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - do not override record with different chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   controllerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "blumbus_100-1",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   controllerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 on host chain if target chain is host chain, case empty chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32C("nim"), // owner but with nim prefix
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 on host chain if target chain is host chain, case use chain-id in request",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    chainId,
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32C("nim"), // owner but with nim prefix
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32C("nim"))
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 on host chain if target chain is host chain, case dym prefix but valoper, not acc addr",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32Valoper(), // owner but with valoper prefix
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32Valoper())
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
			},
		},
		{
			name: "pass - reverse mapping record should be updated accordingly",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   controllerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "a",
						Value:   anotherAcc.bech32C("nim"),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, anotherAcc.bech32C("nim"), recordName)
				require0xMappedNoDymName(ts, ownerAcc.bech32())
				require0xMappedDymNames(ts, controllerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, anotherAcc.bech32C("nim"))
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "",
				SubName:    "",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "a",
						Value:   anotherAcc.bech32C("nim"),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireConfiguredAddressMappedDymNames(ts, anotherAcc.bech32C("nim"), recordName)
				require0xMappedDymNames(ts, ownerAcc.bech32(), recordName)
				require0xMappedNoDymName(ts, controllerAcc.bech32())
				require0xMappedNoDymName(ts, anotherAcc.bech32C("nim"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.preTestFunc)
			require.NotNil(t, tt.postTestFunc)

			dk, ctx := setupTest()

			if tt.dymName != nil {
				if tt.dymName.Name == "" {
					tt.dymName.Name = recordName
				}
				err := dk.SetDymName(ctx, *tt.dymName)
				require.NoError(t, err)
				require.NoError(t, dk.AfterDymNameOwnerChanged(ctx, tt.dymName.Name))
				require.NoError(t, dk.AfterDymNameConfigChanged(ctx, tt.dymName.Name))
			}
			if tt.wantDymName != nil && tt.wantDymName.Name == "" {
				tt.wantDymName.Name = recordName
			}

			tt.preTestFunc(nts(t, dk, ctx))

			tt.msg.Name = recordName
			resp, err := dymnskeeper.NewMsgServerImpl(dk).UpdateResolveAddress(ctx, tt.msg)
			laterDymName := dk.GetDymName(ctx, tt.msg.Name)

			defer func() {
				if tt.wantMinGasConsumed > 0 {
					require.GreaterOrEqual(t,
						ctx.GasMeter().GasConsumed(), tt.wantMinGasConsumed,
						"should consume at least %d gas", tt.wantMinGasConsumed,
					)
				}

				if !t.Failed() {
					tt.postTestFunc(nts(t, dk, ctx))
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				require.Nil(t, resp)

				if tt.wantDymName != nil {
					require.Equal(t, *tt.wantDymName, *laterDymName)

					owned, err := dk.GetDymNamesOwnedBy(ctx, laterDymName.Owner)
					require.NoError(t, err)
					if laterDymName.ExpireAt >= now.Unix() {
						require.Len(t, owned, 1)
					} else {
						require.Empty(t, owned)
					}
				} else {
					require.Nil(t, laterDymName)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, laterDymName)
			require.Equal(t, *tt.wantDymName, *laterDymName)
		})
	}
}
