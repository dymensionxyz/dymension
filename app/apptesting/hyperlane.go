package apptesting

import (
	hyperlanecoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
)

func (s *KeeperTestHelper) SetupHyperlane() string {

	m := hyperlanecoretypes.CreateMailbox{
		Owner:      s.Ctx.AccAddress().String(),
		DefaultIsm: defaultIsm,
	}

	return routerId
}
