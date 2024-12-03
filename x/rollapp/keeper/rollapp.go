package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transferTypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	udenom "github.com/dymensionxyz/dymension/v3/utils/denom"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) CheckAndUpdateRollappFields(ctx sdk.Context, update *types.MsgUpdateRollappInformation) (types.Rollapp, error) {
	current, found := k.GetRollapp(ctx, update.RollappId)
	if !found {
		return current, types.ErrRollappNotFound
	}

	if update.Owner != current.Owner {
		return current, sdkerrors.ErrUnauthorized
	}

	// immutable values cannot be updated when the rollapp is launched
	if update.UpdatingImmutableValues() && current.Launched {
		return current, types.ErrImmutableFieldUpdateAfterLaunched
	}

	if update.UpdatingGenesisInfo() && current.GenesisInfo.Sealed {
		return current, types.ErrGenesisInfoSealed
	}

	if update.InitialSequencer != "" {
		current.InitialSequencer = update.InitialSequencer
	}

	if types.IsUpdateMinSeqBond(update.MinSequencerBond) {
		if err := k.validMinBond(ctx, update.MinSequencerBond); err != nil {
			return current, errorsmod.Wrap(err, "valid min bond")
		}
		current.MinSequencerBond = sdk.NewCoins(update.MinSequencerBond)
	}

	if update.GenesisInfo != nil {
		if update.GenesisInfo.GenesisChecksum != "" {
			current.GenesisInfo.GenesisChecksum = update.GenesisInfo.GenesisChecksum
		}

		if update.GenesisInfo.Bech32Prefix != "" {
			current.GenesisInfo.Bech32Prefix = update.GenesisInfo.Bech32Prefix
		}

		if update.GenesisInfo.NativeDenom.Base != "" {
			current.GenesisInfo.NativeDenom = update.GenesisInfo.NativeDenom
		}

		if !update.GenesisInfo.InitialSupply.IsNil() {
			current.GenesisInfo.InitialSupply = update.GenesisInfo.InitialSupply
		}

		// Frontend always passes new value
		current.GenesisInfo.GenesisAccounts = update.GenesisInfo.GenesisAccounts
	}

	if update.Metadata != nil && !update.Metadata.IsEmpty() {
		current.Metadata = update.Metadata
	}

	if err := current.ValidateBasic(); err != nil {
		return current, fmt.Errorf("validate rollapp: %w", err)
	}

	return current, nil
}

// CheckIfRollappExists checks if a rollapp with the same ID or alias already exists in the store.
// An exception is made for when the rollapp is frozen, in which case it is allowed to replace the existing rollapp (forking).
func (k Keeper) CheckIfRollappExists(ctx sdk.Context, rollappId types.ChainID) error {
	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, rollappId.GetChainID()); isFound {
		return types.ErrRollappExists
	}

	if _, isFound := k.GetRollappByEIP155(ctx, rollappId.GetEIP155ID()); isFound {
		return types.ErrRollappExists
	}

	if _, isFound := k.GetRollappByName(ctx, rollappId.GetName()); isFound {
		return types.ErrRollappExists
	}

	return nil
}

// SetRollapp set a specific rollapp in the store from its index
func (k Keeper) SetRollapp(ctx sdk.Context, rollapp types.Rollapp) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	b := k.cdc.MustMarshal(&rollapp)
	store.Set(types.RollappKey(
		rollapp.RollappId,
	), b)

	// no err check as rollapp is already validated
	rollappID := types.MustNewChainID(rollapp.RollappId)

	store = prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByEIP155KeyPrefix))
	store.Set(types.RollappByEIP155Key(
		rollappID.GetEIP155ID(),
	), []byte(rollapp.RollappId))
}

func (k Keeper) SetRollappAsLaunched(ctx sdk.Context, rollapp *types.Rollapp) error {
	if !rollapp.AllImmutableFieldsAreSet() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "immutable fields not set")
	}

	rollapp.GenesisInfo.Sealed = true
	rollapp.Launched = true
	k.SetRollapp(ctx, *rollapp)

	return nil
}

// SetIROPlanToRollapp modifies the rollapp object due to IRO creation
// This methods:
// - seals the rollapp genesis info
// - set the pre launch time according to the iro plan end time
// Validations:
// - rollapp must not be launched
// - genesis info must be set
// NOTE: we already validated that a genesis account exists for the IRO plan
func (k Keeper) SetIROPlanToRollapp(ctx sdk.Context, rollapp *types.Rollapp, iro irotypes.Plan) error {
	if rollapp.Launched {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp already launched")
	}

	if rollapp.GenesisInfo.Sealed {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis info already sealed")
	}

	if !rollapp.GenesisInfoFieldsAreSet() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis info not set")
	}

	// seal genesis info
	rollapp.GenesisInfo.Sealed = true

	// set pre launch time
	rollapp.PreLaunchTime = &iro.PreLaunchTime

	k.SetRollapp(ctx, *rollapp)
	return nil
}

// GetRollappByEIP155 returns a rollapp from its EIP155 id (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md)
func (k Keeper) GetRollappByEIP155(ctx sdk.Context, eip155 uint64) (val types.Rollapp, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByEIP155KeyPrefix))
	id := store.Get(types.RollappByEIP155Key(
		eip155,
	))
	if id == nil {
		return val, false
	}

	return k.GetRollapp(ctx, string(id))
}

// GetRollapp returns a rollapp from its chain name
func (k Keeper) GetRollapp(
	ctx sdk.Context,
	rollappId string,
) (val types.Rollapp, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))

	b := store.Get(types.RollappKey(
		rollappId,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) GetRollappByName(
	ctx sdk.Context,
	name string,
) (val types.Rollapp, found bool) {
	name = name + "_"
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte(name))

	defer iterator.Close() // nolint: errcheck

	if !iterator.Valid() {
		return val, false
	}

	k.cdc.MustUnmarshal(iterator.Value(), &val)
	return val, true
}

// GetRollappByDenom tries to extract a rollapp ID from the provided denom and returns a rollapp object if found.
// Denom may be either IRO token or IBC token.
func (k Keeper) GetRollappByDenom(ctx sdk.Context, denom string) (*types.Rollapp, error) {
	// by IRO token
	// try to get the rollapp ID from the denom
	rollappID, ok := irotypes.RollappIDFromIRODenom(denom)
	if ok {
		ra, ok := k.GetRollapp(ctx, rollappID)
		if ok {
			return &ra, nil
		}
		return nil, types.ErrUnknownRollappID
	}

	// by IBC token
	// first, validate that the denom is IBC
	hexHash, ok := udenom.ValidateIBCDenom(denom)
	if !ok {
		return nil, errors.New("denom is neither IRO nor IBC")
	}

	// parse IBC denom hash string
	hash, err := transferTypes.ParseHexHash(hexHash)
	if err != nil {
		return nil, fmt.Errorf("parse IBC hex hash: %w", err)
	}
	// get IBC denom trace
	trace, ok := k.transferKeeper.GetDenomTrace(ctx, hash)
	if !ok {
		return nil, errors.New("denom trace not found")
	}
	// try to get source port and channel from the trace
	sourcePort, sourceChan, ok := udenom.SourcePortChanFromTracePath(trace.Path)
	if !ok {
		return nil, errors.New("invalid denom trace path")
	}

	return k.GetRollappByPortChan(ctx, sourcePort, sourceChan)
}

func (k Keeper) GetRollappOwnerByDenom(ctx sdk.Context, denom string) (sdk.AccAddress, error) {
	ra, err := k.GetRollappByDenom(ctx, denom)
	if err != nil {
		return nil, fmt.Errorf("get rollapp by denom: %w", err)
	}
	owner, err := sdk.AccAddressFromBech32(ra.Owner)
	if err != nil {
		return nil, fmt.Errorf("owner account address: %w", err)
	}
	return owner, nil
}

func (k Keeper) MustGetRollappOwner(ctx sdk.Context, rollappID string) sdk.AccAddress {
	ra := k.MustGetRollapp(ctx, rollappID)
	return sdk.MustAccAddressFromBech32(ra.Owner)
}

func (k Keeper) MustGetRollapp(ctx sdk.Context, rollappId string) types.Rollapp {
	ret, found := k.GetRollapp(ctx, rollappId)
	if !found {
		panic(fmt.Sprintf("rollapp not found: id: %s", rollappId))
	}
	return ret
}

// RemoveRollapp removes a rollapp from the store using rollapp name
func (k Keeper) RemoveRollapp(
	ctx sdk.Context,
	rollappId string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	store.Delete(types.RollappKey(
		rollappId,
	))
}

// GetAllRollapps returns all rollapp
func (k Keeper) GetAllRollapps(ctx sdk.Context) []types.Rollapp {
	return k.FilterRollapps(ctx, func(rollapp types.Rollapp) bool { return true })
}

// IsRollappStarted returns true if the rollapp is started
func (k Keeper) IsRollappStarted(ctx sdk.Context, rollappId string) bool {
	_, found := k.GetLatestStateInfoIndex(ctx, rollappId)
	return found
}

func (k Keeper) FilterRollapps(ctx sdk.Context, f func(types.Rollapp) bool) []types.Rollapp {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	var result []types.Rollapp
	for ; iterator.Valid(); iterator.Next() {
		var val types.Rollapp
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if f(val) {
			result = append(result, val)
		}
	}
	return result
}

func (k Keeper) IsDRSVersionObsolete(ctx sdk.Context, version uint32) bool {
	ok, err := k.obsoleteDRSVersions.Has(ctx, version)
	if err != nil {
		panic(fmt.Sprintf("checking if DRS version is obsolete: %v", err))
	}
	return ok
}

func (k Keeper) SetObsoleteDRSVersion(ctx sdk.Context, version uint32) error {
	return k.obsoleteDRSVersions.Set(ctx, version)
}

func (k Keeper) GetAllObsoleteDRSVersions(ctx sdk.Context) ([]uint32, error) {
	iter, err := k.obsoleteDRSVersions.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	return iter.Keys()
}
