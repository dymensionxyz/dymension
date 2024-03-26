package cli_test

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/testutil/network"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/client/cli"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func networkWithLatestStateIndexObjects(t *testing.T, n int) (*network.Network, []types.StateInfoIndex) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {
		rollapp := types.Rollapp{
			Creator:   sample.AccAddress(),
			RollappId: strconv.Itoa(i),
		}
		state.RollappList = append(state.RollappList, rollapp)
	}

	blockDescriptors := types.BlockDescriptors{BD: make([]types.BlockDescriptor, 1)}
	blockDescriptors.BD[0] = types.BlockDescriptor{
		Height:                 1,
		StateRoot:              bytes.Repeat([]byte{byte(1)}, 32),
		IntermediateStatesRoot: bytes.Repeat([]byte{byte(1)}, 32),
	}

	for i := 0; i < n; i++ {
		stateInfo := types.StateInfo{
			StateInfoIndex: types.StateInfoIndex{
				RollappId: strconv.Itoa(i),
				Index:     uint64(i + 1),
			},
			Sequencer:   sample.AccAddress(),
			NumBlocks:   1,
			StartHeight: 1,
			BDs:         blockDescriptors,
		}
		state.StateInfoList = append(state.StateInfoList, stateInfo)
	}

	for i := 0; i < n; i++ {
		latestStateIndex := types.StateInfoIndex{
			RollappId: strconv.Itoa(i),
			Index:     1,
		}
		state.LatestStateInfoIndexList = append(state.LatestStateInfoIndexList, latestStateIndex)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.LatestStateInfoIndexList
}

func TestShowLatestStateInfoIndex(t *testing.T) {
	net, objs := networkWithLatestStateIndexObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc        string
		idRollappId string

		args []string
		err  error
		obj  types.StateInfoIndex
	}{
		{
			desc:        "found",
			idRollappId: objs[0].RollappId,

			args: common,
			obj:  objs[0],
		},
		{
			desc:        "not found",
			idRollappId: strconv.Itoa(100000),

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idRollappId,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowLatestStateIndex(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetLatestStateIndexResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.StateIndex)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.StateIndex),
				)
			}
		})
	}
}
