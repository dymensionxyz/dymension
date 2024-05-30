package transferinject_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	tests := []struct {
		name         string
		fields       fields
		data         []byte
		wantSentData []byte
		wantErr      string
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
			data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
			}),
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  types.AddDenomMetadataToMemo("", validDenomMetadata),
			}),
			wantErr: "",
		}, {
			name: "success: added denom metadata to non-empty memo",
			fields: fields{
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{
					returnMetadata: validDenomMetadata,
				},
			},
			data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  "thanks for the sweater, grandma!",
			}),
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  types.AddDenomMetadataToMemo("thanks for the sweater, grandma!", validDenomMetadata),
			}),
			wantErr: "",
		}, {
			name: "error: unmarshal packet data",
			data: []byte("invalid json"),
			fields: fields{
				ICS4Wrapper: &mockICS4Wrapper{},
			},
			wantSentData: []byte(""),
			wantErr:      "unmarshal ICS-20 transfer packet data: invalid character 'i' looking for beginning of value: unknown request",
		}, {
			name: "error: denom metadata already in memo",
			fields: fields{
				ICS4Wrapper: &mockICS4Wrapper{},
			},
			data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  `{"packet_metadata":{}}`,
			}),
			wantSentData: []byte(""),
			wantErr:      "denom metadata already exists in memo: unauthorized",
		}, {
			name: "error: extract rollapp from channel",
			fields: fields{
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					err: fmt.Errorf("empty channel id"),
				},
			},
			data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
			}),
			wantSentData: []byte(""),
			wantErr:      "extract rollapp id from packet: empty channel id: not found",
		}, {
			name: "send unaltered: rollapp not found",
			fields: fields{
				ICS4Wrapper:   &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{},
			},
			data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  "user memo",
			}),
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
				Memo:  "user memo",
			}),
			wantErr: "",
		}, {
			name: "send unaltered: denom metadata already in rollapp",
			fields: fields{
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{
						TokenMetadata: []*rollapptypes.TokenMetadata{{Base: "adym"}},
					},
				},
			},
			data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
			}),
			wantSentData: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
			}),
			wantErr: "",
		}, {
			name: "error: get denom metadata",
			fields: fields{
				ICS4Wrapper: &mockICS4Wrapper{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
				bankKeeper: mockBankKeeper{},
			},
			data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
				Denom: "adym",
			}),
			wantSentData: []byte(""),
			wantErr:      "denom metadata not found: not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := transferinject.NewIBCSendMiddleware(tt.fields.ICS4Wrapper, tt.fields.rollappKeeper, tt.fields.bankKeeper)

			_, err := m.SendPacket(
				sdk.Context{},
				&capabilitytypes.Capability{},
				"transfer",
				"channel-0",
				clienttypes.Height{},
				0,
				tt.data,
			)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr)
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
		packet          channeltypes.Packet
		acknowledgement []byte
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantRollapp *rollapptypes.Rollapp
		wantErr     string
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
				packet: channeltypes.Packet{
					Data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
						Denom: "adym",
						Memo:  types.AddDenomMetadataToMemo("", validDenomMetadata),
					}),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{
				TokenMetadata: []*rollapptypes.TokenMetadata{tokenMetadata(validDenomMetadata)},
			},
			wantErr: "",
		}, {
			name: "success: added token metadata to rollapp with user memo",
			fields: fields{
				IBCModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
			},
			args: args{
				packet: channeltypes.Packet{
					Data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
						Denom: "adym",
						Memo:  types.AddDenomMetadataToMemo("user memo", validDenomMetadata),
					}),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{
				TokenMetadata: []*rollapptypes.TokenMetadata{tokenMetadata(validDenomMetadata)},
			},
			wantErr: "",
		}, {
			name: "error: unmarshal packet data",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
			},
			args: args{
				packet: channeltypes.Packet{
					Data: []byte("invalid json"),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
			wantErr:     "unmarshal ICS-20 transfer packet data: invalid character 'i' looking for beginning of value: unknown request",
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
			wantErr:     "",
		}, {
			name: "return early: no memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				IBCModule:     mockIBCModule{},
			},
			args: args{
				packet: channeltypes.Packet{
					Data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
						Denom: "adym",
					}),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
			wantErr:     "",
		}, {
			name: "return early: no packet metadata in memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				IBCModule:     mockIBCModule{},
			},
			args: args{
				packet: channeltypes.Packet{
					Data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
						Denom: "adym",
						Memo:  "user memo",
					}),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
			wantErr:     "",
		}, {
			name: "return early: no denom metadata in memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				IBCModule:     mockIBCModule{},
			},
			args: args{
				packet: channeltypes.Packet{
					Data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
						Denom: "adym",
						Memo:  `{"packet_metadata":{}}`,
					}),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
			wantErr:     "",
		}, {
			name: "error: extract rollapp from channel",
			fields: fields{
				IBCModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					err: fmt.Errorf("empty channel id"),
				},
			},
			args: args{
				packet: channeltypes.Packet{
					Data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
						Denom: "adym",
						Memo:  types.AddDenomMetadataToMemo("", validDenomMetadata),
					}),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
			wantErr:     "extract rollapp id from packet: empty channel id: not found",
		}, {
			name: "error: rollapp not found",
			fields: fields{
				IBCModule:     mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{},
			},
			args: args{
				packet: channeltypes.Packet{
					Data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
						Denom: "adym",
						Memo:  types.AddDenomMetadataToMemo("", validDenomMetadata),
					}),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
			wantErr:     "rollapp not found: not found",
		}, {
			name: "return early: rollapp already has token metadata",
			fields: fields{
				IBCModule: mockIBCModule{},
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{
						TokenMetadata: []*rollapptypes.TokenMetadata{tokenMetadata(validDenomMetadata)},
					},
				},
			},
			args: args{
				packet: channeltypes.Packet{
					Data: types.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
						Denom: "adym",
						Memo:  types.AddDenomMetadataToMemo("", validDenomMetadata),
					}),
				},
				acknowledgement: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{
				TokenMetadata: []*rollapptypes.TokenMetadata{tokenMetadata(validDenomMetadata)},
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := transferinject.NewIBCAckMiddleware(tt.fields.IBCModule, tt.fields.rollappKeeper)

			err := m.OnAcknowledgementPacket(sdk.Context{}, tt.args.packet, tt.args.acknowledgement, sdk.AccAddress{})

			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr)
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

func tokenMetadata(metadata banktypes.Metadata) *rollapptypes.TokenMetadata {
	return transferinject.DenomToTokenMetadata(&metadata)
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
}

func (m *mockRollappKeeper) SetRollapp(_ sdk.Context, rollapp rollapptypes.Rollapp) {
	m.returnRollapp = &rollapp
}

func (m *mockRollappKeeper) ExtractRollappFromChannel(sdk.Context, string, string) (*rollapptypes.Rollapp, error) {
	return m.returnRollapp, m.err
}
