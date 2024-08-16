package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// RegisterName is message handler, handles registration of a new Dym-Name
// or extends the ownership duration of an existing Dym-Name.
func (k msgServer) RegisterName(goCtx context.Context, msg *dymnstypes.MsgRegisterName) (*dymnstypes.MsgRegisterNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	dymName, err := k.validateRegisterName(ctx, msg)
	if err != nil {
		return nil, err
	}

	priceParams := k.PriceParams(ctx)

	addDurationInSeconds := 86400 * 365 * msg.Duration

	firstYearPrice := priceParams.GetFirstYearDymNamePrice(msg.Name)

	var prunePreviousDymNameRecord bool
	var totalCost sdk.Coin
	if dymName == nil {
		// register new
		prunePreviousDymNameRecord = true

		dymName = &dymnstypes.DymName{
			Name:       msg.Name,
			Owner:      msg.Owner,
			Controller: msg.Owner,
			ExpireAt:   ctx.BlockTime().Unix() + addDurationInSeconds,
			Configs:    nil,
			Contact:    msg.Contact,
		}

		totalCost = sdk.NewCoin(
			priceParams.PriceDenom,
			firstYearPrice.Add( // first year has different price
				priceParams.PriceExtends.Mul(
					sdkmath.NewInt(
						msg.Duration-1, // subtract first year
					),
				),
			),
		)
	} else if dymName.Owner == msg.Owner {
		if dymName.IsExpiredAtCtx(ctx) {
			// renew
			prunePreviousDymNameRecord = true

			dymName = &dymnstypes.DymName{
				Name:       msg.Name,
				Owner:      msg.Owner,
				Controller: msg.Owner,
				ExpireAt:   ctx.BlockTime().Unix() + addDurationInSeconds,
				Configs:    nil,
				Contact:    msg.Contact,
			}
		} else {
			// extends
			prunePreviousDymNameRecord = false

			// just add duration, no need to change any existing configuration
			dymName.ExpireAt += addDurationInSeconds

			if msg.Contact != "" {
				// update contact if provided
				dymName.Contact = msg.Contact
			}
		}

		totalCost = sdk.NewCoin(
			priceParams.PriceDenom,
			priceParams.PriceExtends.Mul(
				sdkmath.NewInt(msg.Duration),
			),
		)
	} else {
		// take over
		prunePreviousDymNameRecord = true

		dymName = &dymnstypes.DymName{
			Name:       msg.Name,
			Owner:      msg.Owner,
			Controller: msg.Owner,
			ExpireAt:   ctx.BlockTime().Unix() + addDurationInSeconds,
			Configs:    nil,
			Contact:    msg.Contact,
		}

		totalCost = sdk.NewCoin(
			priceParams.PriceDenom,
			firstYearPrice.Add( // first year has different price
				priceParams.PriceExtends.Mul(
					sdkmath.NewInt(
						msg.Duration-1, // subtract first year
					),
				),
			),
		)
	}

	if !totalCost.IsPositive() {
		panic(errorsmod.Wrapf(gerrc.ErrFault, "total cost is not positive: %s", totalCost.String()))
	}

	if !totalCost.Equal(msg.ConfirmPayment) {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"actual payment is is different with provided by user: %s != %s", totalCost.String(), msg.ConfirmPayment,
		)
	}

	// At this place we don't do compare actual payment with estimated payment calculated by EstimateRegisterName
	// because in-case there is different between them, it would prevent user to registration/renew.

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		sdk.MustAccAddressFromBech32(msg.Owner),
		dymnstypes.ModuleName,
		sdk.NewCoins(totalCost),
	); err != nil {
		return nil, err
	}

	if err := k.bankKeeper.BurnCoins(ctx, dymnstypes.ModuleName, sdk.NewCoins(totalCost)); err != nil {
		return nil, err
	}

	if prunePreviousDymNameRecord {
		if err := k.PruneDymName(ctx, msg.Name); err != nil {
			return nil, err
		}
	}

	if err := k.SetDymName(ctx, *dymName); err != nil {
		return nil, err
	}

	if err := k.AfterDymNameOwnerChanged(ctx, dymName.Name); err != nil {
		return nil, err
	}

	if err := k.AfterDymNameConfigChanged(ctx, dymName.Name); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		dymnstypes.EventTypeSell,
		sdk.NewAttribute(dymnstypes.AttributeKeySellAssetType, dymnstypes.TypeName.PrettyName()),
		sdk.NewAttribute(dymnstypes.AttributeKeySellName, dymName.Name),
		sdk.NewAttribute(dymnstypes.AttributeKeySellPrice, totalCost.String()),
		sdk.NewAttribute(dymnstypes.AttributeKeySellTo, msg.Owner),
	))

	return &dymnstypes.MsgRegisterNameResponse{}, nil
}

// validateRegisterName handles validation for the message handled by RegisterName.
func (k msgServer) validateRegisterName(ctx sdk.Context, msg *dymnstypes.MsgRegisterName) (*dymnstypes.DymName, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	miscParams := k.MiscParams(ctx)

	dymName := k.GetDymName(ctx, msg.Name)
	if dymName != nil {
		if dymName.Owner == msg.Owner {
			// just renew or extends
		} else {
			if !dymName.IsExpiredAtCtx(ctx) {
				return nil, gerrc.ErrUnauthenticated
			}

			// take over

			// check grace period.
			// Grace period is the time period after the Dym-Name expired
			// that the previous owner can re-purchase the Dym-Name and no one else can take over.
			// This follow domain specification to prevent user mistake.
			dymNameCanBeTakeOverAfterEpoch := dymName.ExpireAt + int64(miscParams.GracePeriodDuration.Seconds())

			if ctx.BlockTime().Unix() < dymNameCanBeTakeOverAfterEpoch {
				// still in grace period
				return nil, errorsmod.Wrapf(
					gerrc.ErrFailedPrecondition,
					"can be taken over after: %s", time.Unix(dymNameCanBeTakeOverAfterEpoch, 0).UTC().Format(time.DateTime),
				)
			}

			// allowed to take over
		}
	}

	return dymName, nil
}

// EstimateRegisterName returns the estimated amount of coins required to register a new Dym-Name
// or extends the ownership duration of an existing Dym-Name.
func EstimateRegisterName(
	priceParams dymnstypes.PriceParams,
	name string,
	existingDymName *dymnstypes.DymName,
	newOwner string,
	duration int64,
) dymnstypes.EstimateRegisterNameResponse {
	var newFirstYearPrice, extendsPrice sdkmath.Int

	if existingDymName != nil && existingDymName.Owner == newOwner {
		// Dym-Name exists and just renew or extends by the same owner

		newFirstYearPrice = sdk.ZeroInt() // regardless of expired or not, we don't charge this
		extendsPrice = priceParams.PriceExtends.Mul(
			sdkmath.NewInt(duration),
		)
	} else {
		// new registration or take over
		newFirstYearPrice = priceParams.GetFirstYearDymNamePrice(name) // charge based on name length for the first year
		if duration > 1 {
			extendsPrice = priceParams.PriceExtends.Mul(
				sdkmath.NewInt(duration - 1), // subtract first year, which has different price
			)
		} else {
			extendsPrice = sdk.ZeroInt()
		}
	}

	return dymnstypes.EstimateRegisterNameResponse{
		FirstYearPrice: sdk.NewCoin(priceParams.PriceDenom, newFirstYearPrice),
		ExtendPrice:    sdk.NewCoin(priceParams.PriceDenom, extendsPrice),
		TotalPrice:     sdk.NewCoin(priceParams.PriceDenom, newFirstYearPrice.Add(extendsPrice)),
	}
}
