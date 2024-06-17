package transferinject_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/stretchr/testify/require"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/transferinject"
	"github.com/dymensionxyz/dymension/v3/x/transferinject/types"
)

func TestICS4Wrapper_SendPacket(t *testing.T) {
	type fields struct {
		rollappKeeper *mockRollappKeeper
		bankKeeper    types.BankKeeper
		srcPort       string
		srcChannel    string
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
				srcPort:    "port",
				srcChannel: "channel",
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
				srcPort:    "port",
				srcChannel: "channel",
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
				srcPort:       "port",
				srcChannel:    "channel",
				data: &transfertypes.FungibleTokenPacketData{
					Denom: "adym",
					Memo:  `{"transferinject":{}}`,
				},
			},
			wantSentData: []byte(""),
			wantErr:      types.ErrMemoTransferInjectAlreadyExists,
		}, {
			name: "send unaltered: rollapp not found",
			fields: fields{
				rollappKeeper: &mockRollappKeeper{},
				bankKeeper:    mockBankKeeper{},
				srcPort:       "port",
				srcChannel:    "channel",
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
				srcPort:    "transfer",
				srcChannel: "channel-56",
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
				srcPort:    "port",
				srcChannel: "channel",
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
				srcPort:    "port",
				srcChannel: "channel",
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
			tt.fields.rollappKeeper.transfer = tt.fields.data

			m := transferinject.NewICS4Wrapper(&ics4, tt.fields.rollappKeeper, tt.fields.bankKeeper)

			data := types.ModuleCdc.MustMarshalJSON(tt.fields.data)

			_, err := m.SendPacket(
				sdk.Context{},
				&capabilitytypes.Capability{},
				tt.fields.srcPort,
				tt.fields.srcChannel,
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

type mockBankKeeper struct {
	returnMetadata banktypes.Metadata
}

func (m mockBankKeeper) GetDenomMetaData(sdk.Context, string) (banktypes.Metadata, bool) {
	return m.returnMetadata, m.returnMetadata.Base != ""
}
