package types

import (
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
