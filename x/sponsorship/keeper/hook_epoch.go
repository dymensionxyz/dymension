package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
)

type EpochHooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = EpochHooks{}

func (k Keeper) EpochHooks() EpochHooks {
	return EpochHooks{k}
}

func (h EpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) error {
	params, err := h.k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("get sponsorship params: %w", err)
	}

	if epochIdentifier != params.EpochIdentifier {
		return nil
	}

	es, err := h.k.GetAllEndorsements(ctx)
	if err != nil {
		return fmt.Errorf("get all endorsements: %w", err)
	}

	for _, e := range es {
		e.EpochShares = e.TotalShares
		err = h.k.SaveEndorsement(ctx, e)
		if err != nil {
			return fmt.Errorf("save endorsement: rollappId %s: %w", e.RollappId, err)
		}
	}

	err = h.k.RefreshClaimBlacklist(ctx)
	if err != nil {
		return fmt.Errorf("refresh claim blacklist: %w", err)
	}

	return nil
}

func (h EpochHooks) BeforeEpochStart(sdk.Context, string, int64) error {
	return nil
}
