package types

import "github.com/dymensionxyz/dymension/v3/x/common/types"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	rollappPacketMap := make(map[string]struct{})
	for _, rollappPacket := range gs.GetRollappPackets() {
		if err := rollappPacket.ValidateBasic(); err != nil {
			return err
		}
		if _, ok := rollappPacketMap[string(types.RollappPacketKey(&rollappPacket))]; ok {
			return ErrRollappPacketAlreadyExists
		}
		rollappPacketMap[string(types.RollappPacketKey(&rollappPacket))] = struct{}{}
	}
	return gs.Params.Validate()
}
