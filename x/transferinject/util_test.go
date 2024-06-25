package transferinject_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type mockRollappKeeper struct {
	rollapp  *rollapptypes.Rollapp
	transfer *transfertypes.FungibleTokenPacketData
	err      error
}

func (m *mockRollappKeeper) GetValidTransfer(ctx sdk.Context, packetData []byte, raPortOnHub, raChanOnHub string) (data rollapptypes.TransferData, err error) {
	ret := rollapptypes.TransferData{}
	if m.transfer != nil {
		ret.FungibleTokenPacketData = *m.transfer
	}
	if m.rollapp != nil {
		ret.Rollapp = m.rollapp
	}
	return ret, nil
}

func (m *mockRollappKeeper) SetRollapp(_ sdk.Context, rollapp rollapptypes.Rollapp) {
	m.rollapp = &rollapp
}
