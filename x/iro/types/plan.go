package types

import (
	"errors"
	fmt "fmt"
	time "time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const IROTokenPrefix = "future"

var MinTokenAllocation = math.LegacyNewDec(10) // min allocation in decimal representation

func NewPlan(id uint64, rollappId string, allocation sdk.Coin, curve BondingCurve, start time.Time, end time.Time, incentivesParams IncentivePlanParams) Plan {
	plan := Plan{
		Id:                  id,
		RollappId:           rollappId,
		TotalAllocation:     allocation,
		BondingCurve:        curve,
		StartTime:           start,
		PreLaunchTime:       end,
		IncentivePlanParams: incentivesParams,
		SoldAmt:             math.ZeroInt(),
		ClaimedAmt:          math.ZeroInt(),
	}
	plan.ModuleAccAddress = authtypes.NewModuleAddress(plan.ModuleAccName()).String()
	return plan
}

// ValidateBasic checks if the plan is valid
func (p Plan) ValidateBasic() error {
	if err := p.BondingCurve.ValidateBasic(); err != nil {
		return errors.Join(ErrInvalidBondingCurve, err)
	}
	// check that the allocation is greater than the minimal allowed token allocation
	allocationDec := ScaleFromBase(p.TotalAllocation.Amount, p.BondingCurve.SupplyDecimals())
	if !allocationDec.GT(MinTokenAllocation) {
		return ErrInvalidAllocation
	}
	if p.PreLaunchTime.Before(p.StartTime) {
		return ErrInvalidEndTime
	}
	if p.ModuleAccAddress == "" {
		return fmt.Errorf("module account address cannot be empty")
	}
	if p.SoldAmt.IsNegative() {
		return fmt.Errorf("sold amount cannot be negative: %s", p.SoldAmt.String())
	}
	if p.ClaimedAmt.IsNegative() {
		return fmt.Errorf("claimed amount cannot be negative: %s", p.ClaimedAmt.String())
	}

	if err := p.IncentivePlanParams.ValidateBasic(); err != nil {
		return errors.Join(ErrInvalidIncentivePlanParams, err)
	}

	return nil
}

// SpotPrice returns the spot price of the plan
func (p Plan) SpotPrice() math.LegacyDec {
	return p.BondingCurve.SpotPrice(p.SoldAmt)
}

func (p Plan) IsSettled() bool {
	return p.SettledDenom != ""
}

func (p Plan) ModuleAccName() string {
	return ModuleName + "-" + p.RollappId
}

func (p Plan) GetAddress() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(p.ModuleAccAddress)
	return addr
}

func DefaultIncentivePlanParams() IncentivePlanParams {
	return IncentivePlanParams{
		NumEpochsPaidOver: 43200, // 1 month in minute epoch
	}
}

func (i IncentivePlanParams) ValidateBasic() error {
	if i.NumEpochsPaidOver == 0 {
		return fmt.Errorf("number of epochs paid over cannot be zero")
	}
	return nil
}
