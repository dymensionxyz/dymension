package types

import (
	"errors"
	fmt "fmt"
	time "time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func NewPlan(id uint64, rollappId string, allocation sdk.Coin, curve BondingCurve, start time.Time, end time.Time) Plan {
	plan := Plan{
		Id:              id,
		RollappId:       rollappId,
		TotalAllocation: allocation,
		BondingCurve:    curve,
		StartTime:       start,
		EndTime:         end,
		SoldAmt:         math.ZeroInt(),
		ClaimedAmt:      math.ZeroInt(),
	}
	plan.ModuleAccAddress = authtypes.NewModuleAddress(plan.ModuleAccName()).String()
	return plan
}

// ValidateBasic checks if the plan is valid
func (p Plan) ValidateBasic() error {
	if !p.TotalAllocation.IsPositive() {
		return ErrInvalidAllocation
	}
	if err := p.BondingCurve.ValidateBasic(); err != nil {
		return errors.Join(ErrInvalidBondingCurve, err)
	}
	if p.EndTime.Before(p.StartTime) {
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

	return nil
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
