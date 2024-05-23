package types

import (
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
)

type StateStatus common.Status

type TriggerGenesisArgs struct {
	ChannelID string
	RollappID string
}
