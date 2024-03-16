package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewStateInfo(rollappId string, newIndex uint64, creator string, startHeight uint64, numBlocks uint64, daPath string, version uint64, height uint64, BDs BlockDescriptors) *StateInfo {
	stateInfoIndex := StateInfoIndex{RollappId: rollappId, Index: newIndex}
	status := STATE_STATUS_RECEIVED
	return &StateInfo{
		StateInfoIndex: stateInfoIndex,
		Sequencer:      creator,
		StartHeight:    startHeight,
		NumBlocks:      numBlocks,
		DAPath:         daPath,
		Version:        version,
		CreationHeight: height,
		Status:         status,
		BDs:            BDs,
	}
}

func (s *StateInfo) Finalize() {
	s.Status = STATE_STATUS_FINALIZED
}

func (s *StateInfo) GetIndex() StateInfoIndex {
	return s.StateInfoIndex
}

func (s *StateInfo) GetEvents() []sdk.Attribute {
	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(AttributeKeyRollappId, s.StateInfoIndex.RollappId),
		sdk.NewAttribute(AttributeKeyStateInfoIndex, strconv.FormatUint(s.StateInfoIndex.Index, 10)),
		sdk.NewAttribute(AttributeKeyStartHeight, strconv.FormatUint(s.StartHeight, 10)),
		sdk.NewAttribute(AttributeKeyNumBlocks, strconv.FormatUint(s.NumBlocks, 10)),
		sdk.NewAttribute(AttributeKeyDAPath, s.DAPath),
		sdk.NewAttribute(AttributeKeyStatus, s.Status.String()),
	}
	return eventAttributes
}
