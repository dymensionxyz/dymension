package cli_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/testutil/network"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/client/cli"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func networkWithSequencerObjects(t *testing.T, n int) (*network.Network, []types.Sequencer) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {
		sequencer := types.Sequencer{
			SequencerAddress: strconv.Itoa(i),
			Status:           types.Bonded,
		}
		nullify.Fill(&sequencer)
		sequencer.Tokens = sdk.Coin{"", sdk.ZeroInt()}
		if i == 0 {
			sequencer.Status = types.Proposer
		}
		state.SequencerList = append(state.SequencerList, sequencer)
	}

	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.SequencerList
}

func TestShowSequencer(t *testing.T) {
	net, objs := networkWithSequencerObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc               string
		idSequencerAddress string

		args []string
		err  error
		obj  types.Sequencer
	}{
		{
			desc:               "found",
			idSequencerAddress: objs[0].SequencerAddress,

			args: common,
			obj:  objs[0],
		},
		{
			desc:               "not found",
			idSequencerAddress: strconv.Itoa(100000),

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idSequencerAddress,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowSequencer(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetSequencerResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.SequencerInfo)
				require.Equal(t,
					&tc.obj,
					&resp.SequencerInfo,
				)
			}
		})
	}
}

func TestListSequencer(t *testing.T) {
	net, objs := networkWithSequencerObjects(t, 5)

	var sequencerInfoList []types.Sequencer
	for _, obj := range objs {
		sequencerInfoList = append(sequencerInfoList, obj)
	}

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
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListSequencer(), args)
			require.NoError(t, err)
			var resp types.QueryAllSequencerResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.SequencerInfoList), step)
			require.Subset(t,
				sequencerInfoList,
				resp.SequencerInfoList,
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(objs); i += step {
			args := request(next, 0, uint64(step), false)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListSequencer(), args)
			require.NoError(t, err)
			var resp types.QueryAllSequencerResponse
			require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
			require.LessOrEqual(t, len(resp.SequencerInfoList), step)
			require.Subset(t,
				sequencerInfoList,
				resp.SequencerInfoList,
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		args := request(nil, 0, uint64(len(objs)), true)
		out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdListSequencer(), args)
		require.NoError(t, err)
		var resp types.QueryAllSequencerResponse
		require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
		require.NoError(t, err)
		require.Equal(t, len(objs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			sequencerInfoList,
			resp.SequencerInfoList,
		)
	})
}
