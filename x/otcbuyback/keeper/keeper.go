package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		logger   log.Logger

		authority string
		treasury  string

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
		ammKeeper     types.AMMKeeper

		// Collections for storing auction data
		nextAuctionID collections.Sequence
		auctions      collections.Map[uint64, types.Auction]
		purchases     collections.Map[collections.Pair[uint64, string], types.UserVestingPlan] // [auctionID, buyer] -> UserVestingPlan
		params        collections.Item[types.Params]
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	logger log.Logger,
	authority, treasury string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	ammKeeper types.AMMKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))

	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		logger:    logger,
		authority: authority,

		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		ammKeeper:     ammKeeper,

		nextAuctionID: collections.NewSequence(
			sb,
			types.NextAuctionIDKey,
			"next_auction_id",
		),
		auctions: collections.NewMap(
			sb,
			types.AuctionKeyPrefix,
			"auctions",
			collections.Uint64Key,
			collcompat.ProtoValue[types.Auction](cdc),
		),
		purchases: collections.NewMap(
			sb,
			types.PurchaseKeyPrefix,
			"purchases",
			collections.PairKeyCodec(collections.Uint64Key, collections.StringKey),
			collcompat.ProtoValue[types.UserVestingPlan](cdc),
		),
		params: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			collcompat.ProtoValue[types.Params](cdc),
		),
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetModuleAccountAddress returns the module account address
func (k Keeper) GetModuleAccountAddress() sdk.AccAddress {
	return k.accountKeeper.GetModuleAddress(types.ModuleName)
}

func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetAuction stores an auction using collections
func (k Keeper) SetAuction(ctx sdk.Context, auction types.Auction) error {
	return k.auctions.Set(ctx, auction.Id, auction)
}

// GetAuction retrieves an auction by ID using collections
func (k Keeper) GetAuction(ctx sdk.Context, auctionID uint64) (types.Auction, bool) {
	auction, err := k.auctions.Get(ctx, auctionID)
	if err != nil {
		return types.Auction{}, false
	}
	return auction, true
}

// GetAllAuctions retrieves all auctions from the store using collections
func (k Keeper) GetAllAuctions(ctx sdk.Context) ([]types.Auction, error) {
	var auctions []types.Auction
	err := k.auctions.Walk(ctx, nil, func(key uint64, auction types.Auction) (bool, error) {
		auctions = append(auctions, auction)
		return false, nil
	})
	return auctions, err
}

// DeleteAuction removes an auction from the store using collections
func (k Keeper) DeleteAuction(ctx sdk.Context, auctionID uint64) error {
	return k.auctions.Remove(ctx, auctionID)
}

// IncrementNextAuctionID increments and returns the next auction ID using collections
func (k Keeper) IncrementNextAuctionID(ctx sdk.Context) (uint64, error) {
	return k.nextAuctionID.Next(ctx)
}

// SetPurchase stores a purchase using collections
func (k Keeper) SetPurchase(ctx sdk.Context, auctionID uint64, buyer string, purchase types.UserVestingPlan) error {
	key := collections.Join(auctionID, buyer)
	return k.purchases.Set(ctx, key, purchase)
}

// GetPurchase retrieves a purchase by auction ID and buyer using collections
func (k Keeper) GetPurchase(ctx sdk.Context, auctionID uint64, buyer string) (types.UserVestingPlan, bool) {
	key := collections.Join(auctionID, buyer)
	purchase, err := k.purchases.Get(ctx, key)
	if err != nil {
		return types.UserVestingPlan{}, false
	}
	return purchase, true
}

// GetUserPurchases retrieves all purchases for a specific user using collections
func (k Keeper) GetUserPurchases(ctx sdk.Context, buyer string) ([]types.UserVestingPlan, error) {
	var purchases []types.UserVestingPlan
	err := k.purchases.Walk(ctx, nil, func(key collections.Pair[uint64, string], purchase types.UserVestingPlan) (bool, error) {
		// Filter by buyer (second element of the pair)
		if key.K2() == buyer {
			purchases = append(purchases, purchase)
		}
		return false, nil
	})
	return purchases, err
}

// GetAuctionPurchases retrieves all purchases for a specific auction using collections
func (k Keeper) GetAuctionPurchases(ctx sdk.Context, auctionID uint64) ([]types.UserVestingPlan, error) {
	var purchases []types.UserVestingPlan
	rng := collections.NewPrefixedPairRange[uint64, string](auctionID)
	err := k.purchases.Walk(ctx, rng, func(key collections.Pair[uint64, string], purchase types.UserVestingPlan) (bool, error) {
		purchases = append(purchases, purchase)
		return false, nil
	})
	return purchases, err
}
