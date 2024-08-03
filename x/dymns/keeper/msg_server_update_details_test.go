package keeper_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_UpdateDetails(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)
		ctx = ctx.WithChainID(chainId)

		return dk, ctx
	}

	ownerAcc := testAddr(1)
	ownerA := ownerAcc.bech32()

	controllerAcc := testAddr(2)
	controllerA := controllerAcc.bech32()

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

	requireFallbackAddressMappedDymNames := func(ts testSuite, fallbackAddr dymnstypes.FallbackAddress, names ...string) {
		dymNames, err := ts.dk.GetDymNamesContainsFallbackAddress(ts.ctx, fallbackAddr)
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

	requireFallbackAddressMappedNoDymName := func(ts testSuite, fallbackAddr dymnstypes.FallbackAddress) {
		requireFallbackAddressMappedDymNames(ts, fallbackAddr)
	}

	tests := []struct {
		name               string
		dymName            *dymnstypes.DymName
		msg                *dymnstypes.MsgUpdateDetails
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
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
			},
			msg: &dymnstypes.MsgUpdateDetails{},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			wantErr:         true,
			wantErrContains: gerrc.ErrInvalidArgument.Error(),
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name:    "fail - Dym-Name does not exists",
			dymName: nil,
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
				Controller: controllerA,
			},
			wantErr:         true,
			wantErrContains: fmt.Sprintf("Dym-Name: %s: not found", recordName),
			wantDymName:     nil,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if Dym-Name expired",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() - 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
				Controller: controllerA,
			},
			wantErr:         true,
			wantErrContains: "Dym-Name is already expired",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() - 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if sender is neither owner nor controller",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
				Controller: anotherAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "permission denied",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if sender is owner but not controller",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
				Controller: ownerA,
			},
			wantErr:         true,
			wantErrContains: "please use controller account",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if contact is not valid",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901",
				Controller: controllerA,
			},
			wantErr:         true,
			wantErrContains: "contact is too long; max length",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - can update contact",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "old-contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "new-contact@example.com",
				Controller: controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "new-contact@example.com",
			},
			wantMinGasConsumed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - can remove contact",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "old-contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "", // clear
				Controller: controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "", // cleared
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - can remove contact & configs",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: controllerA,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedDymNames(ts, controllerA, recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      "",
				ClearConfigs: true,
				Controller:   controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "",
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - can update contact & remove configs",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "old-contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: controllerA,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedDymNames(ts, controllerA, recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      "new-contact@example.com",
				ClearConfigs: true,
				Controller:   controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "new-contact@example.com",
			},
			wantMinGasConsumed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - expiry not changed",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 99,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
				Controller: controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 99,
				Contact:    "contact@example.com",
			},
			wantMinGasConsumed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - configs should not be changed when update contact and not clear config",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "old-contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: controllerA,
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedDymNames(ts, controllerA, recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      "new-contact@example.com",
				ClearConfigs: false,
				Controller:   controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "new-contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: controllerA,
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedDymNames(ts, controllerA, recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
		},
		{
			name: "pass - reverse mapping record should be updated accordingly",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   controllerA,
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
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedDymNames(ts, controllerA, recordName)
				requireConfiguredAddressMappedDymNames(ts, anotherAcc.bech32C("nim"), recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				ClearConfigs: true,
				Controller:   controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireConfiguredAddressMappedNoDymName(ts, anotherAcc.bech32C("nim"))
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())
			},
		},
		{
			name: "pass - when contact is [do-not-modify], do not update contact",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   controllerA,
					},
				},
				Contact: "contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedDymNames(ts, controllerA, recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      dymnstypes.DoNotModifyDesc,
				ClearConfigs: true,
				Controller:   controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com", // keep
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject message that neither update contact nor clear config",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      dymnstypes.DoNotModifyDesc,
				ClearConfigs: false,
				Controller:   controllerA,
			},
			wantErr:         true,
			wantErrContains: "message neither clears configs nor updates contact information",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject message that not update contact and clear config but no config to clear",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:      dymnstypes.DoNotModifyDesc,
				ClearConfigs: true,
				Controller:   controllerA,
			},
			wantErr:         true,
			wantErrContains: "no existing config to clear",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   now.Unix() + 1,
				Contact:    "contact@example.com",
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerA, recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerA)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
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
