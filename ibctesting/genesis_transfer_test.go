package ibctesting_test

import (
	"testing"

	"cosmossdk.io/math"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
)

type transferGenesisSuite struct {
	utilSuite

	path *ibctesting.Path
}

func TestTransferGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(transferGenesisSuite))
}

func (s *transferGenesisSuite) SetupTest() {
	s.utilSuite.SetupTest()
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	s.createRollapp(false, nil) // genesis protocol is not finished yet
	s.registerSequencer()
	s.path = path

	// set hooks to avoid actually creating VFC contract, as this places extra requirements on the test setup
	// we assume that if the denom metadata was created (checked below), then the hooks ran correctly
	s.hubApp().DenomMetadataKeeper.SetHooks(nil)
}

// In the happy path, the new rollapp will send ibc transfers with a special
// memo immediately when the channel opens. This will cause  all the denoms to get registered, and tokens
// to go to the right addresses. After all transfers are sent, the bridge opens.
func (s *transferGenesisSuite) TestHappyPath() {
	/*
		Send a bunch of transfer packets to the hub
		Check the balances are created
		Check the denoms are created
		Check the bridge is enabled (or not)
	*/

	amt := math.NewIntFromUint64(10000000000000000000)

	denoms := []string{"foo", "bar", "baz"}

	for i, denom := range denoms {
		/* ------------------- move non-registered token from rollapp ------------------- */

		msg := s.transferMsg(amt, denom, i, len(denoms))
		apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.Coins{msg.Token})
		res, err := s.rollappChain().SendMsgs(msg)
		s.Require().NoError(err)
		packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
		s.Require().NoError(err)

		err = s.path.RelayPacket(packet)
		s.Require().NoError(err)
		// after the last one, it should be OK
		transfersEnabled := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
		s.Require().Equal(i == len(denoms)-1, transfersEnabled, "transfers enabled check", "i", i)
	}

	for _, denom := range denoms {
		// has the denom?
		ibcDenom := types.ParseDenomTrace(types.GetPrefixedDenom(s.path.EndpointB.ChannelConfig.PortID, s.path.EndpointB.ChannelID, denom)).IBCDenom()
		metadata, found := s.hubApp().BankKeeper.GetDenomMetaData(s.hubCtx(), ibcDenom)
		s.Require().True(found, "missing denom metadata for rollapps taking token")
		s.Require().Equal(ibcDenom, metadata.Base)
		// has the tokens?
		c := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), ibcDenom)
		s.Require().Equal(amt, c.Amount)
	}
}

func (s *transferGenesisSuite) transferMsg(amt math.Int, denom string, i, nDenomsTotal int) *types.MsgTransfer {
	meta := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom: denom,
			},
			{
				Denom:    "a" + denom,
				Exponent: 18,
			},
		},
		Base:    denom,
		Display: "a" + denom, // TODO: set right
		Name:    denom,
		Symbol:  denom,
		URI:     "",
		URIHash: "",
	}
	s.Require().NoError(meta.Validate()) // sanity check the test is written correctly
	memo := rollapptypes.GenesisTransferMemo{
		Denom:             meta,
		TotalNumTransfers: uint64(nDenomsTotal),
		ThisTransferIx:    uint64(i),
	}.Namespaced().MustString()
	tokens := sdk.NewCoin(meta.Base, amt)

	timeoutHeight := clienttypes.NewHeight(100, 110)

	msg := types.NewMsgTransfer(
		s.path.EndpointB.ChannelConfig.PortID,
		s.path.EndpointB.ChannelID,
		tokens,
		s.rollappChain().SenderAccount.GetAddress().String(),
		s.hubChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		memo,
	)

	return msg
}
