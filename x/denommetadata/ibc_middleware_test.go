package denommetadata_test

import (
	"encoding/json"
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
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestIBCModule_OnRecvPacket(t *testing.T) {
	tests := []struct {
		name          string
		keeper        *mockDenomMetadataKeeper
		rollappKeeper *mockRollappKeeper

		memoData         *memoData
		wantAck          exported.Acknowledgement
		wantSentMemoData *memoData
		wantCreated      bool
	}{
		{
			name:   "valid packet data with packet metadata",
			keeper: &mockDenomMetadataKeeper{},
			rollappKeeper: &mockRollappKeeper{
				returnRollapp: &rollapptypes.Rollapp{},
			},
			memoData:         validMemoData,
			wantAck:          emptyResult,
			wantSentMemoData: nil,
			wantCreated:      true,
		}, {
			name:   "valid packet data with packet metadata and user memo",
			keeper: &mockDenomMetadataKeeper{},
			rollappKeeper: &mockRollappKeeper{
				returnRollapp: &rollapptypes.Rollapp{},
			},
			memoData:         validMemoDataWithUserMemo,
			wantAck:          emptyResult,
			wantSentMemoData: validUserMemo,
			wantCreated:      true,
		}, {
			name:   "no memo",
			keeper: &mockDenomMetadataKeeper{},
			rollappKeeper: &mockRollappKeeper{
				returnRollapp: &rollapptypes.Rollapp{},
			},
			memoData:         nil,
			wantAck:          emptyResult,
			wantSentMemoData: nil,
			wantCreated:      false,
		}, {
			name:   "custom memo",
			keeper: &mockDenomMetadataKeeper{},
			rollappKeeper: &mockRollappKeeper{
				returnRollapp: &rollapptypes.Rollapp{},
			},
			memoData:         validUserMemo,
			wantAck:          emptyResult,
			wantSentMemoData: validUserMemo,
			wantCreated:      false,
		}, {
			name:   "memo has empty denom metadata",
			keeper: &mockDenomMetadataKeeper{},
			rollappKeeper: &mockRollappKeeper{
				returnRollapp: &rollapptypes.Rollapp{},
			},
			memoData:         invalidMemoDataNoDenomMetadata,
			wantAck:          emptyResult,
			wantSentMemoData: nil,
			wantCreated:      false,
		}, {
			name:   "denom metadata already exists in keeper",
			keeper: &mockDenomMetadataKeeper{hasDenomMetaData: true},
			rollappKeeper: &mockRollappKeeper{
				returnRollapp: &rollapptypes.Rollapp{},
			},
			memoData:         validMemoData,
			wantAck:          emptyResult,
			wantSentMemoData: nil,
			wantCreated:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &mockIBCModule{}
			im := denommetadata.NewIBCModule(app, tt.keeper, tt.rollappKeeper)
			var memo string
			if tt.memoData != nil {
				memo = mustMarshalJSON(tt.memoData)
			}
			packetData := packetDataWithMemo(memo)
			tt.rollappKeeper.packetData = packetData
			packetDataBytes := types.ModuleCdc.MustMarshalJSON(&packetData)
			packet := channeltypes.Packet{Data: packetDataBytes, SourcePort: "transfer", SourceChannel: "channel-0"}
			got := im.OnRecvPacket(sdk.NewContext(nil, tmproto.Header{}, false, nil), packet, sdk.AccAddress{})
			require.Equal(t, tt.wantAck, got)
			if !tt.wantAck.Success() {
				return
			}
			var wantMemo string
			if tt.wantSentMemoData != nil {
				wantMemo = mustMarshalJSON(tt.wantSentMemoData)
			}
			wantPacketData := packetDataWithMemo(wantMemo)
			wantPacketDataBytes := types.ModuleCdc.MustMarshalJSON(&wantPacketData)
			require.Equal(t, string(wantPacketDataBytes), string(app.sentData))
			require.Equal(t, tt.wantCreated, tt.keeper.created)
		})
	}
}

func TestICS4Wrapper_SendPacket(t *testing.T) {
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
					Memo:  `{"denom_metadata":{}}`,
				},
			},
			wantSentData: []byte(""),
			wantErr:      types.ErrMemoDenomMetadataAlreadyExists,
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
			m := denommetadata.NewICS4Wrapper(tt.fields.ICS4Wrapper, tt.fields.rollappKeeper, tt.fields.bankKeeper)

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

func TestIBCModule_OnAcknowledgementPacket(t *testing.T) {
	type fields struct {
		IBCModule     porttypes.IBCModule
		rollappKeeper *mockRollappKeeper
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
				IBCModule: &mockIBCModule{},
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
				IBCModule: &mockIBCModule{},
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
				IBCModule:     &mockIBCModule{},
			},
			args: args{
				acknowledgement: badAck(),
			},
			wantRollapp: nil,
		}, {
			name: "return early: no memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{
					returnRollapp: &rollapptypes.Rollapp{},
				},
				IBCModule: &mockIBCModule{},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
				},
				acknowledgement: okAck(),
			},
			wantRollapp: &rollapptypes.Rollapp{},
		}, {
			name: "return early: no packet metadata in memo",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				IBCModule:     &mockIBCModule{},
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
				IBCModule:     &mockIBCModule{},
			},
			args: args{
				packetData: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  `{"denom_metadata":{}}`,
				},
				acknowledgement: okAck(),
			},
			wantRollapp: nil,
		}, {
			name: "error: extract rollapp from channel",
			fields: fields{
				IBCModule: &mockIBCModule{},
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
				IBCModule:     &mockIBCModule{},
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
			wantErr:     gerrc.ErrNotFound,
		}, {
			name: "return early: rollapp already has token metadata",
			fields: fields{
				IBCModule: &mockIBCModule{},
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
			m := denommetadata.NewIBCModule(tt.fields.IBCModule, nil, tt.fields.rollappKeeper)

			packet := channeltypes.Packet{}

			if tt.args.packetData != nil {
				tt.fields.rollappKeeper.packetData = *tt.args.packetData
				packet.Data = types.ModuleCdc.MustMarshalJSON(tt.args.packetData)
			}

			err := m.OnAcknowledgementPacket(sdk.Context{}, packet, tt.args.acknowledgement, sdk.AccAddress{})

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.wantRollapp, tt.fields.rollappKeeper.returnRollapp)
		})
	}
}

var (
	emptyResult   = channeltypes.Acknowledgement{}
	validUserMemo = &memoData{
		User: &validUserData,
	}
	validMemoDataWithUserMemo = &memoData{
		MemoData: validMemoData.MemoData,
		User:     &validUserData,
	}
	validUserData = userData{Data: "data"}
	validMemoData = &memoData{
		MemoData: types.MemoData{
			DenomMetadata: &validDenomMetadata,
		},
	}
	invalidMemoDataNoDenomMetadata = &memoData{
		MemoData: types.MemoData{},
	}
	validDenomMetadata = banktypes.Metadata{
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
)

type memoData struct {
	types.MemoData
	User *userData `json:"user,omitempty"`
}

type userData struct {
	Data string `json:"data"`
}

func packetDataWithMemo(memo string) transfertypes.FungibleTokenPacketData {
	return transfertypes.FungibleTokenPacketData{
		Denom:    "adym",
		Amount:   "100",
		Sender:   "sender",
		Receiver: "receiver",
		Memo:     memo,
	}
}

func addDenomMetadataToPacketData(memo string, metadata banktypes.Metadata) string {
	memo, _ = types.AddDenomMetadataToMemo(memo, metadata)
	return memo
}

func mustMarshalJSON(v any) string {
	bz, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

type mockIBCModule struct {
	porttypes.IBCModule
	sentData []byte
}

func okAck() []byte {
	ack := channeltypes.NewResultAcknowledgement([]byte{})
	return types.ModuleCdc.MustMarshalJSON(&ack)
}

func badAck() []byte {
	ack := channeltypes.NewErrorAcknowledgement(fmt.Errorf("unsuccessful"))
	return types.ModuleCdc.MustMarshalJSON(&ack)
}

func (m *mockIBCModule) OnRecvPacket(_ sdk.Context, p channeltypes.Packet, _ sdk.AccAddress) exported.Acknowledgement {
	m.sentData = p.Data
	return emptyResult
}

func (m *mockIBCModule) OnAcknowledgementPacket(_ sdk.Context, _ channeltypes.Packet, ack []byte, _ sdk.AccAddress) error {
	return nil
}

type mockDenomMetadataKeeper struct {
	hasDenomMetaData, created bool
}

func (m *mockDenomMetadataKeeper) CreateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error {
	m.created = true
	return nil
}

type mockTransferKeeper struct {
	hasDT   bool
	created bool
}

func (m *mockTransferKeeper) HasDenomTrace(sdk.Context, tmbytes.HexBytes) bool {
	return m.hasDT
}

func (m *mockTransferKeeper) SetDenomTrace(sdk.Context, transfertypes.DenomTrace) {
	m.created = true
}

func (m *mockTransferKeeper) OnRecvPacket(sdk.Context, channeltypes.Packet, sdk.AccAddress) exported.Acknowledgement {
	return emptyResult
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

type mockRollappKeeper struct {
	returnRollapp *rollapptypes.Rollapp
	packetData    transfertypes.FungibleTokenPacketData
	err           error
}

func (m *mockRollappKeeper) SetRollapp(_ sdk.Context, rollapp rollapptypes.Rollapp) {
	m.returnRollapp = &rollapp
}

func (m *mockRollappKeeper) GetValidTransfer(sdk.Context, []byte, string, string) (data rollapptypes.TransferData, err error) {
	return rollapptypes.TransferData{
		Rollapp:                 m.returnRollapp,
		FungibleTokenPacketData: m.packetData,
	}, m.err
}

type mockBankKeeper struct {
	returnMetadata banktypes.Metadata
}

func (m mockBankKeeper) SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata) {
}

func (m mockBankKeeper) GetDenomMetaData(sdk.Context, string) (banktypes.Metadata, bool) {
	return m.returnMetadata, m.returnMetadata.Base != ""
}
