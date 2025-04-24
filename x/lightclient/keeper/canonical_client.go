package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var (
	ErrNoMatch        = gerrc.ErrFailedPrecondition.Wrap("not at least one cons state matches the rollapp state")
	ErrMismatch       = gerrc.ErrInvalidArgument.Wrap("consensus state mismatch")
	ErrParamsMismatch = gerrc.ErrInvalidArgument.Wrap("params")
)

// intended to be called by relayer, but can be called by anyone
// verifies that the suggested client is safe to designate canonical and matches state updates from the sequencer
func (k *Keeper) TrySetCanonicalClient(ctx sdk.Context, clientID string) error {
	clientStateI, ok := k.ibcClientKeeper.GetClientState(ctx, clientID)
	if !ok {
		return gerrc.ErrNotFound.Wrap("client")
	}

	clientState, ok := clientStateI.(*ibctm.ClientState)
	if !ok {
		return gerrc.ErrInvalidArgument.Wrap("not tm client")
	}

	chainID := clientState.ChainId
	_, ok = k.rollappKeeper.GetRollapp(ctx, chainID)
	if !ok {
		return gerrc.ErrNotFound.Wrap("rollapp")
	}
	rollappID := chainID

	_, ok = k.GetCanonicalClient(ctx, rollappID)
	if ok {
		return gerrc.ErrAlreadyExists.Wrap("canonical client for rollapp")
	}

	err := k.validClient(ctx, clientID, clientState, rollappID)
	if err != nil {
		return errorsmod.Wrapf(err, "set client canonical")
	}

	// Check if the clientID has any connections
	_, found := k.ibcConnectionK.GetClientConnectionPaths(ctx, clientID)
	if found {
		return gerrc.ErrFailedPrecondition.Wrap("client already has connections")
	}

	k.SetCanonicalClient(ctx, rollappID, clientID)

	if err := uevent.EmitTypedEvent(ctx, &types.EventSetCanonicalClient{
		RollappId: rollappID,
		ClientId:  clientID,
	}); err != nil {
		return errorsmod.Wrap(err, "emit typed event")
	}

	return nil
}

func (k Keeper) GetCanonicalClient(ctx sdk.Context, rollappId string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRollappClientKey(rollappId))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) SetCanonicalClient(ctx sdk.Context, rollappId string, clientID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRollappClientKey(rollappId), []byte(clientID))
	store.Set(types.CanonicalClientKey(clientID), []byte(rollappId))
}

func (k Keeper) GetAllCanonicalClients(ctx sdk.Context) (clients []types.CanonicalClient) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.RollappClientKey)
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		clients = append(clients, types.CanonicalClient{
			RollappId:   string(iterator.Key()[1:]),
			IbcClientId: string(iterator.Value()),
		})
	}
	return
}

func (k Keeper) expectedClient() ibctm.ClientState {
	return types.DefaultExpectedCanonicalClientParams()
}

// The canonical client criteria are:
// 1. The client must be a tendermint client.
// 2. The client state must match the expected client params as configured by the module
// 3. All the existing consensus states much match the corresponding height rollapp block descriptors
func (k Keeper) validClient(ctx sdk.Context, clientID string, cs *ibctm.ClientState, rollappId string) error {
	expClient := k.expectedClient()
	if err := types.IsCanonicalClientParamsValid(cs, &expClient); err != nil {
		return errors.Join(err, ErrParamsMismatch)
	}

	sinfo, ok := k.rollappKeeper.GetLatestStateInfoIndex(ctx, rollappId)
	if !ok {
		return gerrc.ErrNotFound.Wrap("latest state info index")
	}

	baseHeight := k.GetFirstConsensusStateHeight(ctx, clientID)
	atLeastOneMatch := false
	for i := sinfo.Index; i > 0; i-- {
		sInfo, ok := k.rollappKeeper.GetStateInfo(ctx, rollappId, i)
		if !ok {
			return errorsmod.Wrap(gerrc.ErrInternal, "get state info")
		}
		matched, err := k.ValidateStateInfoAgainstConsensusStates(ctx, clientID, &sInfo)
		if err != nil {
			return errors.Join(ErrMismatch, err)
		}

		if matched {
			atLeastOneMatch = true
		}

		// break point when we validate the state info for the first height of the client
		if sInfo.StartHeight < baseHeight {
			break
		}
	}

	// Need to be sure that at least one consensus state agrees with a state update
	// (There are also no disagreeing consensus states. There may be some consensus states
	// for future state updates, which will incur a fraud if they disagree.)
	if !atLeastOneMatch {
		return ErrNoMatch
	}
	return nil
}

func (k Keeper) ValidateHeaderAgainstStateInfo(ctx sdk.Context, sInfo *rollapptypes.StateInfo, consState *ibctm.ConsensusState, h uint64) error {
	bd, ok := sInfo.GetBlockDescriptor(h)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrInternal, "no block descriptor found for height %d", h)
	}

	nextSeq, err := k.SeqK.RealSequencer(ctx, sInfo.NextSequencerForHeight(h))
	if err != nil {
		return errorsmod.Wrap(errors.Join(err, gerrc.ErrInternal), "get sequencer of state info")
	}

	rollappState := types.RollappState{
		BlockDescriptor:    bd,
		NextBlockSequencer: nextSeq,
	}
	return errorsmod.Wrap(types.CheckCompatibility(*consState, rollappState), "check compatibility")
}
