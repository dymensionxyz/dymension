package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// wrapper to allow taking a pointer to mutable value
type enabled struct {
	enabled bool
}

type Keeper struct {
	// if false, will not run the msg update client ante handler. Very hacky
	// use to avoid problems in ibctesting.
	enabled *enabled

	cdc             codec.BinaryCodec
	storeKey        storetypes.Key
	ibcClientKeeper types.IBCClientKeeperExpected
	ibcConnectionK  types.IBCConnectionKeeperExpected
	ibcChannelK     types.IBCChannelKeeperExpected
	SeqK            types.SequencerKeeperExpected
	rollappKeeper   types.RollappKeeperExpected

	// <sequencer addr,client ID, height>
	headerSigners collections.KeySet[collections.Triple[string, string, uint64]]
	// <client ID, height> -> <sequencer addr>
	clientHeightToSigner collections.Map[collections.Pair[string, uint64], string]
}

func (k Keeper) Enabled() bool {
	return k.enabled.enabled
}

func (k Keeper) SetEnabled(b bool) {
	k.enabled.enabled = b
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.Key,
	ibcClientKeeper types.IBCClientKeeperExpected,
	ibcConnectionK types.IBCConnectionKeeperExpected,
	ibcChannelK types.IBCChannelKeeperExpected,
	sequencerKeeper types.SequencerKeeperExpected,
	rollappKeeper types.RollappKeeperExpected,
) *Keeper {
	service := collcompat.NewKVStoreService(storeKey)
	sb := collections.NewSchemaBuilder(service)
	k := &Keeper{
		enabled:         &enabled{true},
		cdc:             cdc,
		storeKey:        storeKey,
		ibcClientKeeper: ibcClientKeeper,
		ibcConnectionK:  ibcConnectionK,
		ibcChannelK:     ibcChannelK,
		SeqK:            sequencerKeeper,
		rollappKeeper:   rollappKeeper,
		headerSigners: collections.NewKeySet(
			sb,
			types.HeaderSignersPrefixKey,
			"header_signers",
			collections.TripleKeyCodec(collections.StringKey, collections.StringKey, collections.Uint64Key),
		),
		clientHeightToSigner: collections.NewMap(
			sb,
			types.ClientHeightToSigner,
			"client_height_to_signer",
			collections.PairKeyCodec(collections.StringKey, collections.Uint64Key),
			collections.StringValue,
		),
	}
	return k
}

func (k Keeper) CanUnbond(ctx sdk.Context, seq sequencertypes.Sequencer) error {
	client, ok := k.GetCanonicalClient(ctx, seq.RollappId)
	if !ok {
		// It doesn't make sense to prevent unbonding here. If there is no canonical client, then
		// there can have been no fraud that needs to be checked.
		// Moreover, if we did prevent unbonding, it would lead to awkward situations where non proposer
		// sequencers of early stage rollapps can't unbond, when the proposer is not doing his job.
		return nil
	}
	rng := collections.NewSuperPrefixedTripleRange[string, string, uint64](seq.Address, client)
	return k.headerSigners.Walk(ctx, rng, func(key collections.Triple[string, string, uint64]) (stop bool, err error) {
		return true, errorsmod.Wrapf(sequencertypes.ErrUnbondNotAllowed, "unverified header: h: %d", key.K3())
	})
}

// PruneSignersAbove removes bookkeeping for all heights ABOVE h for given client
// This should only be called after canonical client set
func (k Keeper) PruneSignersAbove(ctx sdk.Context, client string, h uint64) error {
	return k.pruneSigners(ctx, client, h, true)
}

// PruneSignersBelow removes bookkeeping for all heights BELOW h for given clientId
// This should only be called after canonical client set
func (k Keeper) PruneSignersBelow(ctx sdk.Context, client string, h uint64) error {
	return k.pruneSigners(ctx, client, h, false)
}

// GetSigner returns the sequencer address who signed the header in the update
func (k Keeper) GetSigner(ctx sdk.Context, client string, h uint64) (string, error) {
	return k.clientHeightToSigner.Get(ctx, collections.Join(client, h))
}

func (k Keeper) SaveSigner(ctx sdk.Context, seqAddr string, client string, h uint64) error {
	return errors.Join(
		k.headerSigners.Set(ctx, collections.Join3(seqAddr, client, h)),
		k.clientHeightToSigner.Set(ctx, collections.Join(client, h), seqAddr),
	)
}

func (k Keeper) RemoveSigner(ctx sdk.Context, seqAddr string, client string, h uint64) error {
	return errors.Join(
		k.headerSigners.Remove(ctx, collections.Join3(seqAddr, client, h)),
		k.clientHeightToSigner.Remove(ctx, collections.Join(client, h)),
	)
}

func (k Keeper) GetRollappForClientID(ctx sdk.Context, clientID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CanonicalClientKey(clientID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) LightClient(goCtx context.Context, req *types.QueryGetLightClientRequest) (*types.QueryGetLightClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	id, _ := k.GetCanonicalClient(ctx, req.GetRollappId()) // if not found then empty
	return &types.QueryGetLightClientResponse{ClientId: id}, nil
}

func (k Keeper) ExpectedClientState(context.Context, *types.QueryExpectedClientStateRequest) (*types.QueryExpectedClientStateResponse, error) {
	c := k.expectedClient()
	anyClient, err := ibcclienttypes.PackClientState(&c)
	if err != nil {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrInternal, err), "pack client state")
	}
	return &types.QueryExpectedClientStateResponse{ClientState: anyClient}, nil
}

// a convenience function to get both hub and rollapp channel ids from just the rollapp id
func (k Keeper) RollappCanonChannel(goCtx context.Context, req *types.QueryRollappCanonChannelRequest) (*types.QueryRollappCanonChannelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ra, ok := k.rollappKeeper.GetRollapp(ctx, req.GetRollappId())
	if !ok {
		return nil, rollapptypes.ErrRollappNotFound
	}
	if ra.ChannelId == "" {
		return nil, gerrc.ErrFailedPrecondition.Wrap("canonical channel not set on rollapp")
	}
	cha, ok := k.ibcChannelK.GetChannel(ctx, "transfer", ra.ChannelId)
	if !ok {
		return nil, gerrc.ErrInternal.Wrapf("channel: %s", ra.ChannelId)
	}
	return &types.QueryRollappCanonChannelResponse{
		HubChannelId:     ra.ChannelId,
		RollappChannelId: cha.GetCounterparty().GetChannelID(),
	}, nil
}

func (k Keeper) pruneSigners(ctx sdk.Context, client string, h uint64, isAbove bool) error {
	var rng *collections.PairRange[string, uint64]
	if isAbove {
		rng = collections.NewPrefixedPairRange[string, uint64](client).StartExclusive(h)
	} else {
		rng = collections.NewPrefixedPairRange[string, uint64](client).EndExclusive(h)
	}

	seqs := make([]string, 0)
	heights := make([]uint64, 0)

	// collect first to avoid del while iterating
	if err := k.clientHeightToSigner.Walk(ctx, rng, func(key collections.Pair[string, uint64], value string) (stop bool, err error) {
		seqs = append(seqs, value)
		heights = append(heights, key.K2())
		return false, nil
	}); err != nil {
		return errorsmod.Wrap(err, "walk signers")
	}

	for i := 0; i < len(seqs); i++ {
		if err := k.RemoveSigner(ctx, seqs[i], client, heights[i]); err != nil {
			return errorsmod.Wrap(err, "remove signer")
		}
	}
	return nil
}
