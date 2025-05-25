package ante_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"fmt"

	"github.com/dymensionxyz/dymension/v3/app/ante"
)

const dummyKey1 = "test_key_1"
const dummyValue1 = "test_value_1"
const dummyKey2 = "test_key_2"
const dummyValue2 = "test_value_2"

// Dummy InnerCallback that returns error if depth > 1
func dummyErrorOnDepthCallback(ctx sdk.Context, msg sdk.Msg, simulate bool, depth int) (sdk.Context, error) {
	if depth > 1 {
		return ctx, fmt.Errorf("depth %d is too deep", depth)
	}
	return ctx, nil
}

// Dummy InnerCallback that sets dummy value 1
func dummyInnerCallback1(ctx sdk.Context, msg sdk.Msg, simulate bool, depth int) (sdk.Context, error) {
	return ctx.WithValue(dummyKey1, dummyValue1), nil
}

// Dummy InnerCallback that sets dummy value 2
func dummyInnerCallback2(ctx sdk.Context, msg sdk.Msg, simulate bool, depth int) (sdk.Context, error) {
	return ctx.WithValue(dummyKey2, dummyValue2), nil
}

func (suite *AnteTestSuite) TestInnerDecoratorDummyCallback() {
	suite.SetupTestCheckTx(false)

	decorator := ante.NewInnerDecorator(dummyInnerCallback1)

	testCases := []struct {
		name string
		msgs []sdk.Msg
	}{
		{
			name: "Simple message",
			msgs: []sdk.Msg{
				&banktypes.MsgMultiSend{},
			},
		},
		{
			name: "Message wrapped in MsgExec",
			msgs: []sdk.Msg{
				&authz.MsgExec{
					Grantee: "cosmos1...",
					Msgs: []*codectypes.Any{
						packMsg(suite.T(), &banktypes.MsgMultiSend{}),
					},
				},
			},
		},
		{
			name: "Deeply nested MsgExec",
			msgs: []sdk.Msg{
				generateDeeplyNestedMsgExec(suite.T(), 3),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tx := &mockTx{msgs: tc.msgs}
			ctx := suite.ctx.WithBlockHeight(1)
			ctxOut, err := decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
				return ctx, nil
			})
			suite.Require().NoError(err)
			val := ctxOut.Value(dummyKey1)
			suite.Require().Equal(dummyValue1, val, "context should contain the dummy value after InnerDecorator runs")
		})
	}
}

func (suite *AnteTestSuite) TestInnerDecoratorErrorOnDepth() {
	suite.SetupTestCheckTx(false)

	decorator := ante.NewInnerDecorator(dummyErrorOnDepthCallback)

	testCases := []struct {
		name        string
		msgs        []sdk.Msg
		expectError bool
	}{
		{
			name: "Simple message (depth 0)",
			msgs: []sdk.Msg{
				&banktypes.MsgMultiSend{},
			},
			expectError: false,
		},
		{
			name: "Message wrapped in MsgExec (depth 1)",
			msgs: []sdk.Msg{
				&authz.MsgExec{
					Grantee: "cosmos1...",
					Msgs: []*codectypes.Any{
						packMsg(suite.T(), &banktypes.MsgMultiSend{}),
					},
				},
			},
			expectError: false,
		},
		{
			name: "Deeply nested MsgExec (depth 2)",
			msgs: []sdk.Msg{
				generateDeeplyNestedMsgExec(suite.T(), 2),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tx := &mockTx{msgs: tc.msgs}
			ctx := suite.ctx.WithBlockHeight(1)
			_, err := decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
				return ctx, nil
			})
			if tc.expectError {
				suite.Require().Error(err, "expected error for test case: %s", tc.name)
				suite.Require().Contains(err.Error(), "depth", "error message should mention depth")
			} else {
				suite.Require().NoError(err, "unexpected error for test case: %s", tc.name)
			}
		})
	}
}

func (suite *AnteTestSuite) TestInnerDecoratorMultipleCallbacks() {
	suite.SetupTestCheckTx(false)

	decorator := ante.NewInnerDecorator(dummyInnerCallback1, dummyInnerCallback2, dummyErrorOnDepthCallback)

	testCases := []struct {
		name        string
		msgs        []sdk.Msg
		expectError bool
	}{
		{
			name: "Simple message",
			msgs: []sdk.Msg{
				&banktypes.MsgMultiSend{},
			},
			expectError: false,
		},
		{
			name: "Message wrapped in MsgExec",
			msgs: []sdk.Msg{
				&authz.MsgExec{
					Grantee: "cosmos1...",
					Msgs: []*codectypes.Any{
						packMsg(suite.T(), &banktypes.MsgMultiSend{}),
					},
				},
			},
			expectError: false,
		},
		{
			name: "Deeply nested MsgExec (depth 2)",
			msgs: []sdk.Msg{
				generateDeeplyNestedMsgExec(suite.T(), 2),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tx := &mockTx{msgs: tc.msgs}
			ctx := suite.ctx.WithBlockHeight(1)
			ctxOut, err := decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
				return ctx, nil
			})
			if tc.expectError {
				suite.Require().Error(err, "expected error for test case: %s", tc.name)
				suite.Require().Contains(err.Error(), "depth", "error message should mention depth")
				// dummy value 1 should still be set before error
				val1 := ctxOut.Value(dummyKey1)
				suite.Require().Equal(dummyValue1, val1, "context should contain dummy value 1 even if error occurs")
			} else {
				suite.Require().NoError(err, "unexpected error for test case: %s", tc.name)
				val1 := ctxOut.Value(dummyKey1)
				val2 := ctxOut.Value(dummyKey2)
				suite.Require().Equal(dummyValue1, val1, "context should contain dummy value 1 after InnerDecorator runs")
				suite.Require().Equal(dummyValue2, val2, "context should contain dummy value 2 after InnerDecorator runs")
			}
		})
	}
}

func (suite *AnteTestSuite) TestInnerDecoratorCountCallbackCalls() {
	suite.SetupTestCheckTx(false)

	var callCount int
	countingCallback := func(ctx sdk.Context, msg sdk.Msg, simulate bool, depth int) (sdk.Context, error) {
		callCount++
		return ctx, nil
	}

	// we call the callback twice, just to make sure it's called twice per message
	decorator := ante.NewInnerDecorator(countingCallback, countingCallback)

	testCases := []struct {
		name          string
		msgs          []sdk.Msg
		expectedCalls int
	}{
		{
			name: "Single message",
			msgs: []sdk.Msg{
				&banktypes.MsgMultiSend{},
			},
			expectedCalls: 2, // 1 msg * 2 callbacks
		},
		{
			name: "Two top-level messages",
			msgs: []sdk.Msg{
				&banktypes.MsgMultiSend{},
				&banktypes.MsgMultiSend{},
			},
			expectedCalls: 4, // 2 msgs * 2 callbacks
		},
		{
			name: "MsgExec wrapping two messages",
			msgs: []sdk.Msg{
				&authz.MsgExec{
					Grantee: "cosmos1...",
					Msgs: []*codectypes.Any{
						packMsg(suite.T(), &banktypes.MsgMultiSend{}),
						packMsg(suite.T(), &banktypes.MsgMultiSend{}),
					},
				},
			},
			expectedCalls: 4, // 2 msgs * 2 callbacks
		},
		{
			name: "Deeply nested MsgExec (depth 3)",
			msgs: []sdk.Msg{
				generateDeeplyNestedMsgExec(suite.T(), 3),
			},
			expectedCalls: 2, // 1 msg * 2 callbacks
		},
		{
			name: "MsgExec with two nested MsgExecs",
			msgs: []sdk.Msg{
				&authz.MsgExec{
					Grantee: "cosmos1...",
					Msgs: []*codectypes.Any{
						packMsg(suite.T(), generateDeeplyNestedMsgExec(suite.T(), 2)),
						packMsg(suite.T(), generateDeeplyNestedMsgExec(suite.T(), 1)),
					},
				},
			},
			expectedCalls: 4, // 2 msgs * 2 callbacks
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			callCount = 0
			tx := &mockTx{msgs: tc.msgs}
			ctx := suite.ctx.WithBlockHeight(1)
			_, err := decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
				return ctx, nil
			})
			suite.Require().NoError(err, "unexpected error for test case: %s", tc.name)
			suite.Require().Equal(tc.expectedCalls, callCount, "unexpected number of callback calls for test case: %s", tc.name)
		})
	}
}
