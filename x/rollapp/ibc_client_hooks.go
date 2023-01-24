package rollapp

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dymensionxyz/dymension/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/x/rollapp/types"

	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	ibcdmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/01-dymint/types"
)

var _ exported.ClientHooks = (*RollappClientHooks)(nil)

type RollappClientHooks struct {
	k *keeper.Keeper
}

// NewClientState creates a new ClientState instance
func NewRollappClientHooks(k *keeper.Keeper) exported.ClientHooks {
	return &RollappClientHooks{
		k: k,
	}
}

func (ch RollappClientHooks) OnCreateClient(
	ctx sdk.Context,
	clientState exported.ClientState,
	consensusState exported.ConsensusState,
) error {
	// filter only rollapp chains
	chainID := clientState.GetChainID()
	if isDymint, err := ch.isRollappChain(ctx, clientState.ClientType(), chainID); !isDymint || err != nil {
		return err
	}
	// get application stateRoot
	stateRoot := consensusState.GetRoot().GetHash()
	// get height
	height := clientState.GetLatestHeight().GetRevisionHeight()
	// check stateRoot validity
	return ch.validateStateRoot(ctx, chainID, height, stateRoot)
}

func (ch RollappClientHooks) OnUpdateClient(
	ctx sdk.Context,
	clientID string,
	header exported.Header,
) error {
	// filter only rollapp chains
	chainID := header.GetChainID()
	if isDymint, err := ch.isRollappChain(ctx, header.ClientType(), chainID); !isDymint || err != nil {
		return err
	}
	dymHeader := header.(*ibcdmtypes.Header)
	// get application stateRoot
	stateRoot := dymHeader.Header.GetAppHash()
	// get height
	height := dymHeader.Header.Height
	// check stateRoot validity
	return ch.validateStateRoot(ctx, chainID, uint64(height), stateRoot)
}

func (ch RollappClientHooks) OnUpgradeClient(
	ctx sdk.Context,
	clientID string,
	upgradedClient exported.ClientState,
	upgradedConsState exported.ConsensusState,
	proofUpgradeClient,
	proofUpgradeConsState []byte,
) error {
	// filter only rollapp chains
	chainID := upgradedClient.GetChainID()
	if isDymint, err := ch.isRollappChain(ctx, upgradedClient.ClientType(), chainID); !isDymint || err != nil {
		return err
	}
	// get application stateRoot
	stateRoot := upgradedConsState.GetRoot().GetHash()
	// get height
	height := upgradedClient.GetLatestHeight().GetRevisionHeight()
	// check stateRoot validity
	return ch.validateStateRoot(ctx, chainID, height, stateRoot)
}

func (ch RollappClientHooks) OnCheckMisbehaviourAndUpdateState(
	ctx sdk.Context,
	misbehaviour exported.Misbehaviour,
) error {
	// filter only rollapp chains
	chainID := misbehaviour.GetChainID()
	if isDymint, err := ch.isRollappChain(ctx, misbehaviour.ClientType(), chainID); !isDymint || err != nil {
		return err
	}

	dymHeader1 := misbehaviour.(*ibcdmtypes.Misbehaviour).Header1
	dymHeader2 := misbehaviour.(*ibcdmtypes.Misbehaviour).Header2
	// get application stateRoot
	stateRoot1 := dymHeader1.Header.GetAppHash()
	stateRoot2 := dymHeader2.Header.GetAppHash()
	// get height
	height1 := dymHeader1.Header.Height
	height2 := dymHeader2.Header.Height

	// check stateRoot validity
	if err := ch.validateStateRoot(ctx, chainID, uint64(height1), stateRoot1); err != nil {
		return err
	}
	return ch.validateStateRoot(ctx, chainID, uint64(height2), stateRoot2)
}

// isRollappChain checks that the clientType is Dymint
// and that the rollapp exists
func (ch RollappClientHooks) isRollappChain(
	ctx sdk.Context,
	clientType string,
	chainID string,
) (bool, error) {
	// rollapp client type is Dymint
	isDymint := clientType == exported.Dymint
	// rollappId is the rollapp chainId
	_, isFound := ch.k.GetRollapp(ctx, chainID)
	// if the client type isn't dymint and there is no such rollapp
	// we can be sure that the chain isn't a rollapp
	if !isDymint && !isFound {
		return false, nil
	}
	// client type is dymint and we know this rollapp
	if isDymint && isFound {
		return true, nil
	}
	// client type is dymint but no such rollapp
	if isDymint && !isFound {
		return true, sdkerrors.Wrapf(types.ErrUnknownRollappID, "rollappID: %s", chainID)
	}
	// client type is not dymint but the chain is a rollapp
	return false, sdkerrors.Wrapf(types.ErrInvalidClientType, "clientType: %s", clientType)
}

// validateStateRoot load the stateInfo and verify the state was finalized
// and that the stateRoot is matching to the one which stored
func (ch RollappClientHooks) validateStateRoot(ctx sdk.Context, rollappId string, height uint64, stateRoot []byte) error {
	stateInfo, err := ch.getStateInfo(ctx, rollappId, height)
	if err != nil {
		return err
	}
	if stateInfo.GetStatus() != types.STATE_STATUS_FINALIZED {
		return sdkerrors.Wrapf(types.ErrHeightStateNotFainalized, "rollappID: %s, height=%d", rollappId, height)
	}
	BlockDescriptionIndex := int(height - stateInfo.StartHeight)
	if BlockDescriptionIndex < 0 || BlockDescriptionIndex >= len(stateInfo.BDs.BD) {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic,
			"searching height=%d, found stateInfo.StartHeight=%d, stateInfo.NumBlocks=%d, len(stateInfo.BDs.BD)=%d",
			height, stateInfo.StartHeight, stateInfo.NumBlocks, len(stateInfo.BDs.BD))
	}
	blockDescription := stateInfo.BDs.BD[BlockDescriptionIndex]
	if blockDescription.Height != height {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic,
			"searching height=%d, found stateInfo.StartHeight=%d, stateInfo.NumBlocks=%d, len(stateInfo.BDs.BD)=%d, but BD[%d].Height=%d",
			height, stateInfo.StartHeight, stateInfo.NumBlocks, len(stateInfo.BDs.BD), BlockDescriptionIndex, blockDescription.Height)
	}
	if !bytes.Equal(stateRoot, blockDescription.StateRoot) {
		return sdkerrors.Wrapf(types.ErrInvalidAppHash, "rollappID: %s, appHash=%d", rollappId, blockDescription.StateRoot)
	}
	return nil
}

func (ch RollappClientHooks) getStateInfo(ctx sdk.Context, rollappId string, height uint64) (*types.StateInfo, error) {
	queryGetStateInfoByHeightResponse, err := ch.k.GetStateInfoByHeight(sdk.WrapSDKContext(ctx), &types.QueryGetStateInfoByHeightRequest{
		RollappId: rollappId,
		Height:    height,
	})
	if queryGetStateInfoByHeightResponse != nil {
		return &queryGetStateInfoByHeightResponse.StateInfo, err
	}
	return nil, err
}
