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
	rollapps []*RollappCriteria,
	spendLimit sdk.Coins,
) *FulfillOrderAuthorization {
	return &FulfillOrderAuthorization{
		Rollapps:   rollapps,
		SpendLimit: spendLimit,
	}
}

func NewRollappCriteria(
	rollappID string,
	denoms []string,
	minLPFeePercentage sdk.DecProto,
	maxPrice sdk.Coins,
	fulfillerFeePart sdk.DecProto,
	settlementValidated bool,
) *RollappCriteria {
	return &RollappCriteria{
		RollappId:           rollappID,
		Denoms:              denoms,
		MinLpFeePercentage:  minLPFeePercentage,
		MaxPrice:            maxPrice,
		OperatorFeeShare:    fulfillerFeePart,
		SettlementValidated: settlementValidated,
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

	// Find the RollappCriteria matching the msg.RollappId
	var matchedCriteria *RollappCriteria
	for i, criteria := range a.Rollapps {
		if criteria.RollappId == mFulfill.RollappId {
			matchedCriteria = a.Rollapps[i] // Use pointer to modify if needed
			break
		}
	}

	if matchedCriteria == nil {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrUnauthorized, "rollapp %s is not authorized", mFulfill.RollappId)
	}

	// Check settlement_validated flag
	if matchedCriteria.SettlementValidated != mFulfill.SettlementValidated {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrUnauthorized, "settlement validation flag mismatch")
	}

	// Check operator_fee_share
	if !matchedCriteria.OperatorFeeShare.Dec.Equal(mFulfill.OperatorFeeShare.Dec) {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrUnauthorized, "operator fee share mismatch")
	}

	// Check denoms
	if len(matchedCriteria.Denoms) > 0 {
		for _, orderDenom := range mFulfill.Price.Denoms() {
			if !slices.Contains(matchedCriteria.Denoms, orderDenom) {
				return authz.AcceptResponse{},
					errorsmod.Wrapf(errors.ErrUnauthorized, "denom %s is not authorized", orderDenom)
			}
		}
	}

	// Check if the order fee meets the minimum LP fee percentage
	orderFeeDec, err := sdk.NewDecFromStr(mFulfill.ExpectedFee)
	if err != nil {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrInvalidCoins, "invalid fee amount: %s", err)
	}

	operatorFee := orderFeeDec.Mul(matchedCriteria.OperatorFeeShare.Dec)
	amountDec := mFulfill.Price[0].Amount.Add(orderFeeDec.RoundInt()).ToLegacyDec()
	minLPFee := amountDec.Mul(matchedCriteria.MinLpFeePercentage.Dec)
	lpFee := orderFeeDec.Sub(operatorFee)

	if lpFee.LT(minLPFee) {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrUnauthorized,
				"order LP fee %s is less than minimum LP fee %s",
				lpFee.String(), minLPFee.String())
	}

	// Check if the order price does not exceed the max price
	if !matchedCriteria.MaxPrice.IsZero() {
		orderPrice := mFulfill.Price
		if exceedsMaxPrice(orderPrice, matchedCriteria.MaxPrice) {
			return authz.AcceptResponse{},
				errorsmod.Wrapf(errors.ErrUnauthorized,
					"order price exceeds max price")
		}
	}

	// Check if spend limit is exhausted (spend_limit is at the top level)
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
	// Validate SpendLimit
	if a.SpendLimit != nil && !a.SpendLimit.IsValid() {
		return errorsmod.Wrapf(errors.ErrInvalidCoins, "spend_limit is invalid")
	}

	// Create a set to check for duplicate rollapp_ids
	rollappIDSet := make(map[string]struct{})

	for _, criteria := range a.Rollapps {
		// Validate that rollapp_id is not empty
		if len(criteria.RollappId) == 0 {
			return errorsmod.Wrapf(errors.ErrInvalidRequest, "rollapp_id cannot be empty")
		}

		// Check for duplicate rollapp_ids
		if _, exists := rollappIDSet[criteria.RollappId]; exists {
			return errorsmod.Wrapf(errors.ErrInvalidRequest, "duplicate rollapp_id %s in rollapps", criteria.RollappId)
		}
		rollappIDSet[criteria.RollappId] = struct{}{}

		// Validate MinLpFeePercentage
		if criteria.MinLpFeePercentage.Dec.IsNegative() {
			return errorsmod.Wrapf(errors.ErrInvalidRequest, "min_lp_fee_percentage cannot be negative for rollapp_id %s", criteria.RollappId)
		}

		// Validate OperatorFeeShare
		if criteria.OperatorFeeShare.Dec.IsNegative() || criteria.OperatorFeeShare.Dec.GT(sdk.OneDec()) {
			return errorsmod.Wrapf(errors.ErrInvalidRequest, "operator_fee_share must be between 0 and 1 for rollapp_id %s", criteria.RollappId)
		}

		// Validate MaxPrice (if provided)
		if criteria.MaxPrice != nil && !criteria.MaxPrice.IsValid() {
			return errorsmod.Wrapf(errors.ErrInvalidCoins, "max_price is invalid for rollapp_id %s", criteria.RollappId)
		}

		// Check for duplicates in Denoms
		if hasDuplicates(criteria.Denoms) {
			return errorsmod.Wrapf(errors.ErrInvalidRequest, "duplicate denoms in the list for rollapp_id %s", criteria.RollappId)
		}
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
