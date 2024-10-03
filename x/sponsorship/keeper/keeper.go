package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

type Keeper struct {
	authority string // authority is the x/gov module account

	params                  collections.Item[types.Params]
	delegatorValidatorPower collections.Map[collections.Pair[sdk.AccAddress, sdk.ValAddress], math.Int]
	distribution            collections.Item[types.Distribution]
	votes                   collections.Map[sdk.AccAddress, types.Vote]

	stakingKeeper    types.StakingKeeper
	incentivesKeeper types.IncentivesKeeper
	sequencerKeeper  types.SequencerKeeper
}

// NewKeeper returns a new instance of the x/sponsorship keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ak types.AccountKeeper,
	sk types.StakingKeeper,
	ik types.IncentivesKeeper,
	sqk types.SequencerKeeper,
	authority string,
) Keeper {
	// ensure the module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the x/sponsorship module account has not been set")
	}

	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid x/sponsorship authority address: %w", err))
	}

	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))

	return Keeper{
		authority: authority,
		params: collections.NewItem(
			sb,
			types.ParamsPrefix(),
			"params",
			collcompat.ProtoValue[types.Params](cdc),
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
		distribution: collections.NewItem(
			sb,
			types.DistributionPrefix(),
			"distribution",
			collcompat.ProtoValue[types.Distribution](cdc),
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
		sequencerKeeper:  sqk,
	}
}
