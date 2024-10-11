package types

import (
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func NewStateInfo(
	rollappId string,
	newIndex uint64,
	creator string,
	startHeight uint64,
	numBlocks uint64,
	daPath string,
	height uint64,
	BDs BlockDescriptors,
	createdAt time.Time,
) *StateInfo {
	stateInfoIndex := StateInfoIndex{RollappId: rollappId, Index: newIndex}
	status := common.Status_PENDING
	return &StateInfo{
		StateInfoIndex: stateInfoIndex,
		Sequencer:      creator,
		StartHeight:    startHeight,
		NumBlocks:      numBlocks,
		DAPath:         daPath,
		CreationHeight: height,
		Status:         status,
		BDs:            BDs,
		CreatedAt:      createdAt,
	}
}

func (s *StateInfo) Finalize() {
	s.Status = common.Status_FINALIZED
}

func (s *StateInfo) GetIndex() StateInfoIndex {
	return s.StateInfoIndex
}

func (s *StateInfo) GetLatestHeight() uint64 {
	return s.StartHeight + s.NumBlocks - 1
}

func (s *StateInfo) ContainsHeight(height uint64) bool {
	return s.StartHeight <= height && height <= s.GetLatestHeight()
}

func (s *StateInfo) GetBlockDescriptor(height uint64) (BlockDescriptor, bool) {
	if !s.ContainsHeight(height) {
		return BlockDescriptor{}, false
	}
	return s.BDs.BD[height-s.StartHeight], true
}

func (s *StateInfo) GetLatestBlockDescriptor() BlockDescriptor {
	// return s.BDs.BD[s.NumBlocks-1] // todo: should it be this? or the one below? using this breaks ibctesting tests
	return s.BDs.BD[len(s.BDs.BD)-1]
}

func (s *StateInfo) GetEvents() []sdk.Attribute {
	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(AttributeKeyRollappId, s.StateInfoIndex.RollappId),
		sdk.NewAttribute(AttributeKeyStateInfoIndex, strconv.FormatUint(s.StateInfoIndex.Index, 10)),
		sdk.NewAttribute(AttributeKeyStartHeight, strconv.FormatUint(s.StartHeight, 10)),
		sdk.NewAttribute(AttributeKeyNumBlocks, strconv.FormatUint(s.NumBlocks, 10)),
		sdk.NewAttribute(AttributeKeyDAPath, s.DAPath),
		sdk.NewAttribute(AttributeKeyStatus, s.Status.String()),
		sdk.NewAttribute(AttributeKeyCreatedAt, s.CreatedAt.Format(time.RFC3339)),
	}
	return eventAttributes
}
