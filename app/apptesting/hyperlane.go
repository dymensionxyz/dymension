package apptesting

import (
	"cosmossdk.io/math"

	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uhyp"
)

func (s *KeeperTestHelper) SetupHyperlane() {

	server := uhyp.NewServer(&s.App.HyperCoreKeeper, &s.App.HyperCoreKeeper.PostDispatchKeeper, &s.App.HyperCoreKeeper.IsmKeeper, s.App.HyperWarpKeeper)
	owner := Alice
	denom := "acoin"
	largeAmt := math.NewInt(1e18)
	arbitraryContract := "0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0"
	transferAmt := math.NewInt(1000000)

	FundAccount(s.App, s.Ctx, sdk.MustAccAddressFromBech32(owner), sdk.NewCoins(sdk.NewCoin(denom, largeAmt)))

	mailboxId, err := server.CreateDefaultMailbox(s.Ctx, owner, denom)
	s.NoError(err)
	_ = mailboxId

	tokenID, err := server.CreateCollateralToken(s.Ctx, owner, mailboxId, denom)
	s.NoError(err)

	err = server.EnrollRemoteRouter(s.Ctx, owner, tokenID, warptypes.RemoteRouter{
		ReceiverDomain:   uhyp.RemoteDomain,
		ReceiverContract: arbitraryContract,
		Gas:              math.NewInt(0),
	})
	s.NoError(err)

	var maxFee sdk.Coin = sdk.NewCoin(denom, math.NewInt(1000000))
	var amt math.Int = transferAmt
	var gasLimit math.Int = math.NewInt(0)

	var customHookMetadata string      // ??
	var recipient hyperutil.HexAddress // in practice should be real counterparty recipient

	// TODO: to make this pass need to add some tokens for alice
	messageID, err := server.RemoteTransfer(s.Ctx, owner, tokenID, uhyp.RemoteDomain,
		recipient, nil, // custom hook is nil so use mailbox default hook as the second hook that occurs
		maxFee, customHookMetadata, amt, gasLimit)
	s.NoError(err)
	_ = messageID

}
