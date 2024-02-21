package cli_test

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
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

const (
	rollappId = "rollappID"
)

func networkWithSequencersByRollappObjects(t *testing.T, n int) (*network.Network, types.QueryGetSequencersByRollappResponse) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	var allSequencersByRollappResponse types.QueryGetSequencersByRollappResponse

	for i := 0; i < n; i++ {
		sequencer := types.Sequencer{
			SequencerAddress: strconv.Itoa(i),
			Status:           types.Bonded,
			RollappId:        rollappId,
		}
		nullify.Fill(&sequencer)
		if i == 0 {
			sequencer.Status = types.Proposer
		}
		state.SequencerList = append(state.SequencerList, sequencer)
		allSequencersByRollappResponse.Sequencers = append(allSequencersByRollappResponse.Sequencers, sequencer)
	}

	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), allSequencersByRollappResponse
}

func TestShowSequencersByRollapp(t *testing.T) {
	net, allSequencersByRollappResponse := networkWithSequencersByRollappObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc        string
		idRollappId string

		args []string
		err  error
		obj  types.QueryGetSequencersByRollappResponse
	}{
		{
			desc:        "found",
			idRollappId: rollappId,

			args: common,
			obj:  allSequencersByRollappResponse,
		},
		{
			desc:        "not found",
			idRollappId: strconv.Itoa(100000),

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idRollappId,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowSequencersByRollapp(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetSequencersByRollappResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.Sequencers)

				sortedList := resp.Sequencers
				sort.Slice(sortedList, func(i, j int) bool {
					return sortedList[i].SequencerAddress < sortedList[j].SequencerAddress
				})

				for i := range sortedList {
					require.True(t, equalSequencers(&tc.obj.Sequencers[i], &resp.Sequencers[i]))
				}
			}
		})
	}
}

// equalSequencer receives two sequencers and compares them. If there they not equal, fails the test
func equalSequencers(s1 *types.Sequencer, s2 *types.Sequencer) bool {
	if s1.SequencerAddress != s2.SequencerAddress {
		return false
	}

	s1Pubkey := s1.DymintPubKey
	s2Pubkey := s2.DymintPubKey
	if !s1Pubkey.Equal(s2Pubkey) {
		return false
	}
	if s1.RollappId != s2.RollappId {
		return false
	}

	if s1.Jailed != s2.Jailed {
		return false
	}
	if s1.Status != s2.Status {
		return false
	}

	if !s1.Tokens.IsEqual(s2.Tokens) {
		return false
	}

	if s1.UnbondingHeight != s2.UnbondingHeight {
		return false
	}
	if !s1.UnbondTime.Equal(s2.UnbondTime) {
		return false
	}
	return true
}
