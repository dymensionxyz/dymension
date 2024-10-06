package ibctesting_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/genesisbridge"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
)

var successAck = channeltypes.CommitAcknowledgement(channeltypes.NewResultAcknowledgement([]byte{byte(1)}).Acknowledgement())

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

	// fund the rollapp owner account for iro creation fee
	iroFee := sdk.NewCoin(appparams.BaseDenom, s.hubApp().IROKeeper.GetParams(s.hubCtx()).CreationFee)
	apptesting.FundAccount(s.hubApp(), s.hubCtx(), s.hubChain().SenderAccount.GetAddress(), sdk.NewCoins(iroFee))

	// FIXME: remove?
	// fund the iro module account for pool creation fee
	poolFee := s.hubApp().GAMMKeeper.GetParams(s.hubCtx()).PoolCreationFee
	apptesting.FundAccount(s.hubApp(), s.hubCtx(), sdk.MustAccAddressFromBech32(s.hubApp().IROKeeper.GetModuleAccountAddress()), poolFee)

	// set the canonical client before creating channels
	s.path = path
	s.hubApp().LightClientKeeper.SetCanonicalClient(s.hubCtx(), rollappChainID(), s.path.EndpointA.ClientID)
	s.coordinator.CreateChannels(path)

	// set hooks to avoid actually creating VFC contract, as this places extra requirements on the test setup
	// we assume that if the denom metadata was created (checked below), then the hooks ran correctly
	s.hubApp().DenomMetadataKeeper.SetHooks(nil)
}

// TestHappyPath_NoGenesisAccounts tests a valid genesis info with no genesis accounts
func (s *transferGenesisSuite) TestHappyPath_NoGenesisAccounts() {
	rollapp := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())

	// create the expected genesis bridge packet
	packet := s.genesisBridgePacket(rollapp.GenesisInfo.NativeDenom.Base, rollapp.GenesisInfo.NativeDenom.Display, rollapp.GenesisInfo.InitialSupply, nil)

	// send the packet on the rollapp chain
	seq, err := s.path.EndpointB.SendPacket(packet.TimeoutHeight, packet.TimeoutTimestamp, packet.Data)
	s.Require().NoError(err)
	packet.Sequence = seq

	_, err = s.path.EndpointA.RecvPacketWithResult(packet)
	s.Require().NoError(err)

	// assert the ack succeeded
	ack, found := s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().Equal(successAck, ack)

	// assert the transfers are enabled
	rollapp = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	s.Require().True(rollapp.GenesisState.TransfersEnabled)
}

// TestHappyPath_GenesisAccounts tests a valid genesis info with genesis accounts
func (s *transferGenesisSuite) TestHappyPath_GenesisAccounts() {
	rollapp := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())

	gAccounts := []rollapptypes.GenesisAccount{
		{
			Address: s.rollappChain().SenderAccount.GetAddress().String(),
			Amount:  math.NewIntFromUint64(10000000000000000000),
		},
	}
	s.addGenesisAccounts(gAccounts)

	// create the expected genesis bridge packet
	packet := s.genesisBridgePacket(rollapp.GenesisInfo.NativeDenom.Base, rollapp.GenesisInfo.NativeDenom.Display, rollapp.GenesisInfo.InitialSupply, gAccounts)

	// send the packet on the rollapp chain
	seq, err := s.path.EndpointB.SendPacket(packet.TimeoutHeight, packet.TimeoutTimestamp, packet.Data)
	s.Require().NoError(err)
	packet.Sequence = seq

	_, err = s.path.EndpointA.RecvPacketWithResult(packet)
	s.Require().NoError(err)

	// assert the ack succeeded
	ack, found := s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().Equal(successAck, ack)

	// assert the transfers are enabled
	rollapp = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	s.Require().True(rollapp.GenesisState.TransfersEnabled)

	// FIXME: assert the genesis accounts were funded

}

// TestIRO tests the case where the rollapp has an IRO plan.
// In this case, the genesis transfer is required
// regular transfers should fail until the genesis transfer is done

// TestHappyPath_GenesisAccounts_IRO tests a valid genesis info with genesis accounts, including IRO plan
func (s *transferGenesisSuite) TestIRO() {
	amt := math.NewIntFromUint64(1_000_000).MulRaw(1e18)
	rollapp := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())

	denom := rollapp.GenesisInfo.NativeDenom.Base
	coin := sdk.NewCoin(denom, amt)
	apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.NewCoins(coin))

	// create IRO plan
	_, err := s.hubApp().IROKeeper.CreatePlan(s.hubCtx(), amt, time.Now(), time.Now().Add(time.Hour), rollapp, irotypes.DefaultBondingCurve(), irotypes.DefaultIncentivePlanParams())
	s.Require().NoError(err)

	// non-genesis transfer should fail, as the bridge is not open
	msg := s.transferMsg(amt, denom)
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	ack, found := s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().NotEqual(successAck, ack) // assert for ack error

	transfersEnabled := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
	s.Require().False(transfersEnabled)

	// - TODO: wrong dest, wrong decimals
	/* --------------------- test invalid genesis transfers --------------------- */
	// - wrong amount
	msg = s.transferMsg(amt.Sub(math.NewInt(100)), denom)
	res, err = s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	ack, found = s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().NotEqual(successAck, ack) // assert for ack error

	transfersEnabled = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
	s.Require().False(transfersEnabled)

	// - wrong denom
	wrongCoin := sdk.NewCoin("bar", amt)
	apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.NewCoins(wrongCoin))
	msg = s.transferMsg(amt, "bar")
	res, err = s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	ack, found = s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().NotEqual(successAck, ack) // assert for ack error

	transfersEnabled = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID()).GenesisState.TransfersEnabled
	s.Require().False(transfersEnabled)

	/* ------------------------------- happy case ------------------------------- */
	// genesis transfer, should pass and enable bridge
	msg = s.transferMsg(amt, denom)
	res, err = s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	// assert the ack succeeded
	bz, found := s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().Equal(successAck, bz) // assert for ack success

	// assert the transfers are enabled
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

// TestInvalidGenesisInfo tests an invalid genesis info
// FIXME: TODO

// TestBridgeDisabledEnabled tests that the bridge is disabled until the genesis bridge is completed
// after the genesis bridge is completed, the bridge should be enabled
func (s *transferGenesisSuite) TestBridgeDisabledEnabled() {
	amt := math.NewIntFromUint64(10000000000000000000)
	denom := "foo"
	coin := sdk.NewCoin(denom, amt)
	apptesting.FundAccount(s.rollappApp(), s.rollappCtx(), s.rollappChain().SenderAccount.GetAddress(), sdk.NewCoins(coin))

	msg := s.transferMsg(amt, denom)
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)

	ack, found := s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().NotEqual(successAck, ack) // assert for ack error

	// create the expected genesis bridge packet
	rollapp := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	packet = s.genesisBridgePacket(rollapp.GenesisInfo.NativeDenom.Base, rollapp.GenesisInfo.NativeDenom.Display, rollapp.GenesisInfo.InitialSupply, nil)

	// send the packet on the rollapp chain
	seq, err := s.path.EndpointB.SendPacket(packet.TimeoutHeight, packet.TimeoutTimestamp, packet.Data)
	s.Require().NoError(err)

	packet.Sequence = seq

	_, err = s.path.EndpointA.RecvPacketWithResult(packet)
	s.Require().NoError(err)

	// assert the ack succeeded
	ack, found = s.hubApp().IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)
	s.Require().Equal(successAck, ack)

	// assert the transfers are enabled
	rollapp = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	s.Require().True(rollapp.GenesisState.TransfersEnabled)

	// assert the transfer now goes through
	res, err = s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err)
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	err = s.path.RelayPacket(packet)
	s.Require().Error(err) // expecting error as no AcknowledgePacket expected to return
}

// TestBridgeEnabled tests that the bridge is enabled after the genesis bridge is completed
func (s *transferGenesisSuite) genesisBridgePacket(denom, display string, initialSupply math.Int, gAccounts []rollapptypes.GenesisAccount) channeltypes.Packet {
	var gb genesisbridge.GenesisBridgeData

	meta := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom: denom,
			},
			{
				Denom:    display,
				Exponent: 18,
			},
		},
		Base:    denom,
		Display: display,
		Name:    denom,
		Symbol:  display,
	}
	s.Require().NoError(meta.Validate()) // sanity check the test is written correctly

	gb.GenesisInfo = genesisbridge.GenesisBridgeInfo{
		GenesisChecksum: "TODO", // FIXME
		Bech32Prefix:    "ethm",
		NativeDenom: rollapptypes.DenomMetadata{
			Base:     meta.Base,
			Display:  meta.DenomUnits[1].Denom,
			Exponent: meta.DenomUnits[1].Exponent,
		},
		InitialSupply:   initialSupply,
		GenesisAccounts: gAccounts,
	}
	gb.NativeDenom = meta

	// FIXME: add genesis transfer

	bz, err := gb.Marshal()
	s.Require().NoError(err)

	msg := channeltypes.NewPacket(
		bz,
		0, // will be set after the submission
		s.path.EndpointB.ChannelConfig.PortID,
		s.path.EndpointB.ChannelID,
		s.path.EndpointA.ChannelConfig.PortID,
		s.path.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		uint64(s.hubCtx().BlockTime().Add(10*time.Minute).UnixNano()),
	)

	return msg
}

func (s *transferGenesisSuite) transferMsg(amt math.Int, denom string) *types.MsgTransfer {
	msg := types.NewMsgTransfer(
		s.path.EndpointB.ChannelConfig.PortID,
		s.path.EndpointB.ChannelID,
		sdk.NewCoin(denom, amt),
		s.rollappChain().SenderAccount.GetAddress().String(),
		s.hubChain().SenderAccount.GetAddress().String(),
		clienttypes.ZeroHeight(),
		uint64(s.hubCtx().BlockTime().Add(10*time.Minute).UnixNano()),
		"",
	)

	return msg
}
