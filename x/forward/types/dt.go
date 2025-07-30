package types

import (
	errorsmod "cosmossdk.io/errors"
	proto "github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// ValidateBasic validates that at most one forward type is populated (excluding kaspa which is orthogonal)
func (m *HLMetadata) ValidateBasic() error {
	populatedCount := 0
	
	if len(m.HookForwardToIbc) > 0 {
		populatedCount++
	}
	
	if len(m.HookForwardToHl) > 0 {
		populatedCount++
	}
	
	// Note: kaspa field is orthogonal and not counted
	
	if populatedCount > 1 {
		return gerrc.ErrInvalidArgument.Wrap("at most one forward type can be populated in HLMetadata")
	}
	
	return nil
}

func UnpackHLMetadata(metadata []byte) (*HLMetadata, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	var x HLMetadata
	err := proto.Unmarshal(metadata, &x)
	if err != nil {
		return nil, err
	}
	
	// Validate the metadata
	if err := x.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate HLMetadata")
	}
	
	return &x, nil
}
