package types

import (
	proto "github.com/cosmos/gogoproto/proto"
)

func UnpackHLMetadata(metadata []byte) (*HLMetadata, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	var x HLMetadata
	err := proto.Unmarshal(metadata, &x)
	if err != nil {
		return nil, err
	}
	return &x, nil
}
