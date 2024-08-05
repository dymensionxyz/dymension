package keeper

import (
	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	storeService            corestoretypes.KVStoreService
	params                  collections.Item[types.Params]
	delegatorValidatorPower collections.Map[collections.Pair[sdk.AccAddress, sdk.ValAddress], math.Int]
	distribution            collections.Item[types.Distribution]
	votes                   collections.Map[sdk.AccAddress, types.Vote]

	stakingKeeper    types.StakingKeeper
	incentivesKeeper types.IncentivesKeeper
}

// NewKeeper returns a new instance of the x/sponsorship keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	ak types.AccountKeeper,
	sk types.StakingKeeper,
	ik types.IncentivesKeeper,
) Keeper {
	// ensure the module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the x/sponsorship module account has not been set")
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))

	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		paramSpace:   paramSpace,
		storeService: nil,
		params: collections.NewItem(
			sb,
			types.ParamsPrefix(),
			"params",
			collcompat.ProtoValue[types.Params](cdc),
		),
		distribution: collections.NewItem(
			sb,
			types.DistributionPrefix(),
			"distribution",
			collcompat.ProtoValue[types.Distribution](cdc),
		),
		delegatorValidatorPower: collections.NewMap(
			sb,
			types.DelegatorValidatorPrefix(),
			"delegator_validators",
			collections.PairKeyCodec(
				collcompat.AccAddressKey,
				collcompat.ValAddressKey,
			),
			collcompat.IntValue,
		),
		votes: collections.NewMap(
			sb,
			types.VotePrefix(),
			"votes",
			collcompat.AccAddressKey,
			collcompat.ProtoValue[types.Vote](cdc),
		),
		stakingKeeper:    sk,
		incentivesKeeper: ik,
	}
}
