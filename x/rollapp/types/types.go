package types

import (
	"encoding/json"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
)

type StateStatus common.Status

func (md *RollappMetadata) IsEmpty() bool {
	bz, _ := json.Marshal(md)
	return string(bz) == "{}"
}

type DRSVersion string
