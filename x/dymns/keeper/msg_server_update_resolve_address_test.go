package keeper_test

import (
	"sort"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func Test_msgServer_UpdateResolveAddress(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockHeader(tmproto.Header{
			Time: now,
		})
		ctx = ctx.WithChainID(chainId)

		return dk, ctx
	}

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	const controller = "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4"
	const recordName = "bonded-pool"

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
		dymNames, err := ts.dk.GetDymNamesContainsConfiguredAddress(ts.ctx, bech32Addr, 0)
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

		dymNames, err := ts.dk.GetDymNamesContainsHexAddress(ts.ctx, bz, 0)
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
		name                string
		dymName             *dymnstypes.DymName
		msg                 *dymnstypes.MsgUpdateResolveAddress
		preTestFunc         func(ts testSuite)
		wantErr             bool
		wantErrContains     string
		wantDymName         *dymnstypes.DymName
		wantMinGasConsummed sdk.Gas
		postTestFunc        func(ts testSuite)
	}{
		{
			name: "fail - reject if message not pass validate basic",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			wantErr:         true,
			wantErrContains: dymnstypes.ErrValidationFailed.Error(),
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name:    "fail - Dym-Name does not exists",
			dymName: nil,
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedNoDymName(ts, owner)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr:         true,
			wantErrContains: dymnstypes.ErrDymNameNotFound.Error(),
			wantDymName:     nil,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedNoDymName(ts, owner)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "fail - reject if Dym-Name expired",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() - 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr:         true,
			wantErrContains: "Dym-Name is already expired",
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() - 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "fail - reject if sender is neither owner nor controller",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  owner,
				Controller: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			},
			wantErr:         true,
			wantErrContains: sdkerrors.ErrUnauthorized.Error(),
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "fail - reject if sender is owner but not controller",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  owner,
				Controller: owner,
			},
			wantErr:         true,
			wantErrContains: "please use controller account",
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "fail - reject if config is not valid",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "0x1",
				Controller: controller,
			},
			wantErr:         true,
			wantErrContains: dymnstypes.ErrValidationFailed.Error(),
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "fail - reject if config is not valid. Only accept lowercase",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				SubName:    "AAAAA", // upcase is not accepted
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr:         true,
			wantErrContains: dymnstypes.ErrValidationFailed.Error(),
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - can update",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   owner,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - add new record if not exists",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: owner,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - override record if exists",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controller,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: owner,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - remove record if new resolve to empty, single-config",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "a",
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs:    nil,
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - remove record if new resolve to empty, single-config, not match any",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "non-exists",
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
				},
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - remove record if new resolve to empty, multi-config, first",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controller,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "a",
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controller,
					},
				},
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
		},
		{
			name: "success - remove record if new resolve to empty, multi-configs, last",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controller,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
				},
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - remove record if new resolve to empty, multi-config, not any of existing",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controller,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "non-exists",
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: owner,
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controller,
					},
				},
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
		},
		{
			name: "success - expiry not changed",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 99,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 99,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: owner,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - chain-id automatically removed from record if is host chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    chainId,
				SubName:    "a",
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // empty
						Path:    "a",
						Value:   owner,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - chain-id automatically removed from record if is host chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // originally empty
						Path:    "a",
						Value:   controller,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    chainId,
				SubName:    "a",
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // empty
						Path:    "a",
						Value:   owner,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - chain-id recorded if is NOT host chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "blumbus_100-1",
				SubName:    "a",
				ResolveTo:  owner,
				Controller: controller,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   owner,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - do not override record with different chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   owner,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "blumbus_100-1",
				SubName:    "a",
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   owner,
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   owner,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - do not override record with different chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   controller,
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   controller,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "blumbus_100-1",
				SubName:    "a",
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   controller,
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   owner,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 on host chain if target chain is host chain, case empty chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "",
				SubName:    "a",
				ResolveTo:  "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx", // owner but with nim prefix
				Controller: controller,
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 on host chain if target chain is host chain, case use chain-id in request",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    chainId,
				SubName:    "a",
				ResolveTo:  "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx", // owner but with nim prefix
				Controller: controller,
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				requireConfiguredAddressMappedNoDymName(ts, "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx")
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 on host chain if target chain is host chain, case dym prefix but valoper, not acc addr",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "",
				SubName:    "a",
				ResolveTo:  "dymvaloper1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3ydetzr", // owner but with valoper prefix
				Controller: controller,
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				requireConfiguredAddressMappedNoDymName(ts, "dymvaloper1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3ydetzr")
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - reverse mapping record should be updated accordingly",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   controller,
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "a",
						Value:   "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				requireConfiguredAddressMappedDymNames(ts, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9")
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "",
				SubName:    "",
				ResolveTo:  owner,
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   owner,
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "a",
						Value:   "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				requireConfiguredAddressMappedDymNames(ts, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", recordName)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
				require0xMappedNoDymName(ts, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9")
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
				if tt.wantMinGasConsummed > 0 {
					require.GreaterOrEqual(t,
						ctx.GasMeter().GasConsumed(), tt.wantMinGasConsummed,
						"should consume at least %d gas", tt.wantMinGasConsummed,
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

					owned, err := dk.GetDymNamesOwnedBy(ctx, laterDymName.Owner, 0)
					require.NoError(t, err)
					require.Len(t, owned, 1)
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
