package apptesting

import (
	"cosmossdk.io/math"

	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uhyp"
)

func (s *KeeperTestHelper) SetupHyperlane() {

	server := uhyp.NewServer(&s.App.HyperCoreKeeper, &s.App.HyperCoreKeeper.PostDispatchKeeper, &s.App.HyperCoreKeeper.IsmKeeper, s.App.HyperWarpKeeper)
	owner := Alice

	mailboxId, err := server.CreateDefaultMailbox(s.Ctx, owner, "acoin")
	s.NoError(err)
	_ = mailboxId

	tokenID, err := server.CreateCollateralToken(s.Ctx, owner, mailboxId, "acoin")
	s.NoError(err)

	dstDomain := uint32(1)
	var recipient hyperutil.HexAddress
	var customHookId hyperutil.HexAddress
	var maxFee sdk.Coin
	var customHookMetadata string
	var amt math.Int
	var gasLimit math.Int
	messageID, err := server.RemoteTransfer(s.Ctx, owner, tokenID, dstDomain,
		recipient, customHookId, maxFee, customHookMetadata, amt, gasLimit)
	s.NoError(err)
	_ = messageID

}
