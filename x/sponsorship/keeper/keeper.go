package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

type raEndorsementsIndexes struct {
	// first parameter: uint64 is a RA gauge ID
	// second parameter: string is a RA ID
	// third parameter: types.Endorsement is a RA endorsement
	raGaugeID *indexes.Unique[uint64, string, types.Endorsement]
}

func newRAEndorsementIndexes(sb *collections.SchemaBuilder) raEndorsementsIndexes {
	return raEndorsementsIndexes{
		raGaugeID: indexes.NewUnique(
			sb, types.RAGaugeIDIndexPrefix(), "endorsements_by_ra_gauge_id",
			collections.Uint64Key, collections.StringKey,
			func(_ string, v types.Endorsement) (uint64, error) {
				return v.RollappGaugeId, nil
			},
		),
	}
}

func (i raEndorsementsIndexes) IndexesList() []collections.Index[string, types.Endorsement] {
	return []collections.Index[string, types.Endorsement]{i.raGaugeID}
}

type Keeper struct {
	authority string // authority is the x/gov module account

	schema                  collections.Schema
	params                  collections.Item[types.Params]
	delegatorValidatorPower collections.Map[collections.Pair[sdk.AccAddress, sdk.ValAddress], math.Int]
	distribution            collections.Item[types.Distribution]
	votes                   collections.Map[sdk.AccAddress, types.Vote]
	// rollapp ID -> types.Endorsement mapping
	// also, it has index for rollapp gauge ID -> types.Endorsement mapping
	raEndorsements *collections.IndexedMap[string, types.Endorsement, raEndorsementsIndexes]
	// the list of the users who do not have the right to claim rewards on this epoch
	// the index is refreshed every epoch
	claimBlacklist collections.KeySet[sdk.AccAddress]

	stakingKeeper    types.StakingKeeper
	incentivesKeeper types.IncentivesKeeper
	bankKeeper       types.BankKeeper
}

// NewKeeper returns a new instance of the x/sponsorship keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ak types.AccountKeeper,
	sk types.StakingKeeper,
	ik types.IncentivesKeeper,
	bk types.BankKeeper,
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

	k := Keeper{
		authority: authority,
		schema:    collections.Schema{}, // set later
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
		raEndorsements: collections.NewIndexedMap(
			sb,
			types.RAEndorsementsPrefix(),
			"endorsements",
			collections.StringKey,
			codec.CollValue[types.Endorsement](cdc),
			newRAEndorsementIndexes(sb),
		),
		claimBlacklist: collections.NewKeySet(
			sb,
			types.ClaimBlacklistPrefix(),
			"claim_blacklist",
			sdk.AccAddressKey,
		),
		stakingKeeper:    sk,
		incentivesKeeper: ik,
		bankKeeper:       bk,
	}

	// SchemaBuilder CANNOT be used after Build is called,
	// so we build it after all collections are initialized
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.schema = schema

	return k
}

func (k Keeper) Schema() collections.Schema {
	return k.schema
}
