package transferinject_test

import (
	"fmt"
	"testing"

	"github.com/dymensionxyz/dymension/v3/utils/gerr"

	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/stretchr/testify/require"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/transferinject"
	"github.com/dymensionxyz/dymension/v3/x/transferinject/types"
)

func TestIBCModule_OnAcknowledgementPacket(t *testing.T) {
	type fields struct {
		packetData    *transfertypes.FungibleTokenPacketData
		ack           []byte
		ibcModule     porttypes.IBCModule
		rollappKeeper *mockRollappKeeper
	}
	tests := []struct {
		name        string
		fields      fields
		wantRollapp *rollapptypes.Rollapp
		wantErr     error
	}{
		{
			name: "success: added token metadata to rollapp",
			fields: fields{
				ibcModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					rollapp: &rollapptypes.Rollapp{},
				},
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToExistingMemo("", validDenomMetadata),
				},
				ack: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{
				RegisteredDenoms: []string{validDenomMetadata.Base},
			},
		}, {
			name: "success: added token metadata to rollapp with user memo",
			fields: fields{
				ibcModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					rollapp: &rollapptypes.Rollapp{},
				},
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToExistingMemo("user memo", validDenomMetadata),
				},
				ack: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{
				RegisteredDenoms: []string{validDenomMetadata.Base},
			},
		}, {
			name: "return early: error ack",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				ibcModule:     mockIBCModule{},
				ack:           errAck(),
			},
			wantRollapp: nil,
		}, {
			name: "return early: no memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				ibcModule:     mockIBCModule{},
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
				},
				ack: okAck(),
			},
			wantRollapp: nil,
		}, {
			name: "return early: no packet metadata in memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				ibcModule:     mockIBCModule{},
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  "user memo",
				},
				ack: okAck(),
			},
			wantRollapp: nil,
		}, {
			name: "return early: no denom metadata in memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				ibcModule:     mockIBCModule{},
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  `{"transferinject":{}}`,
				},
				ack: okAck(),
			},
			wantRollapp: nil,
		}, {
			name: "error: extract rollapp from channel",
			fields: fields{
				ibcModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					err: errortypes.ErrInvalidRequest,
				},
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToExistingMemo("", validDenomMetadata),
				},
				ack: okAck(),
			},
			wantRollapp: nil,
			wantErr:     errortypes.ErrInvalidRequest,
		}, {
			name: "error: rollapp not found",
			fields: fields{
				ibcModule:     mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{},
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToExistingMemo("", validDenomMetadata),
				},
				ack: okAck(),
			},
			wantRollapp: nil,
			wantErr:     gerr.ErrNotFound,
		}, {
			name: "return early: rollapp already has token metadata",
			fields: fields{
				ibcModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					rollapp: &rollapptypes.Rollapp{
						RegisteredDenoms: []string{validDenomMetadata.Base},
					},
				},
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToExistingMemo("", validDenomMetadata),
				},
				ack: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{
				RegisteredDenoms: []string{validDenomMetadata.Base},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.rollappKeeper.transfer = tt.fields.packetData

			m := transferinject.NewIBCModule(tt.fields.ibcModule, tt.fields.rollappKeeper)

			packet := channeltypes.Packet{}

			if tt.fields.packetData != nil {
				packet.Data = types.ModuleCdc.MustMarshalJSON(tt.fields.packetData)
			}

			err := m.OnAcknowledgementPacket(sdk.Context{}, packet, tt.fields.ack, sdk.AccAddress{})

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.wantRollapp, tt.fields.rollappKeeper.rollapp)
		})
	}
}

func okAck() []byte {
	ack := channeltypes.NewResultAcknowledgement([]byte{})
	return types.ModuleCdc.MustMarshalJSON(&ack)
}

func errAck() []byte {
	ack := channeltypes.NewErrorAcknowledgement(fmt.Errorf("unsuccessful"))
	return types.ModuleCdc.MustMarshalJSON(&ack)
}

type mockIBCModule struct {
	porttypes.IBCModule
}

func (m mockIBCModule) OnAcknowledgementPacket(sdk.Context, channeltypes.Packet, []byte, sdk.AccAddress) error {
	return nil
}
