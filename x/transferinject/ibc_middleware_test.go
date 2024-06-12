package transferinject_test

import (
	"fmt"
	"testing"

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

func TestIBCMiddleware_SendPacket(t *testing.T) {
	type fields struct {
		ICS4Wrapper   porttypes.ICS4Wrapper
		rollappKeeper types.RollappKeeper
		bankKeeper    types.BankKeeper
	}
	type args struct {
		destinationPort    string
		destinationChannel string
		data               *transfertypes.FungibleTokenPacketData
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantSentData []byte
		wantErr      error
	}{
		{
			name: "success: added denom metadata to memo",
			fields: fields{
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{
					returnMetadata: validDenomMetadata,
				},
			},
			args: args{
				destinationPort:    "port",
				destinationChannel: "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
				},
			},
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  addDenomMetadataToPacketData("", validDenomMetadata),
			}),
		}, {
			name: "success: added denom metadata to non-empty user memo",
			fields: fields{
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{
					returnMetadata: validDenomMetadata,
				},
			},
			args: args{
				destinationPort:    "port",
				destinationChannel: "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  "thanks for the sweater, grandma!",
				},
			},
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  addDenomMetadataToPacketData("thanks for the sweater, grandma!", validDenomMetadata),
			}),
		}, {
			name: "error: denom metadata already in memo",
			fields: fields{
				ICS4Wrapper: &mockICS4Wrapper{},
			},
			args: args{
				destinationPort:    "port",
				destinationChannel: "channel",
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
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					err: errortypes.ErrInvalidRequest,
				},
			},
			args: args{
				destinationPort:    "port",
				destinationChannel: "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
				},
			},
			wantSentData: []byte(""),
			wantErr:      errortypes.ErrInvalidRequest,
		}, {
			name: "send unaltered: rollapp not found",
			fields: fields{
				ICS4Wrapper:   &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{},
			},
			args: args{
				destinationPort:    "port",
				destinationChannel: "channel",
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
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{
					returnMetadata: validDenomMetadata,
				},
			},
			args: args{
				destinationPort:    "transfer",
				destinationChannel: "channel-56",
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
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{
						RegisteredDenoms: []string{"adym"},
					},
				},
			},
			args: args{
				destinationPort:    "port",
				destinationChannel: "channel",
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
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{},
			},
			args: args{
				destinationPort:    "port",
				destinationChannel: "channel",
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
			m := transferinject.NewICS4Wrapper(tt.fields.ICS4Wrapper, tt.fields.rollappKeeper, tt.fields.bankKeeper)

			data := types.ModuleCdc.MustMarshalJSON(tt.args.data)

			_, err := m.SendPacket(
				sdk.Context{},
				&capabilitytypes.Capability{},
				tt.args.destinationPort,
				tt.args.destinationChannel,
				clienttypes.Height{},
				0,
				data,
			)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			require.Equal(t, string(tt.wantSentData), string(tt.fields.ICS4Wrapper.(*mockICS4Wrapper).sentData))
		})
	}
}

func TestIBCMiddleware_OnAcknowledgementPacket(t *testing.T) {
	type fields struct {
		IBCModule     porttypes.IBCModule
		rollappKeeper types.RollappKeeper
	}
	type args struct {
		packetData      *transfertypes.FungibleTokenPacketData
		acknowledgement []byte
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantRollapp *rollapptypes.Rollapp
		wantErr     error
	}{
		{
			name: "success: added token metadata to rollapp",
			fields: fields{
				IBCModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToPacketData("", validDenomMetadata),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{
				RegisteredDenoms: []string{validDenomMetadata.Base},
			},
		}, {
			name: "success: added token metadata to rollapp with user memo",
			fields: fields{
				IBCModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToPacketData("user memo", validDenomMetadata),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{
				RegisteredDenoms: []string{validDenomMetadata.Base},
			},
		}, {
			name: "return early: error acknowledgement",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				IBCModule:     mockIBCModule{},
			},
			args: args{
				acknowledgement: badAck(),
			},
			wantRollapp: nil,
		}, {
			name: "return early: no memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				IBCModule:     mockIBCModule{},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
		}, {
			name: "return early: no packet metadata in memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				IBCModule:     mockIBCModule{},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  "user memo",
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
		}, {
			name: "return early: no denom metadata in memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				IBCModule:     mockIBCModule{},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  `{"transferinject":{}}`,
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
		}, {
			name: "error: extract rollapp from channel",
			fields: fields{
				IBCModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					err: errortypes.ErrInvalidRequest,
				},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToPacketData("", validDenomMetadata),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
			wantErr:     errortypes.ErrInvalidRequest,
		}, {
			name: "error: rollapp not found",
			fields: fields{
				IBCModule:     mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToPacketData("", validDenomMetadata),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
			wantErr:     errortypes.ErrNotFound,
		}, {
			name: "return early: rollapp already has token metadata",
			fields: fields{
				IBCModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{
						RegisteredDenoms: []string{validDenomMetadata.Base},
					},
				},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  addDenomMetadataToPacketData("", validDenomMetadata),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{
				RegisteredDenoms: []string{validDenomMetadata.Base},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := transferinject.NewIBCModule(tt.fields.IBCModule, tt.fields.rollappKeeper)

			packet := channeltypes.Packet{}

			if tt.args.packetData != nil {
				packet.Data = types.ModuleCdc.MustMarshalJSON(tt.args.packetData)
			}

			err := m.OnAcknowledgementPacket(sdk.Context{}, packet, tt.args.acknowledgement, sdk.AccAddress{})

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.wantRollapp, tt.fields.rollappKeeper.(*mockRollappKeeper).returnRollapp)
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

func addDenomMetadataToPacketData(memo string, metadata banktypes.Metadata) string {
	memo, _ = types.AddDenomMetadataToMemo(memo, metadata)
	return memo
}

func okAck() []byte {
	ack := channeltypes.NewResultAcknowledgement([]byte{})
	return types.ModuleCdc.MustMarshalJSON(&ack)
}

func badAck() []byte {
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
	returnRollapp *rollapptypes.Rollapp
	err           error
	transfer      transfertypes.FungibleTokenPacketData
}

func (m *mockRollappKeeper) GetValidTransfer(ctx sdk.Context, packetData []byte, raPortOnHub, raChanOnHub string) (data rollapptypes.TransferData, err error) {
	panic("todo, implement get valid transfer on transferinject ibc middleware test")
}

func (m *mockRollappKeeper) SetRollapp(_ sdk.Context, rollapp rollapptypes.Rollapp) {
	m.returnRollapp = &rollapp
}
