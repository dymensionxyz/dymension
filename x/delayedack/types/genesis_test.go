package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	chantypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	ctypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// this is needed to register the correct BECH32 prefix
const _ = appparams.BaseDenom

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		}, {
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.Params{
					EpochIdentifier: "hour",
					BridgingFee:     sdk.NewDecWithPrec(1, 1),
				},
				RollappPackets: []ctypes.RollappPacket{validRollappPacket},
			},
			valid: true,
		}, {
			desc: "invalid params",
			genState: &types.GenesisState{
				Params: types.Params{
					EpochIdentifier: "",
					BridgingFee:     sdk.Dec{},
				},
			},
			valid: false,
		}, {
			desc:     "invalid rollapp packet",
			genState: &types.GenesisState{RollappPackets: []ctypes.RollappPacket{{}}, Params: types.DefaultParams()},
			valid:    false,
		}, {
			desc: "duplicate rollapp packet",
			genState: &types.GenesisState{RollappPackets: []ctypes.RollappPacket{
				validRollappPacket,
				validRollappPacket,
			}, Params: types.DefaultParams()},
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

var validRollappPacket = ctypes.RollappPacket{
	RollappId: "1",
	Packet: &chantypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		Data:               []byte("data"),
		TimeoutHeight:      clienttypes.NewHeight(0, 1),
		TimeoutTimestamp:   1,
	},
	Acknowledgement:        nil,
	Status:                 ctypes.Status_PENDING,
	ProofHeight:            100,
	Relayer:                []byte("cosmos1"),
	Type:                   ctypes.RollappPacket_ON_RECV,
	Error:                  "error",
	OriginalTransferTarget: "dym1hpnekcl344ckklw07j7qcfs2x3j03zn6rppt2r",
}
