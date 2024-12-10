package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestFulfillOrderAuthorization_Accept(t *testing.T) {
	type acceptTestCase struct {
		name           string
		authorization  FulfillOrderAuthorization
		msg            sdk.Msg
		expectedAccept bool
		expectedDelete bool
		expectedError  string
		updatedAuth    *FulfillOrderAuthorization
	}
	validRollappID1 := "rollappa_1234-1"
	validRollappID2 := "rollappb_2345-1"
	testCases := []acceptTestCase{
		{
			name: "Valid Authorization Acceptance",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 9986)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         "15",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: true,
			expectedDelete: false,
			updatedAuth: &FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 1)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
		},
		{
			name: "Valid Authorization Acceptance: two rollapps - preserve authorization",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
					{
						RollappId:           validRollappID2,
						Denoms:              []string{"btc"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("btc", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("btc", 9986)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         "15",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: true,
			expectedDelete: false,
			updatedAuth: &FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID2,
						Denoms:              []string{"btc"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("btc", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("btc", 9986)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
		},
		{
			name: "Invalid Message Type",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{RollappId: validRollappID1},
				},
			},
			msg: &MsgFulfillOrder{
				OrderId: "order1",
			},
			expectedAccept: false,
			expectedError:  "type mismatch: expected *types.MsgFulfillOrderAuthorized, got *types.MsgFulfillOrder",
		},
		{
			name: "Unauthorized Rollapp ID",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{RollappId: validRollappID1},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId: "unauthorized_rollapp",
			},
			expectedAccept: false,
			expectedError:  "rollapp unauthorized_rollapp is not authorized",
		},
		{
			name: "Settlement Validation Mismatch",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				SettlementValidated: false,
			},
			expectedAccept: false,
			expectedError:  "settlement validation flag mismatch",
		},
		{
			name: "Operator fee share mismatch",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 9986)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         "15",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.03")},
				SettlementValidated: true,
			},
			expectedAccept: false,
			expectedError:  "operator fee share mismatch",
		},
		{
			name: "Unauthorized Denom",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 9986)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("btc", 9985)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         "15",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: false,
			expectedError:  "denom btc is not authorized",
		},
		{
			name: "Invalid fee amount",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 9986)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         "invalid",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: false,
			expectedError:  "invalid fee amount",
		},
		{
			name: "Expected fee lower than minimum fee",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 9986)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         sdk.MustNewDecFromStr("0.0014").Mul(sdk.NewDec(10000)).String(),
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: false,
			expectedError:  "is less than minimum fee",
		},
		{
			name: "Exceeds Max Price",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 9986)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 9986)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         "15",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: false,
			expectedError:  "exceeds max price",
		},
		{
			name: "Spend Limit Exhausted",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 9984)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         "15",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: false,
			expectedError:  fmt.Sprintf("spend limit exhausted for rollapp %s", validRollappID1),
			expectedDelete: true,
		},
		{
			name: "All Rollapps Exhausted",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         "15",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: true,
			expectedDelete: true,
		},
		{
			name: "No Spend Limit",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validRollappID1,
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.0015")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           validRollappID1,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 9985)),
				Amount:              sdk.IntProto{Int: sdk.NewInt(10000)},
				ExpectedFee:         "15",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: true,
			expectedDelete: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.authorization.Accept(sdk.Context{}, tc.msg)

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
				require.False(t, resp.Accept)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedAccept, resp.Accept)
				require.Equal(t, tc.expectedDelete, resp.Delete)

				if tc.updatedAuth != nil {
					require.NotNil(t, resp.Updated)
					require.Equal(t, tc.updatedAuth, resp.Updated)
				} else {
					require.Nil(t, resp.Updated)
				}
			}
		})
	}
}

func TestFulfillOrderAuthorization_ValidateBasic(t *testing.T) {
	type validateBasicTestCase struct {
		name          string
		authorization FulfillOrderAuthorization
		expectedError string
	}
	validaRollappID := "rollapp_1234-1"
	testCases := []validateBasicTestCase{
		{
			name: "Valid Authorization",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           validaRollappID,
						Denoms:              []string{"atom", "btc"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 500)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 1000)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.05")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			expectedError: "",
		},
		{
			name: "Negative MinFeePercentage",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:        validaRollappID,
						MinFeePercentage: sdk.DecProto{Dec: sdk.MustNewDecFromStr("-0.01")},
					},
				},
			},
			expectedError: fmt.Sprintf("min_fee_percentage must be between 0 and 1 for rollapp_id %s", validaRollappID),
		},
		{
			name: "OperatorFeeShare Greater Than One",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:        validaRollappID,
						MinFeePercentage: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.01")},
						OperatorFeeShare: sdk.DecProto{Dec: sdk.MustNewDecFromStr("1.1")},
					},
				},
			},
			expectedError: fmt.Sprintf("operator_fee_share must be between 0 and 1 for rollapp_id %s", validaRollappID),
		},
		{
			name: "Invalid SpendLimit",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:        validaRollappID,
						MinFeePercentage: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.01")},
						OperatorFeeShare: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.1")},
						SpendLimit:       sdk.Coins{sdk.Coin{Denom: "atom", Amount: sdk.NewInt(-100)}},
					},
				},
			},
			expectedError: "spend_limit is invalid",
		},
		{
			name: "Duplicate Rollapp IDs",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:        validaRollappID,
						MinFeePercentage: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.01")},
						OperatorFeeShare: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.1")},
					},
					{
						RollappId:        validaRollappID,
						MinFeePercentage: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.01")},
						OperatorFeeShare: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.1")},
					},
				},
			},
			expectedError: fmt.Sprintf("duplicate rollapp_id %s in rollapps", validaRollappID),
		},
		{
			name: "Empty Rollapp ID",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{RollappId: ""},
				},
			},
			expectedError: "invalid rollapp-id",
		},
		{
			name: "Invalid Rollapp ID",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{RollappId: "invalid-rollapp-id"},
				},
			},
			expectedError: "invalid rollapp-id",
		},
		{
			name: "Duplicate Denoms",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:        validaRollappID,
						MinFeePercentage: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.01")},
						OperatorFeeShare: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.1")},
						Denoms:           []string{"atom", "atom"},
					},
				},
			},
			expectedError: fmt.Sprintf("duplicate denoms in the list for rollapp_id %s", validaRollappID),
		},
		{
			name: "Invalid MaxPrice",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:        validaRollappID,
						MinFeePercentage: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.01")},
						OperatorFeeShare: sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.1")},
						MaxPrice:         sdk.Coins{sdk.Coin{Denom: "atom", Amount: sdk.NewInt(-500)}},
					},
				},
			},
			expectedError: fmt.Sprintf("max_price is invalid for rollapp_id %s", validaRollappID),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.authorization.ValidateBasic()

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
