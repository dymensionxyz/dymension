package types

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
	return gs.Params.Validate()
}
