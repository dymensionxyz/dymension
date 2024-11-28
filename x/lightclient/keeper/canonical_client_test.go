package keeper_test

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	keeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

func TestTrySetCanonicalClient(t *testing.T) {

	type input struct {
		clientID string
	}

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context, k keeper.Keeper) input
		expectErr error
	}{
		{
			name: "not tm client",
			malleate: func(ctx sdk.Context, k keeper.Keeper) input {
				clientID := "not-tm-client"
				k.SetClientState(ctx, clientID, &types.ClientState{})
				return input{clientID: clientID}
			},
			expectErr: types.ErrInvalidArgument,
		},
		{
			name: "rollapp not exists",
			malleate: func(ctx sdk.Context, k keeper.Keeper) input {
				clientID := "rollapp-not-exists"
				clientState := types.DefaultExpectedCanonicalClientParams()
				clientState.ChainId = "non-existent-rollapp"
				k.SetClientState(ctx, clientID, &clientState)
				return input{clientID: clientID}
			},
			expectErr: types.ErrNotFound,
		},
		{
			name: "canon client already exists",
			malleate: func(ctx sdk.Context, k keeper.Keeper) input {
				clientID := "canon-client-exists"
				clientState := types.DefaultExpectedCanonicalClientParams()
				clientState.ChainId = keepertest.DefaultRollapp
				k.SetClientState(ctx, clientID, &clientState)
				k.SetCanonicalClient(ctx, keepertest.DefaultRollapp, clientID)
				return input{clientID: clientID}
			},
			expectErr: types.ErrAlreadyExists,
		},
		{
			name: "no latest height",
			malleate: func(ctx sdk.Context, k keeper.Keeper) input {
				clientID := "no-latest-height"
				clientState := types.DefaultExpectedCanonicalClientParams()
				clientState.ChainId = keepertest.DefaultRollapp
				k.SetClientState(ctx, clientID, &clientState)
				k.rollappKeeper.SetRollapp(ctx, keepertest.DefaultRollapp, types.Rollapp{})
				return input{clientID: clientID}
			},
			expectErr: types.ErrNotFound,
		},
		{
			name: "client not valid because of params",
			malleate: func(ctx sdk.Context, k keeper.Keeper) input {
				clientID := "invalid-params"
				clientState := types.DefaultExpectedCanonicalClientParams()
				clientState.ChainId = keepertest.DefaultRollapp
				clientState.TrustLevel = ibctm.NewFractionFromTm(math.Fraction{Numerator: 1, Denominator: 2}) // Invalid trust level
				k.SetClientState(ctx, clientID, &clientState)
				k.rollappKeeper.SetRollapp(ctx, keepertest.DefaultRollapp, types.Rollapp{})
				k.SetLatestHeight(ctx, keepertest.DefaultRollapp, 1)
				return input{clientID: clientID}
			},
			expectErr: types.ErrInvalidArgument,
		},
		{
			name: "client errNoMatchFound",
			malleate: func(ctx sdk.Context, k keeper.Keeper) input {
				clientID := "no-match-found"
				clientState := types.DefaultExpectedCanonicalClientParams()
				clientState.ChainId = keepertest.DefaultRollapp
				k.SetClientState(ctx, clientID, &clientState)
				k.rollappKeeper.SetRollapp(ctx, keepertest.DefaultRollapp, types.Rollapp{})
				k.SetLatestHeight(ctx, keepertest.DefaultRollapp, 1)
				return input{clientID: clientID}
			},
			expectErr: keeper.ErrNoMatchFound,
		},
		{
			name: "client right chain id and params but wrong consensus state",
			malleate: func(ctx sdk.Context, k keeper.Keeper) input {
				clientID := "wrong-consensus-state"
				clientState := types.DefaultExpectedCanonicalClientParams()
				clientState.ChainId = keepertest.DefaultRollapp
				k.SetClientState(ctx, clientID, &clientState)
				k.rollappKeeper.SetRollapp(ctx, keepertest.DefaultRollapp, types.Rollapp{})
				k.SetLatestHeight(ctx, keepertest.DefaultRollapp, 1)
				k.SetConsensusState(ctx, clientID, 1, &ibctm.ConsensusState{Timestamp: time.Now()})
				return input{clientID: clientID}
			},
			expectErr: keeper.ErrNoMatchFound,
		},
		{
			name: "success case",
			malleate: func(ctx sdk.Context, k keeper.Keeper) input {
				clientID := "success-case"
				clientState := types.DefaultExpectedCanonicalClientParams()
				clientState.ChainId = keepertest.DefaultRollapp
				k.SetClientState(ctx, clientID, &clientState)
				k.rollappKeeper.SetRollapp(ctx, keepertest.DefaultRollapp, types.Rollapp{})
				k.SetLatestHeight(ctx, keepertest.DefaultRollapp, 1)
				k.SetConsensusState(ctx, clientID, 1, &ibctm.ConsensusState{Timestamp: time.Now()})
				k.SetBlockDescriptor(ctx, keepertest.DefaultRollapp, 1, &types.BlockDescriptor{Height: 1, StateRoot: []byte("test")})
				return input{clientID: clientID}
			},
			expectErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k, ctx := keepertest.LightClientKeeper(t)
			input := tc.malleate(ctx, *k)
			err := k.TrySetCanonicalClient(ctx, input.clientID)
			if ((err != nil) != (tc.expectErr != nil)) && !errorsmod.IsOf(err, tc.expectErr) {
				t.Fatalf("expected error: %v, got: %v", tc.expectErr, err)
			}
		})
	}
}
