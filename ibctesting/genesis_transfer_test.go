package ibctesting_test

import (
	"testing"

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
}

func TestTransferGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(transferGenesisSuite))
}

func (s *transferGenesisSuite) SetupTest() {
	s.utilSuite.SetupTest()
}

// In the happy path, the new rollapp will send ibc transfers with a special
// memo immediately when the channel opens. This will cause  all the denoms to get registered, and tokens
// to go to the right addresses.
func (s *transferGenesisSuite) TestHappyPath() {
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	s.createRollapp(false, nil) // genesis protocol is not finished yet
	s.registerSequencer()

	/*
		Send a bunch of transfer packets to the hub
		Check the balances are created
		Check the denoms are created
		Check the bridge is enabled (or not)
	*/

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amt, ok := sdk.NewIntFromString("10000000000000000000") // 10
	s.Require().True(ok)

	denoms := []string{"foo", "bar", "baz"}

	for i, denom := range denoms {
		/* ------------------- move non-registered token from rollapp ------------------- */
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
			Display: denom,
			Name:    denom,
			Symbol:  denom,
			URI:     "",
			URIHash: "",
		}
		s.Require().NoError(meta.Validate()) // sanity check the test is written correctly
		memo := rollapptypes.GenesisTransferMemo{
			Denom:             meta,
			TotalNumTransfers: uint64(len(denoms)),
			ThisTransferIx:    uint64(i),
		}.Namespaced().MustString()
		tokens := sdk.NewCoin(meta.Base, amt)

		apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.Coins{tokens})
		msg := types.NewMsgTransfer(
			path.EndpointB.ChannelConfig.PortID,
			path.EndpointB.ChannelID,
			tokens,
			s.rollappChain().SenderAccount.GetAddress().String(),
			s.hubChain().SenderAccount.GetAddress().String(),
			timeoutHeight,
			0,
			memo,
		)
		res, err := s.rollappChain().SendMsgs(msg)
		s.Require().NoError(err)
		packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
		s.Require().NoError(err)

		err = path.RelayPacket(packet)
		s.Require().NoError(err)

		transfersEnabled := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
		s.Require().Equal(i == len(denoms)-1, transfersEnabled, "transfers enabled check", "i", i)
	}

	for _, denom := range denoms {
		// has the denom?
		ibcDenom := types.ParseDenomTrace(types.GetPrefixedDenom(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, denom)).IBCDenom()
		metadata, found := s.hubApp().BankKeeper.GetDenomMetaData(s.hubCtx(), ibcDenom)
		s.Require().True(found, "missing denom metadata for rollapps taking token")
		s.Require().Equal(denom, metadata.Base)
		// has the tokens?
		c := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), ibcDenom)
		s.Require().Equal(amt, c.Amount)
	}
}
