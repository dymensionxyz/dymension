package keeper_test

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	keeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
)

func TestTrySetCanonicalClient(t *testing.T) {

	type input struct {
		clientID string
	}

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context, k keeper.Keeper) input
		expectErr error
	}{}
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
