package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey

		baseDenom string

		authority string

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
		ammKeeper     types.AMMKeeper

		// Collections for storing auction data
		nextAuctionID collections.Sequence
		auctions      collections.Map[uint64, types.Auction]
		purchases     collections.Map[collections.Pair[uint64, string], types.VestingPlan] // [auctionID, buyer] -> VestingPlan
		params        collections.Item[types.Params]

		// acceptedTokens stores accepted tokens and their pool IDs
		acceptedTokens collections.Map[string, uint64] // [denom] -> poolID
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	ammKeeper types.AMMKeeper,
) Keeper {

	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))

	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,

		baseDenom: params.BaseDenom, // set "adym" as the base denom

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
			collcompat.ProtoValue[types.VestingPlan](cdc),
		),
		params: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			collcompat.ProtoValue[types.Params](cdc),
		),
		acceptedTokens: collections.NewMap(
			sb,
			types.AcceptedTokensKeyPrefix,
			"accepted_tokens",
			collections.StringKey,
			collections.Uint64Value,
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

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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
func (k Keeper) SetPurchase(ctx sdk.Context, auctionID uint64, buyer string, purchase types.VestingPlan) error {
	key := collections.Join(auctionID, buyer)
	return k.purchases.Set(ctx, key, purchase)
}

// GetPurchase retrieves a purchase by auction ID and buyer using collections
func (k Keeper) GetPurchase(ctx sdk.Context, auctionID uint64, buyer string) (types.VestingPlan, bool) {
	key := collections.Join(auctionID, buyer)
	purchase, err := k.purchases.Get(ctx, key)
	if err != nil {
		return types.VestingPlan{}, false
	}
	return purchase, true
}

// SetAcceptedToken stores an accepted token and its pool ID
func (k Keeper) SetAcceptedToken(ctx sdk.Context, token string, poolID uint64) error {
	return k.acceptedTokens.Set(ctx, token, poolID)
}

// GetAcceptedTokenPoolID retrieves the pool ID for a given token
func (k Keeper) GetAcceptedTokenPoolID(ctx sdk.Context, token string) (uint64, error) {
	return k.acceptedTokens.Get(ctx, token)
}

func (k Keeper) IsAcceptedDenom(ctx sdk.Context, denom string) bool {
	ok, _ := k.acceptedTokens.Has(ctx, denom)
	return ok
}

// GetAllAcceptedTokens retrieves all accepted tokens
func (k Keeper) GetAllAcceptedTokens(ctx sdk.Context) ([]types.AcceptedToken, error) {
	var acceptedTokens []types.AcceptedToken
	err := k.acceptedTokens.Walk(ctx, nil, func(token string, poolID uint64) (bool, error) {
		acceptedTokens = append(acceptedTokens, types.AcceptedToken{
			Token:  token,
			PoolId: poolID,
		})
		return false, nil
	})
	return acceptedTokens, err
}

// SetAcceptedTokens replaces all accepted tokens with the provided list
func (k Keeper) SetAcceptedTokens(ctx sdk.Context, tokens []types.AcceptedToken) error {
	// Clear existing tokens
	if err := k.acceptedTokens.Clear(ctx, nil); err != nil {
		return err
	}

	// Set new tokens
	for _, token := range tokens {
		if err := k.acceptedTokens.Set(ctx, token.Token, token.PoolId); err != nil {
			return err
		}
	}
	return nil
}

// ClearAcceptedTokens removes all accepted tokens
func (k Keeper) ClearAcceptedTokens(ctx sdk.Context) error {
	return k.acceptedTokens.Clear(ctx, nil)
}
