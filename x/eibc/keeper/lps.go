package keeper

import (
	"errors"
	"math/rand/v2"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

var (
	LPsByRollAppDenomPrefix = collections.NewPrefix("lps0")
	LPsByIDPrefix           = collections.NewPrefix("lps1")
	LPsNextIDPrefix         = collections.NewPrefix("lps2")
	LPsByAddrPrefix         = collections.NewPrefix("lps3")
)

type LPs struct {
	// <rollapp,denom,id>
	byRollAppDenom collections.KeySet[collections.Triple[string, string, uint64]]
	// id -> lp
	byID collections.Map[uint64, types.OnDemandLPRecord]
	// <addr,id>
	byAddr collections.KeySet[collections.Pair[string, uint64]]
	nextID collections.Sequence
}

func makeLPsStore(sb *collections.SchemaBuilder, cdc codec.BinaryCodec) LPs {
	return LPs{
		byRollAppDenom: collections.NewKeySet[collections.Triple[string, string, uint64]](
			sb, LPsByRollAppDenomPrefix, "byRollAppDenom",
			collections.TripleKeyCodec[string, string, uint64](
				collections.StringKey,
				collections.StringKey,
				collections.Uint64Key,
			)),
		byID: collections.NewMap[uint64, types.OnDemandLPRecord](
			sb, LPsByIDPrefix, "byID",
			collections.Uint64Key, codec.CollValue[types.OnDemandLPRecord](cdc),
		),
		byAddr: collections.NewKeySet[collections.Pair[string, uint64]](
			sb, LPsByAddrPrefix, "byAddr",
			collections.PairKeyCodec[string, uint64](
				collections.StringKey,
				collections.Uint64Key,
			),
		),
		nextID: collections.NewSequence(sb, LPsNextIDPrefix, "nextID"),
	}
}

// create a new one
func (s LPs) Create(ctx sdk.Context, lp *types.OnDemandLP) (uint64, error) {
	id, err := s.nextID.Next(ctx)
	if err != nil {
		return 0, errorsmod.Wrap(err, "next id")
	}
	if err := s.Set(ctx, types.OnDemandLPRecord{
		Id:    id,
		Lp:    lp,
		Spent: math.ZeroInt(),
	}); err != nil {
		return 0, errorsmod.Wrap(err, "set")
	}
	if err := uevent.EmitTypedEvent(ctx, &types.EventCreatedOnDemandLP{
		Id:        id,
		FundsAddr: lp.FundsAddr,
	}); err != nil {
		return 0, errorsmod.Wrap(err, "event")
	}
	return id, nil
}

func (s LPs) Set(ctx sdk.Context, lp types.OnDemandLPRecord) error {
	err := s.byID.Set(ctx, lp.Id, lp)
	if err != nil {
		return errorsmod.Wrap(err, "set id")
	}
	err = s.byAddr.Set(ctx, collections.Join(lp.Lp.FundsAddr, lp.Id))
	if err != nil {
		return errorsmod.Wrap(err, "set by addr")
	}
	err = s.byRollAppDenom.Set(ctx, collections.Join3(lp.Lp.Rollapp, lp.Lp.Denom, lp.Id))
	if err != nil {
		return errorsmod.Wrap(err, "set by rollapp denom")
	}
	return nil
}

func (s LPs) Get(ctx sdk.Context, id uint64) (*types.OnDemandLPRecord, error) {
	ret, err := s.byID.Get(ctx, id)
	return &ret, err
}

func (s LPs) GetAll(ctx sdk.Context) ([]*types.OnDemandLPRecord, error) {
	it, err := s.byID.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	var ret []*types.OnDemandLPRecord
	for ; it.Valid(); it.Next() {
		v, err := it.Value()
		if err != nil {
			return nil, err
		}
		ret = append(ret, &v)
	}
	return ret, nil
}

// reason is human-readable string for debugging/ux
func (s LPs) Del(ctx sdk.Context, id uint64, reason string) error {
	lp, err := s.byID.Get(ctx, id)
	if err != nil {
		return errorsmod.Wrap(err, "get id")
	}
	err = s.byID.Remove(ctx, id)
	if err != nil {
		return errorsmod.Wrap(err, "remove id")
	}
	err = s.byRollAppDenom.Remove(ctx, collections.Join3(lp.Lp.Rollapp, lp.Lp.Denom, id))
	if err != nil {
		return errorsmod.Wrap(err, "remove by rollapp denom")
	}
	err = s.byAddr.Remove(ctx, collections.Join(lp.Lp.FundsAddr, lp.Id))
	if err != nil {
		return errorsmod.Wrap(err, "remove by addr")
	}
	if err := uevent.EmitTypedEvent(ctx, &types.EventDeletedOnDemandLP{
		Id:        id,
		FundsAddr: lp.Lp.FundsAddr,
		Reason:    reason,
	}); err != nil {
		return errorsmod.Wrap(err, "event")
	}
	return nil
}

func (s LPs) GetByAddr(ctx sdk.Context, addr sdk.AccAddress) ([]*types.OnDemandLPRecord, error) {
	var ret []*types.OnDemandLPRecord
	rng := collections.NewPrefixedPairRange[string, uint64](addr.String())
	iter, err := s.byAddr.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		id := key.K2()
		lp, err := s.byID.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &lp)
	}
	return ret, err
}

func (s LPs) GetOrderCompatibleLPs(ctx sdk.Context, o types.DemandOrder) ([]types.OnDemandLPRecord, error) {
	rol := o.RollappId
	denom := o.Denom()
	ranger := collections.NewSuperPrefixedTripleRange[string, string, uint64](rol, denom)
	iter, err := s.byRollAppDenom.Iterate(ctx, ranger)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var compat []types.OnDemandLPRecord
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		id := key.K3()

		lpr, err := s.byID.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if lpr.Accepts(uint64(ctx.BlockHeight()), o) {
			compat = append(compat, lpr)
		}
	}
	return compat, nil
}

func (k Keeper) FulfillByOnDemandLP(ctx sdk.Context, order string, rng uint64) error {
	o, err := k.GetOutstandingOrder(ctx, order)
	if err != nil {
		return errorsmod.Wrap(err, "get outstanding order")
	}
	lps, err := k.LPs.GetOrderCompatibleLPs(ctx, *o)
	if err != nil {
		return errorsmod.Wrap(err, "get compatible lp")
	}
	r := rand.New(rand.NewPCG(rng,0))
	r.Shuffle(len(lps), func(i, j int) {
		lps[i], lps[j] = lps[j], lps[i]
	})
	for _, lp := range lps {
		err := k.Fulfill(ctx, o, lp.Lp.MustAddr())
		if err != nil {
			if errorsmod.IsOf(err, sdkerrors.ErrInsufficientFunds) {
				if err := k.LPs.Del(ctx, lp.Id, "out of funds"); err != nil {
					return errorsmod.Wrapf(err, "delete lp: %d", lp.Id)
				}
				ctx.Logger().Error("Fulfill via on demand dlp - insufficient funds.", "lp", lp.Id)
				continue
			}
			return errorsmod.Wrap(err, "fulfill lp")
		}
		if err = uevent.EmitTypedEvent(ctx, &types.EventMatchedOnDemandLP{
			OrderId:   o.Id,
			LpId:      lp.Id,
			Fulfiller: lp.Lp.MustAddr().String(),
		}); err != nil {
			return errorsmod.Wrap(err, "emit event")
		}
		lp.Spent = lp.Spent.Add(o.PriceAmount())
		if err = k.LPs.Set(ctx, lp); err != nil {
			return errorsmod.Wrap(err, "set lp")
		}
		return nil
	}
	return errorsmod.Wrap(gerrc.ErrNotFound, "no compatible lp")
}

func (k Keeper) CreateLP(ctx sdk.Context, lp *types.OnDemandLP) (uint64, error) {
	return k.LPs.Create(ctx, lp)
}

func (k Keeper) DeleteLP(ctx sdk.Context, owner sdk.AccAddress, id uint64, reason string) error {
	lp, err := k.LPs.Get(ctx, id)
	if errors.Is(err, collections.ErrNotFound) {
		return nil
	}
	if err != nil {
		return errorsmod.Wrap(err, "get")
	}
	if !lp.Lp.MustAddr().Equals(owner) {
		return errorsmod.Wrapf(gerrc.ErrPermissionDenied, "not owner: require %s, got %s", lp.Lp.FundsAddr, owner)
	}
	return k.LPs.Del(ctx, id, reason)
}
