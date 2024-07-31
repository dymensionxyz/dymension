package keeper_test

import (
	"sort"
	"testing"
	"time"

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
func Test_msgServer_UpdateDetails(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)
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
		name                string
		dymName             *dymnstypes.DymName
		msg                 *dymnstypes.MsgUpdateDetails
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
			msg: &dymnstypes.MsgUpdateDetails{},
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
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
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
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedNoDymName(ts, owner)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
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
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedNoDymName(ts, owner)
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
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
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
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
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
			name: "fail - reject if contact is not valid",
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
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901",
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
			name: "success - can update contact",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "old-contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "new-contact@example.com",
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "new-contact@example.com",
			},
			wantMinGasConsummed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - can remove contact",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "old-contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "", // clear
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "", // cleared
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
			name: "success - can remove contact & configs",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: controller,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      "",
				ClearConfigs: true,
				Controller:   controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "",
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
			name: "success - can update contact & remove configs",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "old-contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: controller,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      "new-contact@example.com",
				ClearConfigs: true,
				Controller:   controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "new-contact@example.com",
			},
			wantMinGasConsummed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
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
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
				Controller: controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 99,
				Contact:    "contact@example.com",
			},
			wantMinGasConsummed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
		},
		{
			name: "success - configs should not be changed when update contact and not clear config",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "old-contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: controller,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      "new-contact@example.com",
				ClearConfigs: false,
				Controller:   controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "new-contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: controller,
					},
				},
			},
			wantMinGasConsummed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
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
			msg: &dymnstypes.MsgUpdateDetails{
				ClearConfigs: true,
				Controller:   controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				requireConfiguredAddressMappedNoDymName(ts, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9")
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
				require0xMappedNoDymName(ts, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9")
			},
		},
		{
			name: "success - when contact is [do-not-modify], do not update contact",
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
				},
				Contact: "contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedDymNames(ts, controller, recordName)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, controller, recordName)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      dymnstypes.DoNotModifyDesc,
				ClearConfigs: true,
				Controller:   controller,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com", // keep
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				requireConfiguredAddressMappedNoDymName(ts, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9")
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
				require0xMappedNoDymName(ts, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9")
			},
		},
		{
			name: "fail - reject message that neither update contact nor clear config",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      dymnstypes.DoNotModifyDesc,
				ClearConfigs: false,
				Controller:   controller,
			},
			wantErr:         true,
			wantErrContains: "message neither clears configs nor updates contact information",
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
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
			name: "fail - reject message that not update contact and clear config but no config to clear",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      dymnstypes.DoNotModifyDesc,
				ClearConfigs: true,
				Controller:   controller,
			},
			wantErr:         true,
			wantErrContains: "no existing config to clear",
			wantDymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: controller,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
			},
			wantMinGasConsummed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, owner, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controller)
				require0xMappedDymNames(ts, owner, recordName)
				require0xMappedNoDymName(ts, controller)
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
			resp, err := dymnskeeper.NewMsgServerImpl(dk).UpdateDetails(ctx, tt.msg)
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
