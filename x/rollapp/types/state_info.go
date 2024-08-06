package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func NewStateInfo(rollappId string, newIndex uint64, creator string, startHeight uint64, numBlocks uint64, daPath string, height uint64, BDs BlockDescriptors) *StateInfo {
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

func (s *StateInfo) GetDAPathAsString() string {
	return s.DAPath
}

// GetDAPathAsDAPath returns the structured DAPath from the DAPath string
// This is used for v2.UpdateState to determine the type of the DA client
func (s *StateInfo) GetDAPathAsDAPath() (DAPath, error) {
	if s.DAPath == "" {
		return DAPath{}, nil
	}
	var dapath DAPath
	err := dapath.Unmarshal([]byte(s.DAPath))
	return dapath, err
}

// GetDAType returns the DA type from the DAPath string
func (s *StateInfo) GetDAType() string {
	daPath, err := s.GetDAPathAsDAPath()
	if err != nil {
		return "" // In case of older DAPath
	}
	return daPath.DaType
}
