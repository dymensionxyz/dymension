package keeper

import (
	"fmt"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		hooks      types.MultiRollappHooks
		paramstore paramtypes.Subspace

		ibcclientKeeper types.IBCClientKeeper
		transferKeeper  types.TransferKeeper
		channelKeeper   types.ChannelKeeper
		bankKeeper      types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	ibcclientKeeper types.IBCClientKeeper,
	transferKeeper types.TransferKeeper,
	channelKeeper types.ChannelKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		paramstore:      ps,
		hooks:           nil,
		ibcclientKeeper: ibcclientKeeper,
		transferKeeper:  transferKeeper,
		channelKeeper:   channelKeeper,
		bankKeeper:      bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// TriggerGenesisEvent triggers the genesis event for the rollapp.
func (k Keeper) TriggerRollappGenesisEvent(ctx sdk.Context, rollapp types.Rollapp) error {
	// Validate it hasn't been triggered yet and the gensis event exist
	switch {
	case rollapp.GenesisState == nil:
		return types.ErrGenesisEventNotDefined
	case rollapp.GenesisState.IsGenesisEvent:
		return types.ErrGenesisEventAlreadyTriggered
	}
	// Call the mint genesis tokens function
	if err := k.mintRollappGenesisTokens(ctx, rollapp); err != nil {
		return err
	}
	rollapp.GenesisState.IsGenesisEvent = true
	k.SetRollapp(ctx, rollapp)
	return nil
}

func (k Keeper) mintRollappGenesisTokens(ctx sdk.Context, rollapp types.Rollapp) error {
	for _, acc := range rollapp.GenesisState.GenesisAccounts {
		denomTrace := utils.GetForeignDenomTrace(rollapp.ChannelId, acc.Amount.Denom)
		traceHash := denomTrace.Hash()
		// if the denom trace does not exist, add it
		if !k.transferKeeper.HasDenomTrace(ctx, traceHash) {
			k.transferKeeper.SetDenomTrace(ctx, denomTrace)
		}

		ibcDenom := denomTrace.IBCDenom()

		k.registerDenomOnBank(ctx, rollapp, ibcDenom, acc.Amount.Denom)

		coinsToMint := sdk.NewCoins(sdk.NewCoin(ibcDenom, acc.Amount.Amount))

		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
			return err
		}
		accAddress, err := sdk.AccAddressFromBech32(acc.Address)
		if err != nil {
			return err
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddress, coinsToMint); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) registerDenomOnBank(ctx sdk.Context, rollapp types.Rollapp, ibcDenom, accDenom string) {
	// if the denom metadata already exists, return
	if k.bankKeeper.HasDenomMetaData(ctx, ibcDenom) {
		return
	}
	for i := range rollapp.TokenMetadata {
		// if the rollapp token base matches the accDenom, create a new token denom metadata
		if rollapp.TokenMetadata[i].Base == accDenom {
			// create a new token denom metadata where it's base = ibcDenom,
			// and the rest of the fields are taken from rollapp.metadata
			metadata := banktypes.Metadata{
				Description: "auto-generated metadata for " + ibcDenom + " from rollapp " + rollapp.RollappId,
				Base:        ibcDenom,
				DenomUnits:  make([]*banktypes.DenomUnit, len(rollapp.TokenMetadata[i].DenomUnits)),
				Display:     rollapp.TokenMetadata[i].Display,
				Name:        rollapp.TokenMetadata[i].Name,
				Symbol:      rollapp.TokenMetadata[i].Symbol,
				URI:         rollapp.TokenMetadata[i].URI,
				URIHash:     rollapp.TokenMetadata[i].URIHash,
			}
			// Copy DenomUnits slice
			for j, du := range rollapp.TokenMetadata[i].DenomUnits {
				newDu := banktypes.DenomUnit{
					Aliases:  du.Aliases,
					Denom:    du.Denom,
					Exponent: du.Exponent,
				}
				// base denom_unit should be the same as baseDenom
				if newDu.Exponent == 0 {
					newDu.Denom = ibcDenom
					newDu.Aliases = append(newDu.Aliases, du.Denom)
				}
				metadata.DenomUnits[j] = &newDu
			}

			k.bankKeeper.SetDenomMetaData(ctx, metadata)

			k.Logger(ctx).Info("registered denom metadata for IBC token", "rollappID", rollapp.RollappId, "denom", ibcDenom)
		}
	}
}

/* -------------------------------------------------------------------------- */
/*                                    Hooks                                   */
/* -------------------------------------------------------------------------- */

// Set the rollapp hooks
func (k *Keeper) SetHooks(sh types.MultiRollappHooks) {
	if k.hooks != nil {
		panic("cannot set rollapp hooks twice")
	}
	k.hooks = sh
}

func (k *Keeper) GetHooks() types.MultiRollappHooks {
	return k.hooks
}
