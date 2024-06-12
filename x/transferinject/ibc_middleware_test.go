package transferinject_test

import (
	"fmt"
	"testing"

	"github.com/dymensionxyz/dymension/v3/utils/gerr"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/stretchr/testify/require"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/transferinject"
	"github.com/dymensionxyz/dymension/v3/x/transferinject/types"
)

func TestICS4Wrapper_SendPacket(t *testing.T) {
	type fields struct {
		rollappKeeper types.RollappKeeper
		bankKeeper    types.BankKeeper
		dstPort       string
		dstChannel    string
		data          *transfertypes.FungibleTokenPacketData
	}
	tests := []struct {
		name         string
		fields       fields
		wantSentData []byte
		wantErr      error
	}{
		{
			name: "success: added denom metadata to memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{
					rollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{
					returnMetadata: validDenomMetadata,
				},
				dstPort:    "port",
				dstChannel: "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
				},
			},
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  addDenomMetadataToExistingMemo("", validDenomMetadata),
			}),
		}, {
			name: "success: added denom metadata to non-empty user memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{
					rollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{
					returnMetadata: validDenomMetadata,
				},
				dstPort:    "port",
				dstChannel: "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  "thanks for the sweater, grandma!",
				},
			},
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  addDenomMetadataToExistingMemo("thanks for the sweater, grandma!", validDenomMetadata),
			}),
		}, {
			name: "error: denom metadata already in memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				bankKeeper:    mockBankKeeper{},
				dstPort:       "port",
				dstChannel:    "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  `{"transferinject":{}}`,
				},
			},
			wantSentData: []byte(""),
			wantErr:      types.ErrMemoTransferInjectAlreadyExists,
		}, {
			name: "error: extract rollapp from channel",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{
					err: errortypes.ErrInvalidRequest,
				},
				bankKeeper: mockBankKeeper{},
				dstPort:    "port",
				dstChannel: "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
				},
			},
			wantSentData: []byte(""),
			wantErr:      errortypes.ErrInvalidRequest,
		}, {
			name: "send unaltered: rollapp not found",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				bankKeeper:    mockBankKeeper{},
				dstPort:       "port",
				dstChannel:    "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  "user memo",
				},
			},
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  "user memo",
			}),
		}, {
			name: "send unaltered: receiver chain is source",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{
					rollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{
					returnMetadata: validDenomMetadata,
				},
				dstPort:    "transfer",
				dstChannel: "channel-56",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "transfer/channel-56/alex",
				},
			},
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "transfer/channel-56/alex",
			}),
		}, {
			name: "send unaltered: denom metadata already in rollapp",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{
					rollapp: &rollapptypes.Rollapp{
						RegisteredDenoms: []string{"adym"},
					},
				},
				bankKeeper: mockBankKeeper{},
				dstPort:    "port",
				dstChannel: "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
				},
			},
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
			}),
		}, {
			name: "error: get denom metadata",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{
					rollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{},
				dstPort:    "port",
				dstChannel: "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
				},
			},
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ics4 := mockICS4Wrapper{}

			m := transferinject.NewICS4Wrapper(&ics4, tt.fields.rollappKeeper, tt.fields.bankKeeper)

			data := types.ModuleCdc.MustMarshalJSON(tt.fields.data)

			_, err := m.SendPacket(
				sdk.Context{},
				&capabilitytypes.Capability{},
				tt.fields.dstPort,
				tt.fields.dstChannel,
				clienttypes.Height{},
				0,
				data,
			)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			require.Equal(t, string(tt.wantSentData), string(ics4.sentData))
		})
	}
}

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

var validDenomMetadata = banktypes.Metadata{
	Description: "Denom of the Hub",
	Base:        "adym",
	Display:     "DYM",
	Name:        "DYM",
	Symbol:      "adym",
	DenomUnits: []*banktypes.DenomUnit{
		{
			Denom:    "adym",
			Exponent: 0,
		}, {
			Denom:    "DYM",
			Exponent: 18,
		},
	},
}

func addDenomMetadataToExistingMemo(memo string, metadata banktypes.Metadata) string {
	memo, _ = types.AddDenomMetadataToMemo(memo, metadata)
	return memo
}

func okAck() []byte {
	ack := channeltypes.NewResultAcknowledgement([]byte{})
	return types.ModuleCdc.MustMarshalJSON(&ack)
}

func errAck() []byte {
	ack := channeltypes.NewErrorAcknowledgement(fmt.Errorf("unsuccessful"))
	return types.ModuleCdc.MustMarshalJSON(&ack)
}

type mockICS4Wrapper struct {
	porttypes.ICS4Wrapper
	sentData []byte
}

func (m *mockICS4Wrapper) SendPacket(
	_ sdk.Context,
	_ *capabilitytypes.Capability,
	_ string, _ string,
	_ clienttypes.Height,
	_ uint64,
	data []byte,
) (sequence uint64, err error) {
	m.sentData = data
	return 0, nil
}

type mockIBCModule struct {
	porttypes.IBCModule
}

func (m mockIBCModule) OnAcknowledgementPacket(sdk.Context, channeltypes.Packet, []byte, sdk.AccAddress) error {
	return nil
}

type mockBankKeeper struct {
	returnMetadata banktypes.Metadata
}

func (m mockBankKeeper) GetDenomMetaData(sdk.Context, string) (banktypes.Metadata, bool) {
	return m.returnMetadata, m.returnMetadata.Base != ""
}

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
