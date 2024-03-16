package keeper

import (
	"fmt"

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
	if err := k.mintRollappGenesisTokens(ctx, rollapp.GenesisState.GenesisAccounts, rollapp.ChannelId); err != nil {
		return err
	}
	rollapp.GenesisState.IsGenesisEvent = true
	k.SetRollapp(ctx, rollapp)
	return nil
}

func (k Keeper) mintRollappGenesisTokens(ctx sdk.Context, accounts []types.GenesisAccount, channelId string) error {
	for _, acc := range accounts {
		denomTrace := utils.GetForeignDenomTrace(channelId, acc.Amount.Denom)
		traceHash := denomTrace.Hash()
		// if the denom trace does not exist, add it
		if !k.transferKeeper.HasDenomTrace(ctx, traceHash) {
			k.transferKeeper.SetDenomTrace(ctx, denomTrace)
		}

		ibcDenom := denomTrace.IBCDenom()
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
