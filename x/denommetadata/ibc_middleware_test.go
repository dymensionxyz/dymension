package denommetadata_test

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/stretchr/testify/require"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

func TestIBCMiddleware_OnRecvPacket(t *testing.T) {
	tests := []struct {
		name           string
		keeper         *mockDenomMetadataKeeper
		transferKeeper *mockTransferKeeper

		memoData         *memoData
		wantAck          exported.Acknowledgement
		wantSentMemoData *memoData
		wantCreated      bool
	}{
		{
			name:             "valid packet data with packet metadata",
			keeper:           &mockDenomMetadataKeeper{},
			transferKeeper:   &mockTransferKeeper{},
			memoData:         validMemoData,
			wantAck:          emptyResult,
			wantSentMemoData: nil,
			wantCreated:      true,
		}, {
			name:             "valid packet data with packet metadata and user memo",
			keeper:           &mockDenomMetadataKeeper{},
			transferKeeper:   &mockTransferKeeper{},
			memoData:         validMemoDataWithUserMemo,
			wantAck:          emptyResult,
			wantSentMemoData: validUserMemo,
			wantCreated:      true,
		}, {
			name:             "no memo",
			keeper:           &mockDenomMetadataKeeper{},
			transferKeeper:   &mockTransferKeeper{},
			memoData:         nil,
			wantAck:          emptyResult,
			wantSentMemoData: nil,
			wantCreated:      false,
		}, {
			name:             "custom memo",
			keeper:           &mockDenomMetadataKeeper{},
			transferKeeper:   &mockTransferKeeper{},
			memoData:         validUserMemo,
			wantAck:          emptyResult,
			wantSentMemoData: validUserMemo,
			wantCreated:      false,
		}, {
			name:             "memo has empty packet metadata",
			keeper:           &mockDenomMetadataKeeper{},
			transferKeeper:   &mockTransferKeeper{},
			memoData:         invalidMemoDataNoTransferInject,
			wantAck:          emptyResult,
			wantSentMemoData: invalidMemoDataNoTransferInject,
			wantCreated:      false,
		}, {
			name:             "memo has empty denom metadata",
			keeper:           &mockDenomMetadataKeeper{},
			transferKeeper:   &mockTransferKeeper{},
			memoData:         invalidMemoDataNoDenomMetadata,
			wantAck:          emptyResult,
			wantSentMemoData: nil,
			wantCreated:      false,
		}, {
			name:             "denom metadata already exists in keeper",
			keeper:           &mockDenomMetadataKeeper{hasDenomMetaData: true},
			transferKeeper:   &mockTransferKeeper{},
			memoData:         validMemoData,
			wantAck:          emptyResult,
			wantSentMemoData: nil,
			wantCreated:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &mockIBCModule{}
			im := denommetadata.NewIBCMiddleware(app, tt.transferKeeper, tt.keeper)
			var memo string
			if tt.memoData != nil {
				memo = mustMarshalJSON(tt.memoData)
			}
			packetData := packetDataWithMemo(memo)
			packet := channeltypes.Packet{Data: packetData, SourcePort: "transfer", SourceChannel: "channel-0"}
			got := im.OnRecvPacket(sdk.Context{}, packet, sdk.AccAddress{})
			require.Equal(t, tt.wantAck, got)
			if !tt.wantAck.Success() {
				return
			}
			var wantMemo string
			if tt.wantSentMemoData != nil {
				wantMemo = mustMarshalJSON(tt.wantSentMemoData)
			}
			require.Equal(t, string(packetDataWithMemo(wantMemo)), string(app.sentData))
			require.Equal(t, tt.wantCreated, tt.keeper.created)
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
			TransferInject: &types.TransferInject{
				DenomMetadata: &validDenomMetadata,
			},
		},
	}
	invalidMemoDataNoDenomMetadata = &memoData{
		MemoData: types.MemoData{
			TransferInject: &types.TransferInject{},
		},
	}
	invalidMemoDataNoTransferInject = &memoData{
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

func packetDataWithMemo(memo string) []byte {
	byt, _ := types.ModuleCdc.MarshalJSON(&transfertypes.FungibleTokenPacketData{
		Denom:    "adym",
		Amount:   "100",
		Sender:   "sender",
		Receiver: "receiver",
		Memo:     memo,
	})
	return byt
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

func (m *mockIBCModule) OnRecvPacket(_ sdk.Context, p channeltypes.Packet, _ sdk.AccAddress) exported.Acknowledgement {
	m.sentData = p.Data
	return emptyResult
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
