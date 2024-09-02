package ibctesting_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/transfergenesis"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
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
	s.coordinator.SetupConnections(path)
	s.createRollapp(false, nil) // genesis protocol is not finished yet
	s.registerSequencer()
	s.path = path
	// set the canonical client before creating channels
	s.hubApp().LightClientKeeper.SetCanonicalClient(s.hubCtx(), rollappChainID(), s.path.EndpointA.ClientID)
	s.coordinator.CreateChannels(path)

	// set hooks to avoid actually creating VFC contract, as this places extra requirements on the test setup
	// we assume that if the denom metadata was created (checked below), then the hooks ran correctly
	s.hubApp().DenomMetadataKeeper.SetHooks(nil)
}

// TestNoIRO tests the case where the rollapp has no IRO plan.
// In this case, the genesis transfer should fail, but regular transfers should succeed.
func (s *transferGenesisSuite) TestNoIRO() {
	amt := math.NewIntFromUint64(10000000000000000000)
	denom := "foo"
	coin := sdk.NewCoin(denom, amt)
	apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.NewCoins(coin))

	// no iro plan, so no genesis transfers allowed
	msg := s.transferMsg(amt, denom, true)
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()
	ack, found := hubIBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().NotEqual(channeltypes.CommitAcknowledgement(nil), ack)

	transfersEnabled := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
	s.Require().False(transfersEnabled)

	// regular transfer, should pass (with delay) and enable bridge
	msg = s.transferMsg(amt, denom, false)
	res, err = s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	err = s.path.RelayPacket(packet)
	s.Require().Error(err) // ack is delayed, so error is returned from the framework

	transfersEnabled = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
	s.Require().True(transfersEnabled)
}

// TestIRO tests the case where the rollapp has an IRO plan.
// In this case, the genesis transfer is required
// regular transfers should fail until the genesis transfer is done
func (s *transferGenesisSuite) TestIRO() {
	amt := math.NewIntFromUint64(10000000000000000000)
	denom := "foo"
	coin := sdk.NewCoin(denom, amt)
	apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.NewCoins(coin))

	// create IRO plan
	rollapp := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	_, err := s.hubApp().IROKeeper.CreatePlan(s.hubCtx(), amt, time.Time{}, time.Time{}, rollapp, irotypes.DefaultBondingCurve())
	s.Require().NoError(err)

	// non-genesis transfer should fail, as the bridge is not open
	msg := s.transferMsg(amt, denom, false)
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	ack, found := s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().NotEqual(channeltypes.CommitAcknowledgement(nil), ack) // check for error ack

	transfersEnabled := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
	s.Require().False(transfersEnabled)

	// test invalid genesis transfers
	// - TODO: wrong dest
	// - wrong amount
	msg = s.transferMsg(amt.Sub(math.NewInt(100)), denom, true)
	res, err = s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	ack, found = s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().NotEqual(channeltypes.CommitAcknowledgement(nil), ack) // check for error ack

	transfersEnabled = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
	s.Require().False(transfersEnabled)

	// genesis transfer, should pass and enable bridge
	msg = s.transferMsg(amt, denom, true)
	res, err = s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	transfersEnabled = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
	s.Require().True(transfersEnabled, "transfers enabled check")

	// has the denom?
	ibcDenom := types.ParseDenomTrace(types.GetPrefixedDenom(s.path.EndpointB.ChannelConfig.PortID, s.path.EndpointB.ChannelID, denom)).IBCDenom()
	metadata, found := s.hubApp().BankKeeper.GetDenomMetaData(s.hubCtx(), ibcDenom)
	s.Require().True(found, "missing denom metadata for rollapps taking token", "denom", ibcDenom)
	s.Require().Equal(ibcDenom, metadata.Base)
	// the iro plan should be settled
	plan, found := s.hubApp().IROKeeper.GetPlanByRollapp(s.hubCtx(), rollappChainID())
	s.Require().True(found)
	s.Require().Equal(plan.SettledDenom, ibcDenom)
}

// In the fault path, a chain tries to do another genesis transfer (to skip eibc) after the genesis phase
// is already complete. It triggers a fraud.
func (s *transferGenesisSuite) TestCannotDoGenesisTransferAfterBridgeEnabled() {
	amt := math.NewIntFromUint64(10000000000000000000)
	denom := "foo"
	apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.NewCoins(sdk.NewCoin(denom, amt.Mul(math.NewInt(10)))))

	// non-genesis transfer should enable the bridge
	msg := s.transferMsg(amt, denom, false)
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().Error(err) // ack is delayed, so error is returned from the framework

	// genesis transfer after bridge enabled should fail
	msg = s.transferMsg(amt, denom, true)
	res, err = s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	expect := channeltypes.NewErrorAcknowledgement(transfergenesis.ErrDisabled)
	bz, _ := s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().Equal(channeltypes.CommitAcknowledgement(expect.Acknowledgement()), bz)
}

func (s *transferGenesisSuite) transferMsg(amt math.Int, denom string, isGenesis bool) *types.MsgTransfer {
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
		"",
	)

	if isGenesis {
		msg.Memo = rollapptypes.GenesisTransferMemo{
			Denom: meta,
		}.Namespaced().MustString()

		msg.Receiver = s.hubApp().IROKeeper.GetModuleAccountAddress()
	}

	return msg
}
