package ibctesting_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
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
	// create the expected genesis bridge packet
	rollapp := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	packet := s.genesisBridgePacket(rollapp.GenesisInfo)

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

	rollapp = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	// assert the transfers are enabled
	s.Require().True(rollapp.GenesisState.TransfersEnabled)

	// assert denom registered
	expectedIBCdenom := types.ParseDenomTrace(types.GetPrefixedDenom(s.path.EndpointB.ChannelConfig.PortID, s.path.EndpointB.ChannelID, rollapp.GenesisInfo.NativeDenom.Base)).IBCDenom()
	metadata, found := s.hubApp().BankKeeper.GetDenomMetaData(s.hubCtx(), expectedIBCdenom)
	s.Require().True(found)
	s.Require().Equal(rollapp.GenesisInfo.NativeDenom.Display, metadata.Display)
}

// TestHappyPath_GenesisAccounts tests a valid genesis info with genesis accounts
func (s *transferGenesisSuite) TestHappyPath_GenesisAccounts() {
	gAddr := s.rollappChain().SenderAccount.GetAddress()
	gAccounts := []rollapptypes.GenesisAccount{
		{
			Address: gAddr.String(),
			Amount:  math.NewIntFromUint64(10000000000000000000),
		},
	}
	s.addGenesisAccounts(gAccounts)

	// create the expected genesis bridge packet
	rollapp := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	packet := s.genesisBridgePacket(rollapp.GenesisInfo)

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

	// assert the genesis accounts were funded
	ibcDenom := types.ParseDenomTrace(types.GetPrefixedDenom(s.path.EndpointB.ChannelConfig.PortID, s.path.EndpointB.ChannelID, rollapp.GenesisInfo.NativeDenom.Base)).IBCDenom()
	balance := s.hubApp().BankKeeper.GetBalance(s.hubCtx(), gAddr, ibcDenom)
	s.Require().Equal(gAccounts[0].Amount, balance.Amount)
}

// TestHappyPath_GenesisAccounts_IRO tests a valid genesis info with genesis accounts, including IRO plan
// We expect the IRO plan to be settled once the genesis bridge is completed
func (s *transferGenesisSuite) TestIRO() {
	amt := math.NewIntFromUint64(1_000_000).MulRaw(1e18)
	rollapp := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())

	// create IRO plan
	_, err := s.hubApp().IROKeeper.CreatePlan(s.hubCtx(), amt, time.Now(), time.Now().Add(time.Hour), rollapp, irotypes.DefaultBondingCurve(), irotypes.DefaultIncentivePlanParams())
	s.Require().NoError(err)

	// create the expected genesis bridge packet
	rollapp = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), rollappChainID())
	packet := s.genesisBridgePacket(rollapp.GenesisInfo)

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

	// the iro plan should be settled
	plan, found := s.hubApp().IROKeeper.GetPlanByRollapp(s.hubCtx(), rollappChainID())
	s.Require().True(found)
	expectedIBCdenom := types.ParseDenomTrace(types.GetPrefixedDenom(s.path.EndpointB.ChannelConfig.PortID, s.path.EndpointB.ChannelID, rollapp.GenesisInfo.NativeDenom.Base)).IBCDenom()
	s.Require().Equal(plan.SettledDenom, expectedIBCdenom)
}

// TestInvalidGenesisInfo tests an invalid genesis info
// FIXME: TODO
/*
// - TODO: wrong dest, wrong decimals

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

*/

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
	packet = s.genesisBridgePacket(rollapp.GenesisInfo)

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

/* ---------------------------------- utils --------------------------------- */
// genesisBridgePacket creates a genesis bridge packet with the given parameters
func (s *transferGenesisSuite) genesisBridgePacket(raGenesisInfo rollapptypes.GenesisInfo) channeltypes.Packet {
	denom := raGenesisInfo.NativeDenom.Base
	display := raGenesisInfo.NativeDenom.Display
	initialSupply := raGenesisInfo.InitialSupply
	gAccounts := raGenesisInfo.GenesisAccounts

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

	// add genesis transfer if needed
	if len(gAccounts) > 0 {
		total := math.ZeroInt()
		for _, acc := range gAccounts {
			total = total.Add(acc.Amount)
		}

		gTransfer := transfertypes.NewFungibleTokenPacketData(
			denom,
			total.String(),
			s.rollappChain().SenderAccount.GetAddress().String(),
			genesisbridge.HubRecipient,
			"",
		)
		gb.GenesisTransfer = &gTransfer
	}

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
