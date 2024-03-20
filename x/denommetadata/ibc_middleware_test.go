package denommetadata_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v6/packetforward/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	tmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata"
	rollappmoduletypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

func TestIBCMiddleware_OnRecvPacket(t *testing.T) {
	type setup struct {
		transferKeeperMock mockTransferKeeper
		channelKeeperMock  mockChannelKeeper
		rollappKeeperMock  *mockRollappKeeper
		logger             *collectingLogger
	}
	type args struct {
		packet  channeltypes.Packet
		relayer sdktypes.AccAddress
	}
	tests := []struct {
		name             string
		setup            setup
		args             args
		expDenomMetadata string
		expLogMsg        string
	}{
		{
			name: "test OnRecvPacket",
			args: args{
				packet:  transferPacket(t, senderAddr, hostAddr, testDenom, metadata),
				relayer: sample.Acc(),
			},
			setup: setup{
				transferKeeperMock: mockTransferKeeper{},
				channelKeeperMock: mockChannelKeeper{
					clientState: &tmtypes.ClientState{
						ChainId: testChainID,
					},
				},
				rollappKeeperMock: &mockRollappKeeper{
					rollapp: &rollappmoduletypes.Rollapp{
						RollappId: testChainID,
					},
					params: rollappmoduletypes.Params{
						RollappsEnabled: true,
					},
				},
				logger: &collectingLogger{},
			},
			expDenomMetadata: "registered",
		}, {
			name: "test OnRecvPacket with no rollapp",
			args: args{
				packet:  transferPacket(t, senderAddr, hostAddr, testDenom, metadata),
				relayer: sample.Acc(),
			},
			setup: setup{
				channelKeeperMock: mockChannelKeeper{
					clientState: &tmtypes.ClientState{
						ChainId: testChainID,
					},
				},
				rollappKeeperMock: &mockRollappKeeper{
					params: rollappmoduletypes.Params{
						RollappsEnabled: true,
					},
				},
				logger: &collectingLogger{},
			},
			expDenomMetadata: "",
			expLogMsg:        "Skipping denommetadata middleware. Chain is not a rollapp. ",
		}, {
			name: "test OnRecvPacket with rollapps disabled",
			args: args{
				packet:  transferPacket(t, senderAddr, hostAddr, testDenom, metadata),
				relayer: sample.Acc(),
			},
			setup: setup{
				rollappKeeperMock: &mockRollappKeeper{
					params: rollappmoduletypes.Params{
						RollappsEnabled: false,
					},
				},
				logger: &collectingLogger{},
			},
			expDenomMetadata: "",
			expLogMsg:        "Skipping IBC transfer OnRecvPacket for rollapps not enabled",
		}, {
			name: "test OnRecvPacket with no rollapp client state",
			args: args{
				packet:  transferPacket(t, senderAddr, hostAddr, testDenom, metadata),
				relayer: sample.Acc(),
			},
			setup: setup{
				channelKeeperMock: mockChannelKeeper{
					err: clienttypes.ErrClientNotFound,
				},
				rollappKeeperMock: &mockRollappKeeper{
					params: rollappmoduletypes.Params{
						RollappsEnabled: true,
					},
				},
				logger: &collectingLogger{},
			},
			expDenomMetadata: "",
			expLogMsg:        "failed to extract clientID from channel",
		}, {
			name: "test OnRecvPacket with receiver chain being the source chain",
			args: args{
				packet:  transferPacket(t, hostAddr, senderAddr, prefixedDenom, metadata),
				relayer: sample.Acc(),
			},
			setup: setup{
				rollappKeeperMock: &mockRollappKeeper{
					params: rollappmoduletypes.Params{
						RollappsEnabled: true,
					},
				},
				logger: &collectingLogger{},
			},
			expDenomMetadata: "",
			expLogMsg:        "Skipping IBC transfer OnRecvPacket for receiver chain being the source chain",
		}, {
			name: "test OnRecvPacket with wrong client state",
			args: args{
				packet:  transferPacket(t, senderAddr, hostAddr, testDenom, metadata),
				relayer: sample.Acc(),
			},
			setup: setup{
				channelKeeperMock: mockChannelKeeper{
					clientState: nil,
				},
				rollappKeeperMock: &mockRollappKeeper{
					params: rollappmoduletypes.Params{
						RollappsEnabled: true,
					},
				},
				logger: &collectingLogger{},
			},
			expDenomMetadata: "",
			expLogMsg:        "failed to extract chainID from clientState",
		}, {
			name: "test OnRecvPacket when already has denom trace",
			args: args{
				packet:  transferPacket(t, senderAddr, hostAddr, testDenom, metadata),
				relayer: sample.Acc(),
			},
			setup: setup{
				transferKeeperMock: mockTransferKeeper{
					hasDenomTrace: true,
				},
				channelKeeperMock: mockChannelKeeper{
					clientState: &tmtypes.ClientState{
						ChainId: testChainID,
					},
				},
				rollappKeeperMock: &mockRollappKeeper{
					rollapp: &rollappmoduletypes.Rollapp{
						RollappId: testChainID,
					},
					params: rollappmoduletypes.Params{
						RollappsEnabled: true,
					},
				},
				logger: &collectingLogger{},
			},
			expDenomMetadata: "",
			expLogMsg:        "Skipping denommetadata middleware. Denom trace already exists",
		}, {
			name: "test OnRecvPacket with wrong packet data",
			args: args{
				packet:  packetWrongTransferData,
				relayer: sample.Acc(),
			},
			setup: setup{
				rollappKeeperMock: &mockRollappKeeper{
					params: rollappmoduletypes.Params{
						RollappsEnabled: true,
					},
				},
				logger: &collectingLogger{},
			},
			expDenomMetadata: "",
			expLogMsg:        "failed to unmarshal fungible token packet data",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdktypes.NewContext(store.NewCommitMultiStore(tmdb.NewMemDB()), tmproto.Header{}, false, tt.setup.logger)

			middleware := denommetadata.NewIBCMiddleware(
				mockIBCModule{},
				tt.setup.channelKeeperMock,
				nil,
				tt.setup.transferKeeperMock,
				tt.setup.rollappKeeperMock,
			)

			_ = middleware.OnRecvPacket(ctx, tt.args.packet, tt.args.relayer)
			denomMetadata := tt.setup.rollappKeeperMock.getDenomMetadata()
			require.Equal(t, tt.expDenomMetadata, denomMetadata)
			logMsg := tt.setup.logger.getMsg()
			require.Equal(t, tt.expLogMsg, logMsg)
		})
	}
}

var (
	testDenom  = "urax"
	testAmount = "100"

	testSourcePort         = "transfer"
	testSourceChannel      = "channel-10"
	testDestinationPort    = "transfer"
	testDestinationChannel = "channel-11"
	testChainID            = "test"

	senderAddr = sample.AccAddress()
	hostAddr   = sample.AccAddress()
	destAddr   = sample.AccAddress()

	port                    = "transfer"
	channel                 = "channel-0"
	prefixedDenom           = fmt.Sprintf("%s/%s/%s", testSourcePort, testSourceChannel, testDenom)
	packetWrongTransferData = channeltypes.Packet{
		Data: []byte("wrong data"),
	}

	metadata = &ibctypes.PacketMetadata{Forward: &ibctypes.ForwardMetadata{
		Receiver: destAddr,
		Port:     port,
		Channel:  channel,
	}}
)

func transferPacket(t *testing.T, sender, receiver, denom string, metadata any) channeltypes.Packet {
	t.Helper()
	transferPacket := transfertypes.FungibleTokenPacketData{
		Denom:    denom,
		Amount:   testAmount,
		Sender:   sender,
		Receiver: receiver,
	}

	memo, err := json.Marshal(metadata)
	require.NoError(t, err)
	transferPacket.Memo = string(memo)

	transferData, err := transfertypes.ModuleCdc.MarshalJSON(&transferPacket)
	require.NoError(t, err)

	return channeltypes.Packet{
		SourcePort:         testSourcePort,
		SourceChannel:      testSourceChannel,
		DestinationPort:    testDestinationPort,
		DestinationChannel: testDestinationChannel,
		Data:               transferData,
	}
}

type mockIBCModule struct {
	porttypes.IBCModule
}

func (mockIBCModule) OnRecvPacket(_ sdktypes.Context, _ channeltypes.Packet, _ sdktypes.AccAddress) exported.Acknowledgement {
	return nil
}

type mockTransferKeeper struct {
	hasDenomTrace bool
}

func (m mockTransferKeeper) HasDenomTrace(sdktypes.Context, tmbytes.HexBytes) bool {
	return m.hasDenomTrace
}

type mockChannelKeeper struct {
	clientState exported.ClientState
	err         error
}

func (m mockChannelKeeper) GetChannelClientState(sdktypes.Context, string, string) (string, exported.ClientState, error) {
	return "", m.clientState, m.err
}

type mockRollappKeeper struct {
	params  rollappmoduletypes.Params
	rollapp *rollappmoduletypes.Rollapp
	sync.Mutex
	denomMetadata string
}

func (m *mockRollappKeeper) GetParams(sdktypes.Context) rollappmoduletypes.Params {
	return m.params
}

func (m *mockRollappKeeper) GetRollapp(sdktypes.Context, string) (rollapp rollappmoduletypes.Rollapp, found bool) {
	if m.rollapp == nil {
		return rollappmoduletypes.Rollapp{}, false
	}
	return *m.rollapp, m.rollapp != nil
}

func (m *mockRollappKeeper) RegisterDenomMetadata(sdktypes.Context, rollappmoduletypes.Rollapp, string, string) {
	m.Lock()
	defer m.Unlock()
	m.denomMetadata = "registered"
}

func (m *mockRollappKeeper) getDenomMetadata() string {
	m.Lock()
	defer m.Unlock()
	return m.denomMetadata
}

type collectingLogger struct {
	msg string
	sync.Mutex
}

func (c *collectingLogger) Debug(msg string, _ ...interface{}) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.msg = msg
}

func (c *collectingLogger) Info(msg string, _ ...interface{}) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.msg = msg
}

func (c *collectingLogger) Error(msg string, _ ...interface{}) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.msg = msg
}

func (c *collectingLogger) getMsg() string {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.msg
}

func (c *collectingLogger) With(...interface{}) log.Logger {
	return c
}

var _ log.Logger = &collectingLogger{}
