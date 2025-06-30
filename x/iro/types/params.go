package types

import (
	fmt "fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// Default parameter values

var (
	DefaultTakerFee                                     = "0.02"                      // 2%
	DefaultCreationFee                                  = math.NewInt(1).MulRaw(1e18) /* 1 Rollapp token */
	DefaultMinPlanDuration                              = 0 * time.Hour               // no enforced minimum by default
	DefaultIncentivePlanMinimumNumEpochsPaidOver        = uint64(364)                 // default: min 364 days (based on 1 day distribution epoch)
	DefaultIncentivePlanMinimumStartTimeAfterSettlement = 60 * time.Minute            // default: min 1 hour after settlement
	DefaultMinLiquidityPart                             = "0.4"                       // default: at least 40% goes to the liquidity pool
	DefaultMinVestingDuration                           = 7 * 24 * time.Hour          // default: min 7 days
	DefaultMinVestingStartTimeAfterSettlement           = 0 * time.Minute             // default: no enforced minimum by default
)

// NewParams creates a new Params object
func NewParams(takerFee, liquidityPart math.LegacyDec, creationFee math.Int, minPlanDuration time.Duration, minIncentivePlanParams IncentivePlanParams, minVestingDuration, minVestingStartTimeAfterSettlement time.Duration) Params {
	return Params{
		TakerFee:                              takerFee,
		CreationFee:                           creationFee,
		MinPlanDuration:                       minPlanDuration,
		IncentivesMinStartTimeAfterSettlement: minIncentivePlanParams.StartTimeAfterSettlement,
		IncentivesMinNumEpochsPaidOver:        minIncentivePlanParams.NumEpochsPaidOver,
		MinLiquidityPart:                      liquidityPart,
		MinVestingDuration:                    minVestingDuration,
		MinVestingStartTimeAfterSettlement:    minVestingStartTimeAfterSettlement,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		TakerFee:                              math.LegacyMustNewDecFromStr(DefaultTakerFee),
		CreationFee:                           DefaultCreationFee,
		MinPlanDuration:                       DefaultMinPlanDuration,
		IncentivesMinStartTimeAfterSettlement: DefaultIncentivePlanMinimumStartTimeAfterSettlement,
		IncentivesMinNumEpochsPaidOver:        DefaultIncentivePlanMinimumNumEpochsPaidOver,
		MinLiquidityPart:                      math.LegacyMustNewDecFromStr(DefaultMinLiquidityPart),
		MinVestingDuration:                    DefaultMinVestingDuration,
		MinVestingStartTimeAfterSettlement:    DefaultMinVestingStartTimeAfterSettlement,
	}
}

// Validate checks that the parameters have valid values.
func (p Params) ValidateBasic() error {
	if err := validateTakerFee(p.TakerFee); err != nil {
		return err
	}

	if err := validateCreationFee(p.CreationFee); err != nil {
		return err
	}

	if p.MinPlanDuration < 0 {
		return fmt.Errorf("minimum plan duration must be non-negative: %v", p.MinPlanDuration)
	}

	if p.IncentivesMinNumEpochsPaidOver < 1 {
		return fmt.Errorf("incentive plan num epochs paid over must be greater than 0: %d", p.IncentivesMinNumEpochsPaidOver)
	}

	if p.IncentivesMinStartTimeAfterSettlement <= 0 {
		return fmt.Errorf("incentive plan start time after settlement must be greater than 0: %v", p.IncentivesMinStartTimeAfterSettlement)
	}

	if !p.MinLiquidityPart.IsPositive() {
		return fmt.Errorf("min liquidity part must be positive: %s", p.MinLiquidityPart)
	}

	if p.MinVestingDuration < 0 {
		return fmt.Errorf("minimum vesting duration must be non-negative: %v", p.MinVestingDuration)
	}

	return nil
}

func validateTakerFee(v math.LegacyDec) error {
	if v.IsNil() || v.IsNegative() {
		return fmt.Errorf("taker fee must be a non-negative decimal: %s", v)
	}

	if v.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("taker fee must be less than 1: %s", v)
	}

	return nil
}

func validateCreationFee(v math.Int) error {
	// creation fee must be a positive integer greater than 1^18 (1 Rollapp token)
	if v.LT(math.NewIntWithDecimal(1, 18)) {
		return fmt.Errorf("creation fee must be a positive integer: %s", v)
	}

	return nil
}

// CalcCreationFee is utility function that calculates the creation fee for a given curve and creation fee
// it's created to avoid zero cost for creation fee by increasing the fee amount until the cost is positive
// returns the cost and the fee amount
func CalcCreationFee(creationFee math.Int, curve BondingCurve) (math.Int, math.Int, error) {
	var cost math.Int
	feeAmt := creationFee
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		cost = curve.Cost(math.ZeroInt(), feeAmt)
		if cost.IsPositive() {
			break
		}
		feeAmt = feeAmt.MulRaw(100)
	}

	if !cost.IsPositive() {
		return math.Int{}, math.Int{}, errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid cost for fee charge")
	}

	return cost, feeAmt, nil
}
