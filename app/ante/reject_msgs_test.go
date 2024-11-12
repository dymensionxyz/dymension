package ante_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/ante"
)

func (suite *AnteTestSuite) TestRejectMessagesDecoratorCustom() {
	suite.SetupTestCheckTx(false)

	decorator := ante.NewRejectMessagesDecorator().WithPredicate(ante.BlockTypeUrls(1, sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{})))

	{

		m := []sdk.Msg{
			// nest = 0 is OK
			&ibcclienttypes.MsgUpdateClient{},
		}
		tx := &mockTx{msgs: m}

		ctx := suite.ctx.WithBlockHeight(1)
		_, err := decorator.AnteHandle(ctx, tx, false, func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return ctx, nil })

		suite.NoError(err)
	}
	{
		m := []sdk.Msg{
			&authz.MsgExec{
				Grantee: "cosmos1...",
				Msgs: []*codectypes.Any{
					packMsg(suite.T(), &ibcclienttypes.MsgUpdateClient{}),
				},
			},
		}
		tx := &mockTx{msgs: m}

		ctx := suite.ctx.WithBlockHeight(1)
		_, err := decorator.AnteHandle(ctx, tx, false, func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return ctx, nil })

		suite.Error(err)
	}
}

func (suite *AnteTestSuite) TestRejectMessagesDecorator() {
	suite.SetupTestCheckTx(false)

	disabledMsgTypes := []string{
		sdk.MsgTypeURL(&banktypes.MsgSend{}),
		sdk.MsgTypeURL(&types.MsgDelegate{}),
	}

	decorator := ante.NewRejectMessagesDecorator().WithPredicate(ante.BlockTypeUrls(0, disabledMsgTypes...))

	testCases := []struct {
		name          string
		msgs          []sdk.Msg
		expectPass    bool
		expectedError string
	}{
		{
			name: "Transaction with direct disabled message (MsgSend)",
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: "cosmos1...",
					ToAddress:   "cosmos1...",
					Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
				},
			},
			expectPass:    false,
			expectedError: "/cosmos.bank.v1beta1.MsgSend",
		},
		{
			name: "Transaction with allowed message (MsgMultiSend)",
			msgs: []sdk.Msg{
				&banktypes.MsgMultiSend{
					Inputs:  []banktypes.Input{},
					Outputs: []banktypes.Output{},
				},
			},
			expectPass: true,
		},
		{
			name: "Transaction with disabled message nested in MsgExec",
			msgs: []sdk.Msg{
				&authz.MsgExec{
					Grantee: "cosmos1...",
					Msgs: []*codectypes.Any{
						packMsg(suite.T(), &banktypes.MsgSend{
							FromAddress: "cosmos1...",
							ToAddress:   "cosmos1...",
							Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
						}),
					},
				},
			},
			expectPass:    false,
			expectedError: "/cosmos.bank.v1beta1.MsgSend",
		},
		{
			name: "Transaction with allowed message nested in MsgExec",
			msgs: []sdk.Msg{
				&authz.MsgExec{
					Grantee: "cosmos1...",
					Msgs: []*codectypes.Any{
						packMsg(suite.T(), &banktypes.MsgMultiSend{
							Inputs:  []banktypes.Input{},
							Outputs: []banktypes.Output{},
						}),
					},
				},
			},
			expectPass: true,
		},
		{
			name: "Transaction with disabled message in gov v1 MsgSubmitProposal",
			msgs: []sdk.Msg{
				&govtypesv1.MsgSubmitProposal{
					Messages: []*codectypes.Any{
						packMsg(suite.T(), &banktypes.MsgSend{
							FromAddress: "cosmos1...",
							ToAddress:   "cosmos1...",
							Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
						}),
					},
					InitialDeposit: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
					Proposer:       "cosmos1...",
				},
			},
			expectPass:    false,
			expectedError: "/cosmos.bank.v1beta1.MsgSend",
		},
		{
			name: "Transaction exceeding max nested messages",
			msgs: []sdk.Msg{
				generateDeeplyNestedMsgExec(suite.T(), 7), // exceeds maxDepth (6)
			},
			expectPass:    false,
			expectedError: "found more nested msgs than permitted. Limit is : 6",
		},
		{
			name: "Transaction with authz.MsgGrant granting disabled message",
			msgs: []sdk.Msg{
				&authz.MsgGrant{
					Granter: "cosmos1...",
					Grantee: "cosmos1...",
					Grant: authz.Grant{
						Authorization: packAuthorization(suite.T(), &authz.GenericAuthorization{
							Msg: sdk.MsgTypeURL(&banktypes.MsgSend{}),
						}),
						Expiration: nil,
					},
				},
			},
			expectPass:    false,
			expectedError: "/cosmos.bank.v1beta1.MsgSend",
		},
		{
			name: "Transaction with authz.MsgGrant granting allowed message",
			msgs: []sdk.Msg{
				&authz.MsgGrant{
					Granter: "cosmos1...",
					Grantee: "cosmos1...",
					Grant: authz.Grant{
						Authorization: packAuthorization(suite.T(), &authz.GenericAuthorization{
							Msg: sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
						}),
						Expiration: nil,
					},
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tx := &mockTx{msgs: tc.msgs}

			ctx := suite.ctx.WithBlockHeight(1)
			_, err := decorator.AnteHandle(ctx, tx, false, func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return ctx, nil })

			if tc.expectPass {
				suite.NoError(err, "Test case %s failed unexpectedly", tc.name)
			} else {
				suite.Error(err, "Test case %s expected error but got none", tc.name)
				suite.Contains(err.Error(), tc.expectedError, "Test case %s error message mismatch", tc.name)
			}
		})
	}
}

func packMsg(t *testing.T, msg sdk.Msg) *codectypes.Any {
	a, err := codectypes.NewAnyWithValue(msg)
	require.NoError(t, err)
	return a
}

func packAuthorization(t *testing.T, authorization authz.Authorization) *codectypes.Any {
	a, err := codectypes.NewAnyWithValue(authorization)
	require.NoError(t, err)
	return a
}

type mockTx struct {
	msgs []sdk.Msg
}

func (tx *mockTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

func (tx *mockTx) ValidateBasic() error {
	return nil
}

func generateDeeplyNestedMsgExec(t *testing.T, depth int) sdk.Msg {
	if depth <= 0 {
		return &banktypes.MsgMultiSend{
			Inputs:  []banktypes.Input{},
			Outputs: []banktypes.Output{},
		}
	}

	innerMsg := generateDeeplyNestedMsgExec(t, depth-1)
	anyMsg := packMsg(t, innerMsg)

	return &authz.MsgExec{
		Grantee: "cosmos1...",
		Msgs:    []*codectypes.Any{anyMsg},
	}
}
