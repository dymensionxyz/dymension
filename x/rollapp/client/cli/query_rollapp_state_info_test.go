package cli_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/testutil/network"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/rollapp/client/cli"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func networkWithRollappStateInfoObjects(t *testing.T, n int) (*network.Network, []types.RollappStateInfo) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {
		rollappStateInfo := types.RollappStateInfo{
			RollappId:  strconv.Itoa(i),
			StateIndex: uint64(i),
		}
		nullify.Fill(&rollappStateInfo)
		state.RollappStateInfoList = append(state.RollappStateInfoList, rollappStateInfo)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.RollappStateInfoList
}

func TestShowRollappStateInfo(t *testing.T) {
	net, objs := networkWithRollappStateInfoObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc         string
		idRollappId  string
		idStateIndex uint64

		args []string
		err  error
		obj  types.RollappStateInfo
	}{
		{
			desc:         "found",
			idRollappId:  objs[0].RollappId,
			idStateIndex: objs[0].StateIndex,

			args: common,
			obj:  objs[0],
		},
		{
			desc:         "not found",
			idRollappId:  strconv.Itoa(100000),
			idStateIndex: objs[0].StateIndex,

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
		{
			desc:         "not found",
			idRollappId:  objs[0].RollappId,
			idStateIndex: 1000,

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idRollappId,
				fmt.Sprint(tc.idStateIndex),
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowRollappStateInfo(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetRollappStateInfoResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.RollappStateInfo)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.RollappStateInfo),
				)
			}
		})
	}
}

func TestListRollappStateInfo(t *testing.T) {
	net, objs := networkWithRollappStateInfoObjects(t, 5)

	ctx := net.Validators[0].ClientCtx
	request := func(next []byte, offset, limit uint64, total bool) []string {
		args := []string{
			fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		}
		if next == nil {
			args = append(args, fmt.Sprintf("--%s=%d", flags.FlagOffset, offset))
		} else {
			args = append(args, fmt.Sprintf("--%s=%s", flags.FlagPageKey, next))
		}
		args = append(args, fmt.Sprintf("--%s=%d", flags.FlagLimit, limit))
		if total {
			args = append(args, fmt.Sprintf("--%s", flags.FlagCountTotal))
		}
		return args
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(objs); i += step {
			args := request(nil, uint64(i), uint64(step), false)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListRollappStateInfo(), args)
			require.NoError(t, err)
			var resp types.QueryAllRollappStateInfoResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.RollappStateInfo), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.RollappStateInfo),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(objs); i += step {
			args := request(next, 0, uint64(step), false)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListRollappStateInfo(), args)
			require.NoError(t, err)
			var resp types.QueryAllRollappStateInfoResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.RollappStateInfo), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.RollappStateInfo),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		args := request(nil, 0, uint64(len(objs)), true)
		out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListRollappStateInfo(), args)
		require.NoError(t, err)
		var resp types.QueryAllRollappStateInfoResponse
		require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
		require.NoError(t, err)
		require.Equal(t, len(objs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(objs),
			nullify.Fill(resp.RollappStateInfo),
		)
	})
}
