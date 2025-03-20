package apptesting

import (
	hyperlanecorekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	hyperlanecoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	"github.com/dymensionxyz/dymension/v3/utils/uhyp"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	ismkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/keeper"
	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	pdkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/keeper"
	pdtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
)

/*
TODO: this is all a big wip
*/

func (s *KeeperTestHelper) SetupHyperlane() {

	coreServer := hyperlanecorekeeper.NewMsgServerImpl(&s.App.HyperCoreKeeper)
	ismServer := ismkeeper.NewMsgServerImpl(&s.App.HyperCoreKeeper.IsmKeeper)
	pdServer := pdkeeper.NewMsgServerImpl(&s.App.HyperCoreKeeper.PostDispatchKeeper)

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

	// 1) Create a NoopISM for the mailbox's default ISM

	ismRes, err := ismServer.CreateNoopIsm(s.Ctx, &ismtypes.MsgCreateNoopIsm{
		Creator: Alice,
	})
	s.Require().NoError(err)

	// 2) Create mailbox with that default ISM
	mbRes, err := coreServer.CreateMailbox(s.Ctx, &hyperlanecoretypes.MsgCreateMailbox{
		Owner:       Alice,
		DefaultIsm:  uhyp.MustDecodeHexAddress(ismRes.Id),
		LocalDomain: 42,
	})
	s.Require().NoError(err)

	// 3) Create a merkle tree hook, which we'll use as requiredHook

	mtRes, err := pdServer.CreateMerkleTreeHook(s.Ctx, &pdtypes.MsgCreateMerkleTreeHook{
		Owner:     Alice,
		MailboxId: mbRes.Id,
	})
	s.Require().NoError(err)

	// 4) Create a noop hook for the mailbox's defaultHook
	noopRes, err := pdServer.CreateNoopHook(s.Ctx, &pdtypes.MsgCreateNoopHook{
		Owner: Alice,
	})
	s.Require().NoError(err)

	// 5) Set the mailbox requiredHook and defaultHook
	_, err = coreServer.SetMailbox(s.Ctx, &hyperlanecoretypes.MsgSetMailbox{
		Owner:        Alice,
		MailboxId:    hyperutil.MustDecodeHexAddress(mbRes.Id),
		RequiredHook: hyperutil.MustDecodeHexAddress(mtRes.Id),
		DefaultHook:  hyperutil.MustDecodeHexAddress(noopRes.Id),
	})
	s.Require().NoError(err)
}
