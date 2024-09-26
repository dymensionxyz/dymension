package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (s *KeeperTestSuite) Test_msgServer_UpdateDetails() {
	ownerAcc := testAddr(1)
	ownerA := ownerAcc.bech32()

	controllerAcc := testAddr(2)
	controllerA := controllerAcc.bech32()

	anotherAcc := testAddr(3)

	const recordName = "my-name"

	params.SetAddressPrefixes()

	tests := []struct {
		name               string
		dymName            *dymnstypes.DymName
		msg                *dymnstypes.MsgUpdateDetails
		preTestFunc        func(s *KeeperTestSuite)
		wantErr            bool
		wantErrContains    string
		wantDymName        *dymnstypes.DymName
		wantMinGasConsumed sdk.Gas
		postTestFunc       func(s *KeeperTestSuite)
	}{
		{
			name: "fail - reject if message not pass validate basic",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			msg: &dymnstypes.MsgUpdateDetails{},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
			wantErr:         true,
			wantErrContains: gerrc.ErrInvalidArgument.Error(),
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name:    "fail - Dym-Name does not exists",
			dymName: nil,
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
				Controller: controllerA,
			},
			wantErr:         true,
			wantErrContains: fmt.Sprintf("Dym-Name: %s: not found", recordName),
			wantDymName:     nil,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "fail - reject if Dym-Name expired",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() - 1,
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
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
				ExpireAt:   s.now.Unix() - 1,
			},
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "fail - reject if sender is neither owner nor controller",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
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
				ExpireAt:   s.now.Unix() + 100,
			},
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "fail - reject if sender is owner but not controller",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
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
				ExpireAt:   s.now.Unix() + 100,
			},
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "fail - reject if contact is not valid",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
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
				ExpireAt:   s.now.Unix() + 100,
			},
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "pass - can update contact",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "old-contact@example.com",
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "new-contact@example.com",
				Controller: controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "new-contact@example.com",
			},
			wantMinGasConsumed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "pass - can remove contact",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "old-contact@example.com",
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "", // clear
				Controller: controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "", // cleared
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "pass - can remove contact & configs",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Value: controllerA,
					},
				},
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).mappedDymNames(recordName)
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).mappedDymNames(recordName)
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
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "",
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "pass - can update contact & remove configs",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "old-contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Value: controllerA,
					},
				},
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).mappedDymNames(recordName)
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).mappedDymNames(recordName)
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
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "new-contact@example.com",
			},
			wantMinGasConsumed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "pass - expiry not changed",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 99,
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "contact@example.com",
				Controller: controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 99,
				Contact:    "contact@example.com",
			},
			wantMinGasConsumed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "pass - configs should not be changed when update contact and not clear config",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "old-contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Value: controllerA,
					},
				},
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).mappedDymNames(recordName)
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).mappedDymNames(recordName)
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
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "new-contact@example.com",
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Value: controllerA,
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasUpdateContact,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).mappedDymNames(recordName)
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).mappedDymNames(recordName)
			},
		},
		{
			name: "pass - reverse mapping record should be updated accordingly",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "",
						Path:    "",
						Value:   controllerA,
					},
					{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "nim_1122-1",
						Path:    "a",
						Value:   anotherAcc.bech32C("nim"),
					},
				},
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(anotherAcc.bech32C("nim")).mappedDymNames(recordName)
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(anotherAcc.fallback()).notMappedToAnyDymName()
			},
			msg: &dymnstypes.MsgUpdateDetails{
				ClearConfigs: true,
				Controller:   controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(anotherAcc.bech32C("nim")).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(anotherAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "pass - when contact is [do-not-modify], do not update contact",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "",
						Path:    "",
						Value:   controllerA,
					},
				},
				Contact: "contact@example.com",
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(controllerA).mappedDymNames(recordName)
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(controllerAcc.fallback()).mappedDymNames(recordName)
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
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "contact@example.com", // keep
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "fail - reject message that neither update contact nor clear config",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "contact@example.com",
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
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
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "contact@example.com",
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "fail - reject message that not update contact and clear config but no config to clear",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "contact@example.com",
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
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
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "contact@example.com",
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(recordName)
				s.requireConfiguredAddress(controllerA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(recordName)
				s.requireFallbackAddress(controllerAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name: "pass - independently charge gas",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "old-contact@example.com",
			},
			preTestFunc: func(s *KeeperTestSuite) {
				s.ctx.GasMeter().ConsumeGas(100_000_000, "simulate previous run")
			},
			msg: &dymnstypes.MsgUpdateDetails{
				Contact:    "new-contact@example.com",
				Controller: controllerA,
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: controllerA,
				ExpireAt:   s.now.Unix() + 100,
				Contact:    "new-contact@example.com",
			},
			wantMinGasConsumed: 100_000_000 + dymnstypes.OpGasUpdateContact,
			postTestFunc:       func(*KeeperTestSuite) {},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().NotNil(tt.preTestFunc)
			s.Require().NotNil(tt.postTestFunc)

			s.RefreshContext()

			if tt.dymName != nil {
				if tt.dymName.Name == "" {
					tt.dymName.Name = recordName
				}
				err := s.dymNsKeeper.SetDymName(s.ctx, *tt.dymName)
				s.Require().NoError(err)
				s.Require().NoError(s.dymNsKeeper.AfterDymNameOwnerChanged(s.ctx, tt.dymName.Name))
				s.Require().NoError(s.dymNsKeeper.AfterDymNameConfigChanged(s.ctx, tt.dymName.Name))
			}
			if tt.wantDymName != nil && tt.wantDymName.Name == "" {
				tt.wantDymName.Name = recordName
			}

			tt.preTestFunc(s)

			tt.msg.Name = recordName
			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).UpdateDetails(s.ctx, tt.msg)
			laterDymName := s.dymNsKeeper.GetDymName(s.ctx, tt.msg.Name)

			defer func() {
				if tt.wantMinGasConsumed > 0 {
					s.Require().GreaterOrEqual(
						s.ctx.GasMeter().GasConsumed(), tt.wantMinGasConsumed,
						"should consume at least %d gas", tt.wantMinGasConsumed,
					)
				}

				if !s.T().Failed() {
					tt.postTestFunc(s)
				}
			}()

			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)
				s.Require().Nil(resp)

				if tt.wantDymName != nil {
					s.Require().Equal(*tt.wantDymName, *laterDymName)

					owned, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, laterDymName.Owner)
					s.Require().NoError(err)
					if laterDymName.ExpireAt >= s.now.Unix() {
						s.Require().Len(owned, 1)
					} else {
						s.Require().Empty(owned)
					}
				} else {
					s.Require().Nil(laterDymName)
				}
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().NotNil(laterDymName)
			s.Require().Equal(*tt.wantDymName, *laterDymName)
		})
	}
}
