package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	demandOrdersMap := make(map[string]struct{})
	for _, demandOrder := range gs.GetDemandOrders() {
		if err := demandOrder.Validate(); err != nil {
			return err
		}
		if _, ok := demandOrdersMap[demandOrder.Id]; ok {
			return ErrDemandOrderAlreadyExist
		}
		demandOrdersMap[demandOrder.Id] = struct{}{}
	}
	lpIDs := make(map[uint64]struct{})
	for _, lp := range gs.OnDemandLps {
		if err := lp.Validate(); err != nil {
			return err
		}
		if _, ok := lpIDs[lp.Id]; ok {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "duplicate lp id: %d", lp.Id)
		}
		if lp.Id >= gs.LpsNextId {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "lp id not below next id: %d", lp.Id)
		}
		lpIDs[lp.Id] = struct{}{}
	}
	return gs.Params.ValidateBasic()
}
