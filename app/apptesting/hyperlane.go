package apptesting

import (
	"cosmossdk.io/math"

	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uhyp"
)

/*
A reminder on how things work

Mailbox has
	id
	owner
	sent and received cnt
	ref to default ism
	ref to default hook
	ref to required hook
	ref to local domain

Creating a token (warp route)
	 specify owner
	 origin mailbox
	 origin denom

Sending a transfer:
	specify
		token id
		sender
		destination domain
		recipient
		amt
		custom hook id
		gas limit
		max fee
		custom hook metadata(??)
	it finds a router contract for token id and destination domain (??)
	it dispatches a message with
		origin mailbox
		token id (=sender)(?)
		max fee
		router receiver domain
		router receiver contract
		warp route recipient and amount gets bundled into opaque payload
		given gas limit, sender account
		forward the custom hook metadata
		uses custom hook id
	dispatch logic:
		add a message to origin mailbox
		calls post dispatch:
			give actual message <version, nonce, local domain, sender, dst domain, recipient, body>
			use origin mailbox id
			use required hook
			pass custom hook metadata
		use post dispatch router for first the required hook then the default hook
		IN CASE OF TEST: REQUIRED HOOK IS IGP, DEFAULT HOOK IS NO OP

Receiving a transfer:
	Relayer does processMessage
		args are mailbox, raw message and metadata


*/

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
	var amt math.Int = math.NewInt(1000000)
	var gasLimit math.Int

	// TODO: to make this pass need to add some tokens for alice
	messageID, err := server.RemoteTransfer(s.Ctx, owner, tokenID, dstDomain,
		recipient, customHookId, maxFee, customHookMetadata, amt, gasLimit)
	s.NoError(err)
	_ = messageID

}
