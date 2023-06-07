package cli_test

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/testutil/network"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/rollapp/client/cli"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

var RollappIds = []string{"1", "2"}

func networkWithStateInfoObjects(t *testing.T, n int) (*network.Network, []types.StateInfo, []types.StateInfoSummary) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 1; i <= n; i++ {
		stateInfo := types.StateInfo{
			StateInfoIndex: types.StateInfoIndex{
				RollappId: RollappIds[i%2],
				Index:     uint64(i)},
		}
		nullify.Fill(&stateInfo)
		state.StateInfoList = append(state.StateInfoList, stateInfo)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	var stateInfoSummaries []types.StateInfoSummary
	for _, state := range state.StateInfoList {
		stateInfoSummary := types.StateInfoSummary{
			StateInfoIndex: state.StateInfoIndex,
			Status:         state.Status,
			CreationHeight: state.CreationHeight,
		}
		stateInfoSummaries = append(stateInfoSummaries, stateInfoSummary)
	}
	return network.New(t, cfg), state.StateInfoList, stateInfoSummaries
}

func TestShowStateInfo(t *testing.T) {
	net, objs, _ := networkWithStateInfoObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	invalidFlagsError := status.Error(codes.InvalidArgument, fmt.Sprintf("only one flag can be use for %s, %s or %s", cli.FlagStateIndex, cli.FlagRollappHeight, cli.FlagFinalized))
	for _, tc := range []struct {
		desc         string
		idRollappId  string
		idStateIndex uint64

		args []string
		err  error
		obj  types.StateInfo
	}{
		{
			desc:         "found",
			idRollappId:  objs[0].StateInfoIndex.RollappId,
			idStateIndex: objs[0].StateInfoIndex.Index,

			args: common,
			obj:  objs[0],
		},
		{
			desc:         "invalid flags",
			idRollappId:  objs[0].StateInfoIndex.RollappId,
			idStateIndex: objs[0].StateInfoIndex.Index,

			args: append(common, fmt.Sprintf("--%s=10", cli.FlagRollappHeight), fmt.Sprintf("--%s=2", cli.FlagStateIndex)),
			err:  invalidFlagsError,
		},
		{
			desc:         "invalid flags",
			idRollappId:  objs[0].StateInfoIndex.RollappId,
			idStateIndex: objs[0].StateInfoIndex.Index,

			args: append(common, fmt.Sprintf("--%s=10", cli.FlagRollappHeight), fmt.Sprintf("--%s", cli.FlagFinalized)),
			err:  invalidFlagsError,
		},
		{
			desc:         "invalid flags",
			idRollappId:  objs[0].StateInfoIndex.RollappId,
			idStateIndex: objs[0].StateInfoIndex.Index,

			args: append(common, fmt.Sprintf("--%s=10", cli.FlagStateIndex), fmt.Sprintf("--%s", cli.FlagFinalized)),
			err:  invalidFlagsError,
		},
		{
			desc:         "not found",
			idRollappId:  strconv.Itoa(100000),
			idStateIndex: objs[0].StateInfoIndex.Index,

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
		{
			desc:         "not found",
			idRollappId:  objs[0].StateInfoIndex.RollappId,
			idStateIndex: 1000,

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idRollappId,
			}
			args = append(args, tc.args...)
			args = append(args, fmt.Sprintf("--%s=%d", cli.FlagStateIndex, tc.idStateIndex))
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowStateInfo(), args)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				var resp types.QueryGetStateInfoResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.StateInfo)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.StateInfo),
				)
			}
		})
	}
}

func TestListStateInfo(t *testing.T) {
	net, _, objs := networkWithStateInfoObjects(t, 5)

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
			requestArgs := request(nil, uint64(i), uint64(step), false)
			args := []string{RollappIds[0]}
			args = append(args, requestArgs...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListStateInfo(), args)
			require.NoError(t, err)
			var resp types.QueryAllStateInfoResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.StateInfo), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.StateInfo),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(objs); i += step {
			requestArgs := request(next, 0, uint64(step), false)
			args := []string{RollappIds[0]}
			args = append(args, requestArgs...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListStateInfo(), args)
			require.NoError(t, err)
			var resp types.QueryAllStateInfoResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.StateInfo), step)
			require.Subset(t,
				nullify.Fill(objs),
				nullify.Fill(resp.StateInfo),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		requestArgs := request(nil, 0, uint64(len(objs)), true)
		args := []string{RollappIds[0]}
		args = append(args, requestArgs...)
		out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListStateInfo(), args)
		require.NoError(t, err)
		var resp types.QueryAllStateInfoResponse
		require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
		require.NoError(t, err)

		var filteredObjects []types.StateInfoSummary
		for _, state := range objs {
			if state.StateInfoIndex.RollappId == RollappIds[0] {
				filteredObjects = append(filteredObjects, state)
			}
		}
		require.Equal(t, len(resp.StateInfo), len(filteredObjects))

		require.ElementsMatch(t,
			nullify.Fill(filteredObjects),
			nullify.Fill(resp.StateInfo),
		)
	})
}
