package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        storetypes.StoreKey
	ibcClientKeeper types.IBCClientKeeperExpected
	sequencerKeeper types.SequencerKeeperExpected
	rollappKeeper   types.RollappKeeperExpected

	// <sequencer addr,client ID, height>
	headerSigners collections.KeySet[collections.Triple[string, string, uint64]]
	// <client ID, height> -> <sequencer addr>
	clientHeightToSigner collections.Map[collections.Pair[string, uint64], string]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ibcKeeper types.IBCClientKeeperExpected,
	sequencerKeeper types.SequencerKeeperExpected,
	rollappKeeper types.RollappKeeperExpected,
) *Keeper {
	k := &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		ibcClientKeeper: ibcKeeper,
		sequencerKeeper: sequencerKeeper,
		rollappKeeper:   rollappKeeper,
	}
	return k
}

func (k Keeper) CanUnbond(ctx sdk.Context, seq sequencertypes.Sequencer) error {
	client, ok := k.GetCanonicalClient(ctx, seq.RollappId)
	if !ok {
		return errorsmod.Wrap(sequencertypes.ErrUnbondNotAllowed, "no canonical client")
	}
	rng := collections.NewSuperPrefixedTripleRange[string, string, uint64](seq.Address, client)
	return k.headerSigners.Walk(ctx, rng, func(key collections.Triple[string, string, uint64]) (stop bool, err error) {
		return true, errorsmod.Wrapf(sequencertypes.ErrUnbondNotAllowed, "unverified header: h: %d", key.K3())
	})
}

// PruneSigners removes bookkeeping for all heights ABOVE h for given rollapp
// This should only be called after canonical client set
// TODO: plug into hard fork
func (k Keeper) PruneSigners(ctx sdk.Context, rollapp string, h uint64) error {
	client, ok := k.GetCanonicalClient(ctx, rollapp)
	if !ok {
		return gerrc.ErrInternal.Wrap(`
prune light client signers for rollapp before canonical client is set
this suggests fork happened prior to genesis bridge completion, which
shouldnt be allowed
`)
	}
	rng := collections.NewPrefixedPairRange[string, uint64](client).StartExclusive(h)

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
		k.headerSigners.Set(ctx, collections.Join3(seqAddr, client, h)),
		k.clientHeightToSigner.Set(ctx, collections.Join(client, h), seqAddr),
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

func (k Keeper) ExpectedClientState(goCtx context.Context, req *types.QueryExpectedClientStateRequest) (*types.QueryExpectedClientStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	c := k.expectedClient(ctx)
	anyClient, err := ibcclienttypes.PackClientState(&c)
	if err != nil {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrInternal, err), "pack client state")
	}
	return &types.QueryExpectedClientStateResponse{ClientState: anyClient}, nil
}
