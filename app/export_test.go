package app_test

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	agenttypes "github.com/dymensionxyz/dymension/v3/x/agent/types"
)

// A zero-height export rebases agent heartbeats onto the restarted chain's
// height coordinates, so liveness ages survive the height reset.
func TestZeroHeightExportRebasesAgentHeartbeats(t *testing.T) {
	a := apptesting.Setup(t)
	ctx := a.NewContext(false)
	require.NoError(t, a.AgentKeeper.SetAgent(ctx, agenttypes.Agent{Id: "agent1", Active: true, LastActionHeight: 1}))
	_, err := a.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = a.Commit()
	require.NoError(t, err)
	require.Equal(t, int64(1), a.LastBlockHeight())

	exported, err := a.ExportAppStateAndValidators(true, nil, []string{agenttypes.ModuleName})
	require.NoError(t, err)

	var state map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(exported.AppState, &state))
	var g agenttypes.GenesisState
	require.NoError(t, a.AppCodec().UnmarshalJSON(state[agenttypes.ModuleName], &g))
	require.Len(t, g.Agents, 1)
	require.Equal(t, int64(0), g.Agents[0].LastActionHeight)
}
