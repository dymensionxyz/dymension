package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestFulfillOrderAuthorization_Accept(t *testing.T) {
	type acceptTestCase struct {
		name           string
		authorization  FulfillOrderAuthorization
		msg            *MsgFulfillOrderAuthorized
		expectedAccept bool
		expectedDelete bool
		expectedError  string
		updatedAuth    *FulfillOrderAuthorization
	}
	testCases := []acceptTestCase{
		{
			name: "Valid Authorization Acceptance",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           "rollapp1",
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 500)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 1000)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.05")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           "rollapp1",
				Price:               sdk.NewCoins(sdk.NewInt64Coin("atom", 400)),
				ExpectedFee:         "23",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
				SettlementValidated: true,
			},
			expectedAccept: true,
			expectedDelete: false,
			updatedAuth: &FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           "rollapp1",
						Denoms:              []string{"atom"},
						MaxPrice:            sdk.NewCoins(sdk.NewInt64Coin("atom", 500)),
						SpendLimit:          sdk.NewCoins(sdk.NewInt64Coin("atom", 600)),
						MinFeePercentage:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.05")},
						OperatorFeeShare:    sdk.DecProto{Dec: sdk.MustNewDecFromStr("0.02")},
						SettlementValidated: true,
					},
				},
			},
		},
		{
			name: "Unauthorized Rollapp ID",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{RollappId: "rollapp1"},
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
						RollappId:           "rollapp1",
						SettlementValidated: true,
					},
				},
			},
			msg: &MsgFulfillOrderAuthorized{
				RollappId:           "rollapp1",
				SettlementValidated: false,
			},
			expectedAccept: false,
			expectedError:  "settlement validation flag mismatch",
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
	testCases := []validateBasicTestCase{
		{
			name: "Valid Authorization",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:           "rollapp1",
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
						RollappId:        "rollapp1",
						MinFeePercentage: sdk.DecProto{Dec: sdk.MustNewDecFromStr("-0.01")},
					},
				},
			},
			expectedError: "min_fee_percentage cannot be negative for rollapp_id rollapp1",
		},
		{
			name: "OperatorFeeShare Greater Than One",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:        "rollapp1",
						OperatorFeeShare: sdk.DecProto{Dec: sdk.MustNewDecFromStr("1.1")},
					},
				},
			},
			expectedError: "operator_fee_share must be between 0 and 1 for rollapp_id rollapp1",
		},
		{
			name: "Invalid SpendLimit",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId:  "rollapp1",
						SpendLimit: sdk.Coins{sdk.Coin{Denom: "atom", Amount: sdk.NewInt(-100)}},
					},
				},
			},
			expectedError: "spend_limit is invalid",
		},
		{
			name: "Duplicate Rollapp IDs",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{RollappId: "rollapp1"},
					{RollappId: "rollapp1"},
				},
			},
			expectedError: "duplicate rollapp_id rollapp1 in rollapps",
		},
		{
			name: "Empty Rollapp ID",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{RollappId: ""},
				},
			},
			expectedError: "rollapp_id cannot be empty",
		},
		{
			name: "Duplicate Denoms",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId: "rollapp1",
						Denoms:    []string{"atom", "atom"},
					},
				},
			},
			expectedError: "duplicate denoms in the list for rollapp_id rollapp1",
		},
		{
			name: "Invalid MaxPrice",
			authorization: FulfillOrderAuthorization{
				Rollapps: []*RollappCriteria{
					{
						RollappId: "rollapp1",
						MaxPrice:  sdk.Coins{sdk.Coin{Denom: "atom", Amount: sdk.NewInt(-500)}},
					},
				},
			},
			expectedError: "max_price is invalid for rollapp_id rollapp1",
		},
		// Add more test cases as needed...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			err := tc.authorization.ValidateBasic()

			// Verify
			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
