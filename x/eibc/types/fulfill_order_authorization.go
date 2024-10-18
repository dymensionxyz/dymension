package types

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// NewFulfillOrderAuthorization creates a new FulfillOrderAuthorization object.
func NewFulfillOrderAuthorization(
	rollapps []string,
	denoms []string,
	minLPFeePercentage sdk.DecProto,
	maxPrice sdk.Coins,
	fulfillerFeePart sdk.DecProto,
	settlementValidated bool,
	spendLimit sdk.Coins,
) *FulfillOrderAuthorization {
	return &FulfillOrderAuthorization{
		Rollapps:            rollapps,
		Denoms:              denoms,
		MinLpFeePercentage:  minLPFeePercentage,
		MaxPrice:            maxPrice,
		OperatorFeeShare:    fulfillerFeePart,
		SettlementValidated: settlementValidated,
		SpendLimit:          spendLimit,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a FulfillOrderAuthorization) MsgTypeURL() string {
	return "/dymensionxyz.dymension.eibc.MsgFulfillOrderAuthorized"
}

// Accept implements Authorization.Accept.
func (a FulfillOrderAuthorization) Accept(
	_ sdk.Context,
	msg sdk.Msg,
) (authz.AcceptResponse, error) {
	mFulfill, ok := msg.(*MsgFulfillOrderAuthorized)
	if !ok {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrInvalidType,
				"type mismatch: expected %T, got %T",
				&MsgFulfillOrderAuthorized{}, msg)
	}

	// Check if the settlement validation flag matches
	if a.SettlementValidated != mFulfill.SettlementValidated {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrUnauthorized, "settlement validation flag mismatch")
	}

	// Check if the fulfiller fee part matches
	if !a.OperatorFeeShare.Dec.Equal(mFulfill.OperatorFeeShare.Dec) {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrUnauthorized, "fulfiller fee part mismatch")
	}

	// Check if the rollapp is allowed
	if len(a.Rollapps) > 0 && !slices.Contains(a.Rollapps, mFulfill.RollappId) {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrUnauthorized, "rollapp %s is not authorized", mFulfill.RollappId)
	}

	// Check if the denom is allowed
	if len(a.Denoms) > 0 {
		for _, orderDenom := range mFulfill.Price.Denoms() {
			if !slices.Contains(a.Denoms, orderDenom) {
				return authz.AcceptResponse{},
					errorsmod.Wrapf(errors.ErrUnauthorized, "denom %s is not authorized", orderDenom)
			}
		}
	}

	// Check if the order fee meets the minimum fee
	orderFeeDec, err := sdk.NewDecFromStr(mFulfill.ExpectedFee)
	if err != nil {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrInvalidCoins, "invalid fee amount: %s", err)
	}

	operatorFee := orderFeeDec.Mul(a.OperatorFeeShare.Dec)
	amountDec := mFulfill.Price[0].Amount.Add(orderFeeDec.RoundInt()).ToLegacyDec()
	minLPFee := amountDec.Mul(a.MinLpFeePercentage.Dec)
	lpFee := orderFeeDec.Sub(operatorFee)

	if lpFee.LT(minLPFee) {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrUnauthorized,
				"order lp fee %s is less than minimum lp fee %s",
				lpFee.String(), minLPFee.String())
	}

	// Check if the order price does not exceed the max price
	if !a.MaxPrice.IsZero() {
		orderPrice := mFulfill.Price
		if exceedsMaxPrice(orderPrice, a.MaxPrice) {
			return authz.AcceptResponse{},
				errorsmod.Wrapf(errors.ErrUnauthorized,
					"order price exceeds max price")
		}
	}

	// Check if spend limit is exhausted
	if !a.SpendLimit.IsZero() {
		spendLeft, isNegative := a.SpendLimit.SafeSub(mFulfill.Price...)
		if isNegative {
			return authz.AcceptResponse{},
				errorsmod.Wrapf(errors.ErrInsufficientFunds,
					"spend limit exhausted")
		}
		if spendLeft.IsZero() {
			return authz.AcceptResponse{
				Accept: true,
				Delete: true,
			}, nil
		}
		// Update the authorization with the new spend limit
		a.SpendLimit = spendLeft
		return authz.AcceptResponse{
			Accept:  true,
			Delete:  false,
			Updated: &a,
		}, nil
	}

	// If all checks pass and there's no spend limit
	return authz.AcceptResponse{
		Accept: true,
		Delete: false,
	}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a FulfillOrderAuthorization) ValidateBasic() error {
	// Validate MinFee
	if a.MinLpFeePercentage.Dec.IsNegative() {
		return errorsmod.Wrapf(errors.ErrInvalidRequest,
			"min_lp_fee cannot be negative")
	}

	// Validate OperatorFeeShare
	if a.OperatorFeeShare.Dec.IsNegative() ||
		a.OperatorFeeShare.Dec.GT(sdk.OneDec()) {
		return errorsmod.Wrapf(errors.ErrInvalidRequest,
			"operator_fee_share must be between 0 and 1")
	}

	// Validate SpendLimit
	if a.SpendLimit != nil && !a.SpendLimit.IsValid() {
		return errorsmod.Wrapf(errors.ErrInvalidCoins,
			"spend_limit is invalid")
	}

	// Check for duplicate entries in Rollapps and Denoms
	if hasDuplicates(a.Rollapps) {
		return errorsmod.Wrapf(errors.ErrInvalidRequest,
			"duplicate rollapps in the list")
	}
	if hasDuplicates(a.Denoms) {
		return errorsmod.Wrapf(errors.ErrInvalidRequest,
			"duplicate denoms in the list")
	}

	return nil
}

// Helper function to check for duplicates in a slice
func hasDuplicates(list []string) bool {
	seen := make(map[string]bool)
	for _, v := range list {
		if seen[v] {
			return true
		}
		seen[v] = true
	}
	return false
}

// Helper function to check if order price exceeds max price
func exceedsMaxPrice(orderPrice, maxPrice sdk.Coins) bool {
	for _, coin := range orderPrice {
		maxCoin := maxPrice.AmountOf(coin.Denom)
		if !maxCoin.IsZero() && coin.Amount.GT(maxCoin) {
			return true
		}
	}
	return false
}
