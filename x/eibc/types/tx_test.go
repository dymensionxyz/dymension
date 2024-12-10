package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgFulfillOrderAuthorized_ValidateBasic(t *testing.T) {
	validOrderID := "8f833734cf6b3890c386b8f7d0dc2c9ef077e8b1f3a8cf03874d37a316eb1308"
	validRollappID := "rollapp_1234-1"
	validPrice := sdk.NewCoins(sdk.NewInt64Coin("udenom", 100))
	negativePrice := sdk.Coins{sdk.Coin{Denom: "udenom", Amount: sdk.NewInt(-1)}}

	validAmount := sdk.IntProto{Int: sdk.NewInt(10)}
	nilAmount := sdk.IntProto{}                          // Int is nil
	zeroAmount := sdk.IntProto{Int: sdk.NewInt(0)}       // zero
	negativeAmount := sdk.IntProto{Int: sdk.NewInt(-10)} // negative

	validFeeShare := sdk.DecProto{Dec: sdk.NewDecWithPrec(5, 1)} // 0.5
	nilFeeShare := sdk.DecProto{}                                // nil dec
	negFeeShare := sdk.DecProto{Dec: sdk.NewDec(-1)}             // negative dec
	gtOneFeeShare := sdk.DecProto{Dec: sdk.NewDec(2)}            // >1

	validBech32 := "dym17g9cn4ss0h0dz5qhg2cg4zfnee6z3ftg3q6v58"
	invalidBech32 := "notanaddress"

	tests := []struct {
		name          string
		msg           MsgFulfillOrderAuthorized
		expectedError string
	}{
		{
			name: "empty rollapp id",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          "",
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "invalid rollapp id",
		},
		{
			name: "invalid rollapp id",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          "invalid_rollapp",
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "invalid rollapp id",
		},
		{
			name: "empty order id",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            "", // assume empty is invalid
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "Invalid order ID",
		},
		{
			name: "invalid order id",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            "invalid_order",
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "Invalid order ID",
		},
		{
			name: "invalid operator fee address",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: invalidBech32, // invalid
			},
			expectedError: "decoding bech32 failed",
		},
		{
			name: "invalid lp address",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          invalidBech32, // invalid
				OperatorFeeAddress: validBech32,
			},
			expectedError: "decoding bech32 failed",
		},
		{
			name: "invalid fee string",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "notanint",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "parse fee: notanint",
		},
		{
			name: "negative fee",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "-1",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "Fee must be greater than or equal to 0",
		},
		{
			name: "invalid price (negative coin)",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              negativePrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "price is invalid",
		},
		{
			name: "nil amount",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             nilAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "amount is invalid",
		},
		{
			name: "zero amount",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             zeroAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "amount is invalid",
		},
		{
			name: "negative amount",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             negativeAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "amount is invalid",
		},
		{
			name: "nil operator fee share",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   nilFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "operator fee share cannot be empty or negative",
		},
		{
			name: "negative operator fee share",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   negFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "operator fee share cannot be empty or negative",
		},
		{
			name: "operator fee share > 1",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   gtOneFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "operator fee share cannot be greater than 1",
		},
		{
			name: "all fields valid",
			msg: MsgFulfillOrderAuthorized{
				RollappId:          validRollappID,
				OrderId:            validOrderID,
				Price:              validPrice,
				Amount:             validAmount,
				OperatorFeeShare:   validFeeShare,
				ExpectedFee:        "100",
				LpAddress:          validBech32,
				OperatorFeeAddress: validBech32,
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}
