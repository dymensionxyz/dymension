package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

const (
	sellOrderExpirationInvariantName = "sell-order-expiration"
)

// RegisterInvariants registers the DymNS module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(dymnstypes.ModuleName, sellOrderExpirationInvariantName, SellOrderExpirationInvariant(k))
}

// SellOrderExpirationInvariant checks that the records in the `ActiveSellOrdersExpiration` (aSoe) records are valid.
// The `ActiveSellOrdersExpiration` is used in hook and records are fetch from the store frequently,
// so we need to make sure the records are correct.
func SellOrderExpirationInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		blockEpochUTC := ctx.BlockTime().Unix()

		for _, assetType := range []dymnstypes.AssetType{dymnstypes.TypeName, dymnstypes.TypeAlias} {
			activeSellOrdersExpiration := k.GetActiveSellOrdersExpiration(ctx, assetType)
			for _, record := range activeSellOrdersExpiration.Records {
				if record.ExpireAt > blockEpochUTC {
					// skip not expired ones.
					// Why we skip expired ones? Because the hook only process the expired ones,
					// so it is not necessary to check the not expired ones, to reduce store read.
					continue
				}

				sellOrder := k.GetSellOrder(ctx, record.AssetId, assetType)
				if sellOrder == nil {
					msg += fmt.Sprintf(
						"sell order not found: assetId(%s), assetType(%s), expiry(%d)\n",
						record.AssetId, assetType.PrettyName(),
						record.ExpireAt,
					)
					broken = true
				} else if sellOrder.ExpireAt != record.ExpireAt {
					msg += fmt.Sprintf(
						"sell order expiration mismatch: assetId(%s), assetType(%s), expiry(%d) != actual(%d)\n",
						record.AssetId, assetType.PrettyName(),
						record.ExpireAt, sellOrder.ExpireAt,
					)
					broken = true
				}
			}
		}
		return sdk.FormatInvariant(dymnstypes.ModuleName, sellOrderExpirationInvariantName, msg), broken
	}
}
