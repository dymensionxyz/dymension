package types

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// NewFulfillOrderAuthorization creates a new FulfillOrderAuthorization object.
func NewFulfillOrderAuthorization(rollapps []*RollappCriteria) *FulfillOrderAuthorization {
	return &FulfillOrderAuthorization{
		Rollapps: rollapps,
	}
}

func NewRollappCriteria(
	rollappID string,
	denoms []string,
	minFeePercentage sdk.DecProto,
	maxPrice sdk.Coins,
	spendLimit sdk.Coins,
	fulfillerFeePart sdk.DecProto,
	settlementValidated bool,
) *RollappCriteria {
	return &RollappCriteria{
		RollappId:           rollappID,
		Denoms:              denoms,
		MinFeePercentage:    minFeePercentage,
		MaxPrice:            maxPrice,
		SpendLimit:          spendLimit,
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
	for i := range a.Rollapps {
		if a.Rollapps[i].RollappId == mFulfill.RollappId {
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

	minFee := sdk.NewDecFromInt(mFulfill.Amount.Int).Mul(matchedCriteria.MinFeePercentage.Dec)

	if orderFeeDec.LT(minFee) {
		return authz.AcceptResponse{},
			errorsmod.Wrapf(errors.ErrUnauthorized,
				"order fee %s is less than minimum fee %s",
				orderFeeDec.String(), minFee.String())
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

	// Check if spend limit is exhausted (spend_limit is now in matchedCriteria)
	if !matchedCriteria.SpendLimit.IsZero() {
		spendLeft, isNegative := matchedCriteria.SpendLimit.SafeSub(mFulfill.Price...)
		if isNegative {
			return authz.AcceptResponse{},
				errorsmod.Wrapf(errors.ErrInsufficientFunds,
					"spend limit exhausted for rollapp %s", mFulfill.RollappId)
		}

		// Update the spend limit in matchedCriteria
		matchedCriteria.SpendLimit = spendLeft

		// If spendLeft is zero, remove the matchedCriteria
		if spendLeft.IsZero() {
			a.removeRollappCriteria(mFulfill.RollappId)
		}

		// If all rollapps are exhausted, delete the authorization
		if len(a.Rollapps) == 0 {
			return authz.AcceptResponse{
				Accept: true,
				Delete: true,
			}, nil
		}

		// Return the updated authorization
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

func (a *FulfillOrderAuthorization) removeRollappCriteria(rollappId string) {
	for i, criteria := range a.Rollapps {
		if criteria.RollappId == rollappId {
			// Remove the criteria from the slice
			a.Rollapps = append(a.Rollapps[:i], a.Rollapps[i+1:]...)
			return
		}
	}
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a FulfillOrderAuthorization) ValidateBasic() error {
	// Create a set to check for duplicate rollapp_ids
	rollappIDSet := make(map[string]struct{})

	for _, criteria := range a.Rollapps {
		// Validate that rollapp_id is not empty
		if err := validateRollappID(criteria.RollappId); err != nil {
			return errorsmod.Wrapf(errors.ErrInvalidRequest, err.Error())
		}

		// Check for duplicate rollapp_ids
		if _, exists := rollappIDSet[criteria.RollappId]; exists {
			return errorsmod.Wrapf(errors.ErrInvalidRequest, "duplicate rollapp_id %s in rollapps", criteria.RollappId)
		}
		rollappIDSet[criteria.RollappId] = struct{}{}

		// Validate MinFeePercentage
		if criteria.MinFeePercentage.Dec.IsNil() || criteria.MinFeePercentage.Dec.IsNegative() || criteria.MinFeePercentage.Dec.GT(sdk.OneDec()) {
			return errorsmod.Wrapf(errors.ErrInvalidRequest, "min_fee_percentage must be between 0 and 1 for rollapp_id %s", criteria.RollappId)
		}

		// Validate OperatorFeeShare
		if criteria.OperatorFeeShare.Dec.IsNil() || criteria.OperatorFeeShare.Dec.IsNegative() || criteria.OperatorFeeShare.Dec.GT(sdk.OneDec()) {
			return errorsmod.Wrapf(errors.ErrInvalidRequest, "operator_fee_share must be between 0 and 1 for rollapp_id %s", criteria.RollappId)
		}

		// Validate MaxPrice (if provided)
		if criteria.MaxPrice != nil && !criteria.MaxPrice.IsValid() {
			return errorsmod.Wrapf(errors.ErrInvalidCoins, "max_price is invalid for rollapp_id %s", criteria.RollappId)
		}

		// Validate SpendLimit
		if criteria.SpendLimit != nil && !criteria.SpendLimit.IsValid() {
			return errorsmod.Wrapf(errors.ErrInvalidCoins, "spend_limit is invalid")
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
