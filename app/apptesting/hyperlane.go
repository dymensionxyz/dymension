package apptesting

import (
	"cosmossdk.io/math"

	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uhyp"
)

/*
Synthesis
	Setup

	Send (assume Dym from Hub to Ethereum)
		WR:
			Pass token, real recipient, dst, amt
			Need a router: mapping <token, dst, dst contract (warp route)>
			Dispatches a message with src mailbox as origin (token.OriginMailbox)
			Target is the dst contract
		Dispatch:
			Need required and default hooks
			Does post dispatch with the required hook


	Questions:
		Necessity of IGP?
		What is remote router gas?



*/

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

Modules
	A collateral warp route is an example of a module.

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
			mock:
				no op
		 	igp:
				pay for gas
			merkle tree:
				adds message ID to a merkle tree corresponding to the mailbox
			noop:
				no op

		IN CASE OF TEST: REQUIRED HOOK IS IGP, DEFAULT HOOK IS NO OP

Receiving a transfer:
	Relayer does processMessage
		args are mailbox, raw message and metadata
		mailbox local domain must be message destination
		uses approuter to get module and then ISM ID for the recipient contract
			(this calls into warp route and returns either the default ism of the token origin mailbox, or the token ISM if there is one) (for overriding?)
		uses ism router to get ism id
		metadata and message is passed to ISM:
			noop ISM:
				says it's valid
			merkle root multisig ISM:
				?
			message id multisig ISM:
				?
		after verification, handle it
		handling is via app router


Setting up the router?
	A warp route keeper has EnrolledRouters
		(token id, dst domain -> router)
		used for outbound message
		they have to be set explicitly for each token
		it has contract addresses for the receiving contract on the other side
		it has a gas field which works as default gas limit if user does not specify in their send

	A core keeper has routers:
		appRouter
			(recipient 	-> (recipient -> ism ID)
			used to find ism ID for inbound message
		ismRouter
			(id -> verifier)
			used to get ISM verifier for inbound message
		postDispatchRouter
			hook id -> f
			where f
			used after commiting a message to a mailbox


Setting up the ISM? Hooks?
*/

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

	var customHookMetadata string         // ??
	var recipient hyperutil.HexAddress    // should be real counterparty recipient
	var customHookId hyperutil.HexAddress // empty means fall back to default hook

	var maxFee sdk.Coin = sdk.NewCoin(denom, math.NewInt(1000000))
	var amt math.Int = transferAmt
	var gasLimit math.Int = math.NewInt(0)

	err = server.EnrollRemoteRouter(s.Ctx, owner, tokenID, warptypes.RemoteRouter{
		ReceiverDomain:   uhyp.RemoteDomain,
		ReceiverContract: arbitraryContract,
		Gas:              math.NewInt(0),
	})
	s.NoError(err)

	// TODO: to make this pass need to add some tokens for alice
	messageID, err := server.RemoteTransfer(s.Ctx, owner, tokenID, uhyp.RemoteDomain,
		recipient, customHookId, maxFee, customHookMetadata, amt, gasLimit)
	s.NoError(err)
	_ = messageID

}
