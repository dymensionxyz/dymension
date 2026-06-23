package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

func (k Keeper) ValidateAttestation(ctx sdk.Context, nonce, token string) error {
	policy := k.GetParams(ctx).TeeConfig.Policy()
	return tee.NewVerifier().Verify(ctx, policy, nonce, token)
}
