package apptesting

import (
	hyperlanecorekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	hyperlanecoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
)

/*
TODO: this is all a big wip
*/

func (s *KeeperTestHelper) SetupHyperlane() {

	coreServer := hyperlanecorekeeper.NewMsgServerImpl(&s.App.HyperCoreKeeper)
	var owner string
	localDomain := uint32(1)
	defaultIsm := hyperutil.NewZeroAddress()
	defaultHook := uptr.To(hyperutil.NewZeroAddress())
	defaultRequiredHook := uptr.To(hyperutil.NewZeroAddress())
	m := hyperlanecoretypes.MsgCreateMailbox{
		Owner:        owner,
		LocalDomain:  localDomain,
		DefaultIsm:   defaultIsm,
		DefaultHook:  defaultHook,
		RequiredHook: defaultRequiredHook,
	}

	_, err := coreServer.CreateMailbox(s.Ctx, &m)
	s.Require().NoError(err)

}
