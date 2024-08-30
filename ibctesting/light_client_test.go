package ibctesting_test

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/testing/simapp"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/stretchr/testify/suite"
)

var canonicalClientConfig = ibctesting.TendermintConfig{
	TrustLevel:      types.ExpectedCanonicalClientParams.TrustLevel,
	TrustingPeriod:  types.ExpectedCanonicalClientParams.TrustingPeriod,
	UnbondingPeriod: types.ExpectedCanonicalClientParams.UnbondingPeriod,
	MaxClockDrift:   types.ExpectedCanonicalClientParams.MaxClockDrift,
}

type lightClientSuite struct {
	utilSuite
	path *ibctesting.Path
}

func TestLightClientSuite(t *testing.T) {
	suite.Run(t, new(lightClientSuite))
}

func (s *lightClientSuite) TestSetCanonicalClient_FailsTrustRequirements() {
	s.createRollapp(false, nil)
	s.registerSequencer()
	// The default tm client does not match the trust requirements of a canonical client.
	// So it should not be set as one.
	s.path = s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.SetupClients(s.path)

	// Update rollapp state - this will trigger the check for prospective canonical client
	currentRollappBlockHeight := uint64(s.rollappChain().App.LastBlockHeight())
	s.updateRollappState(currentRollappBlockHeight)

	_, found := s.hubApp().LightClientKeeper.GetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID)
	s.False(found)
}

func (s *lightClientSuite) TestSetCanonicalClient_FailsIncompatibleState() {
	s.createRollapp(false, nil)
	s.registerSequencer()
	// create a custom tm client which matches the trust requirements of a canonical client
	endpointA := ibctesting.NewEndpoint(s.hubChain(), &canonicalClientConfig, ibctesting.NewConnectionConfig(), ibctesting.NewChannelConfig())
	endpointB := ibctesting.NewEndpoint(s.rollappChain(), ibctesting.NewTendermintConfig(), ibctesting.NewConnectionConfig(), ibctesting.NewChannelConfig())
	endpointA.Counterparty = endpointB
	endpointB.Counterparty = endpointA
	s.path = &ibctesting.Path{EndpointA: endpointA, EndpointB: endpointB}

	// Creating the tm client - this will take us to the next block
	s.coordinator.SetupClients(s.path)

	// Update the rollapp state - this will trigger the check for prospective canonical client
	// The block descriptor root has dummy values and will not match the IBC roots for the same height
	currentRollappBlockHeight := uint64(s.rollappChain().App.LastBlockHeight())
	s.updateRollappState(currentRollappBlockHeight)

	_, found := s.hubApp().LightClientKeeper.GetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID)
	s.False(found)
}

func (s *lightClientSuite) TestSetCanonicalClient_Succeeds() {
	s.createRollapp(false, nil)
	s.registerSequencer()
	// create a custom tm client which matches the trust requirements of a canonical client
	endpointA := ibctesting.NewEndpoint(s.hubChain(), &canonicalClientConfig, ibctesting.NewConnectionConfig(), ibctesting.NewChannelConfig())
	endpointB := ibctesting.NewEndpoint(s.rollappChain(), ibctesting.NewTendermintConfig(), ibctesting.NewConnectionConfig(), ibctesting.NewChannelConfig())
	endpointA.Counterparty = endpointB
	endpointB.Counterparty = endpointA
	s.path = &ibctesting.Path{EndpointA: endpointA, EndpointB: endpointB}

	currentHeader := s.rollappChain().CurrentHeader
	startHeight := uint64(currentHeader.Height)
	bd := rollapptypes.BlockDescriptor{Height: startHeight, StateRoot: currentHeader.AppHash, Timestamp: currentHeader.Time}

	// Creating the tm client - this will take us to the next block
	s.NoError(s.path.EndpointA.CreateClient())

	currentHeader = s.rollappChain().CurrentHeader
	bdNext := rollapptypes.BlockDescriptor{Height: uint64(currentHeader.Height), StateRoot: currentHeader.AppHash, Timestamp: currentHeader.Time}

	// Update the rollapp state - this will trigger the check for prospective canonical client
	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		startHeight,
		2,
		&rollapptypes.BlockDescriptors{BD: []rollapptypes.BlockDescriptor{bd, bdNext}},
	)
	_, err := s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Require().NoError(err)

	canonClientID, found := s.hubApp().LightClientKeeper.GetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID)
	s.True(found)
	s.Equal(endpointA.ClientID, canonClientID)
}

func (s *lightClientSuite) TestMsgUpdateClient_StateUpdateDoesntExist() {
	s.createRollapp(false, nil)
	s.registerSequencer()
	currentRollappBlockHeight := uint64(s.rollappChain().App.LastBlockHeight())
	s.updateRollappState(currentRollappBlockHeight)
	s.path = s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.SetupClients(s.path)
	s.hubApp().LightClientKeeper.SetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID, s.path.EndpointA.ClientID)

	for i := 0; i < 10; i++ {
		s.hubChain().NextBlock()
		s.rollappChain().NextBlock()
	}

	s.NoError(s.path.EndpointA.UpdateClient())
	// As there was no stateinfo found for the height, should have accepted the update optimistically.
	seqValHash, found := s.hubApp().LightClientKeeper.GetConsensusStateValHash(s.hubCtx(), s.path.EndpointA.ClientID, s.path.EndpointA.GetClientState().GetLatestHeight().GetRevisionHeight())
	s.True(found)
	seqAddr, err := s.hubApp().LightClientKeeper.GetSequencerFromValHash(s.hubCtx(), seqValHash)
	s.NoError(err)
	s.Equal(s.hubChain().SenderAccount.GetAddress().String(), seqAddr)
}

func (s *lightClientSuite) TestMsgUpdateClient_StateUpdateExists_Compatible() {
	s.createRollapp(false, nil)
	s.registerSequencer()
	s.path = s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.SetupClients(s.path)
	s.NoError(s.path.EndpointA.UpdateClient())
	s.hubApp().LightClientKeeper.SetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID, s.path.EndpointA.ClientID)

	bds := rollapptypes.BlockDescriptors{}
	for i := 0; i < 2; i++ {
		lastHeader := s.rollappChain().LastHeader
		bd := rollapptypes.BlockDescriptor{Height: uint64(lastHeader.Header.Height), StateRoot: lastHeader.Header.AppHash, Timestamp: lastHeader.Header.Time}
		bds.BD = append(bds.BD, bd)
		s.hubChain().NextBlock()
		s.rollappChain().NextBlock()
	}
	header, err := s.path.EndpointA.Chain.ConstructUpdateTMClientHeader(s.path.EndpointA.Counterparty.Chain, s.path.EndpointA.ClientID)

	for i := 0; i < 2; i++ {
		lastHeader := s.rollappChain().LastHeader
		bd := rollapptypes.BlockDescriptor{Height: uint64(lastHeader.Header.Height), StateRoot: lastHeader.Header.AppHash, Timestamp: lastHeader.Header.Time}
		bds.BD = append(bds.BD, bd)
		s.hubChain().NextBlock()
		s.rollappChain().NextBlock()
	}
	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		bds.BD[0].Height, uint64(len(bds.BD)), &bds,
	)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.NoError(err)

	msg, err := clienttypes.NewMsgUpdateClient(
		s.path.EndpointA.ClientID, header,
		s.path.EndpointA.Chain.SenderAccount.GetAddress().String(),
	)
	s.NoError(err)

	// As there was compatible stateinfo found, should accept the ClientUpdate without any error.
	_, err = s.path.EndpointA.Chain.SendMsgs(msg)
	s.NoError(err)
	s.Equal(uint64(header.Header.Height), s.path.EndpointA.GetClientState().GetLatestHeight().GetRevisionHeight())
	// There shouldnt be any optimistic updates as the roots were verified
	_, found := s.hubApp().LightClientKeeper.GetConsensusStateValHash(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
	s.False(found)
}

func (s *lightClientSuite) TestMsgUpdateClient_StateUpdateExists_NotCompatible() {
	s.createRollapp(false, nil)
	s.registerSequencer()
	s.path = s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.SetupClients(s.path)
	s.NoError(s.path.EndpointA.UpdateClient())
	s.hubApp().LightClientKeeper.SetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID, s.path.EndpointA.ClientID)

	bds := rollapptypes.BlockDescriptors{}
	for i := 0; i < 2; i++ {
		lastHeader := s.rollappChain().LastHeader
		bd := rollapptypes.BlockDescriptor{Height: uint64(lastHeader.Header.Height), StateRoot: lastHeader.Header.AppHash, Timestamp: lastHeader.Header.Time}
		bds.BD = append(bds.BD, bd)
		s.hubChain().NextBlock()
		s.rollappChain().NextBlock()
	}
	header, err := s.path.EndpointA.Chain.ConstructUpdateTMClientHeader(s.path.EndpointA.Counterparty.Chain, s.path.EndpointA.ClientID)

	for i := 0; i < 2; i++ {
		lastHeader := s.rollappChain().LastHeader
		bd := rollapptypes.BlockDescriptor{Height: uint64(lastHeader.Header.Height), StateRoot: lastHeader.Header.AppHash, Timestamp: lastHeader.Header.Time}
		bd.Timestamp = bd.Timestamp.AddDate(0, 0, 1) // wrong timestamp to cause state mismatch
		bds.BD = append(bds.BD, bd)
		s.hubChain().NextBlock()
		s.rollappChain().NextBlock()
	}
	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		bds.BD[0].Height, uint64(len(bds.BD)), &bds,
	)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.NoError(err)

	msg, err := clienttypes.NewMsgUpdateClient(
		s.path.EndpointA.ClientID, header,
		s.path.EndpointA.Chain.SenderAccount.GetAddress().String(),
	)
	s.NoError(err)

	// As there was incompatible stateinfo found, should prevent light client update.
	s.path.EndpointA.Chain.Coordinator.UpdateTimeForChain(s.path.EndpointA.Chain)
	_, _, err = simapp.SignAndDeliver( // Explicitly submitting msg as we expect it to fail
		s.path.EndpointA.Chain.T,
		s.path.EndpointA.Chain.TxConfig,
		s.path.EndpointA.Chain.App.GetBaseApp(),
		s.path.EndpointA.Chain.GetContext().BlockHeader(),
		[]sdk.Msg{msg},
		s.path.EndpointA.Chain.ChainID,
		[]uint64{s.path.EndpointA.Chain.SenderAccount.GetAccountNumber()},
		[]uint64{s.path.EndpointA.Chain.SenderAccount.GetSequence()},
		true, false, s.path.EndpointA.Chain.SenderPrivKey,
	)
	s.Error(err)
	s.True(errorsmod.IsOf(err, types.ErrTimestampMismatch))
}

func (s *lightClientSuite) TestAfterUpdateState_OptimisticUpdateExists_Compatible() {
	s.createRollapp(false, nil)
	s.registerSequencer()
	s.path = s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.SetupClients(s.path)
	s.NoError(s.path.EndpointA.UpdateClient())
	s.hubApp().LightClientKeeper.SetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID, s.path.EndpointA.ClientID)

	bds := rollapptypes.BlockDescriptors{}
	for i := 0; i < 2; i++ {
		lastHeader := s.rollappChain().LastHeader
		bd := rollapptypes.BlockDescriptor{Height: uint64(lastHeader.Header.Height), StateRoot: lastHeader.Header.AppHash, Timestamp: lastHeader.Header.Time}
		bds.BD = append(bds.BD, bd)
		s.hubChain().NextBlock()
		s.rollappChain().NextBlock()
	}
	header, err := s.path.EndpointA.Chain.ConstructUpdateTMClientHeader(s.path.EndpointA.Counterparty.Chain, s.path.EndpointA.ClientID)

	for i := 0; i < 2; i++ {
		lastHeader := s.rollappChain().LastHeader
		bd := rollapptypes.BlockDescriptor{Height: uint64(lastHeader.Header.Height), StateRoot: lastHeader.Header.AppHash, Timestamp: lastHeader.Header.Time}
		bds.BD = append(bds.BD, bd)
		s.hubChain().NextBlock()
		s.rollappChain().NextBlock()
	}

	msg, err := clienttypes.NewMsgUpdateClient(
		s.path.EndpointA.ClientID, header,
		s.path.EndpointA.Chain.SenderAccount.GetAddress().String(),
	)
	s.NoError(err)
	_, err = s.path.EndpointA.Chain.SendMsgs(msg)
	s.NoError(err)
	// There should be one optimistic update for the header height
	_, found := s.hubApp().LightClientKeeper.GetConsensusStateValHash(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
	s.True(found)

	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		bds.BD[0].Height, uint64(len(bds.BD)), &bds,
	)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.NoError(err)
	// The optimistic update valhash should be removed as the state has been confirmed to be compatible
	_, found = s.hubApp().LightClientKeeper.GetConsensusStateValHash(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
	s.False(found)
	// Ensuring that the stateinfo is now upto date as well
	state, found := s.hubApp().RollappKeeper.GetLatestStateInfo(s.hubCtx(), s.rollappChain().ChainID)
	s.True(found)
	s.True(state.ContainsHeight(uint64(header.Header.Height)))
}

func (s *lightClientSuite) TestAfterUpdateState_OptimisticUpdateExists_NotCompatible() {
	s.createRollapp(false, nil)
	s.registerSequencer()
	s.path = s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.SetupConnections(s.path)
	s.hubApp().LightClientKeeper.SetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID, s.path.EndpointA.ClientID)
	s.coordinator.CreateChannels(s.path)
	s.NoError(s.path.EndpointA.UpdateClient())

	bds := rollapptypes.BlockDescriptors{}
	for i := 0; i < 2; i++ {
		lastHeader := s.rollappChain().LastHeader
		bd := rollapptypes.BlockDescriptor{Height: uint64(lastHeader.Header.Height), StateRoot: lastHeader.Header.AppHash, Timestamp: lastHeader.Header.Time}
		bds.BD = append(bds.BD, bd)
		s.hubChain().NextBlock()
		s.rollappChain().NextBlock()
	}
	header, err := s.path.EndpointA.Chain.ConstructUpdateTMClientHeader(s.path.EndpointA.Counterparty.Chain, s.path.EndpointA.ClientID)

	for i := 0; i < 2; i++ {
		lastHeader := s.rollappChain().LastHeader
		bd := rollapptypes.BlockDescriptor{Height: uint64(lastHeader.Header.Height), StateRoot: lastHeader.Header.AppHash, Timestamp: lastHeader.Header.Time}
		bd.Timestamp = bd.Timestamp.AddDate(0, 0, 1) // wrong timestamp to cause state mismatch
		bds.BD = append(bds.BD, bd)
		s.hubChain().NextBlock()
		s.rollappChain().NextBlock()
	}

	msg, err := clienttypes.NewMsgUpdateClient(
		s.path.EndpointA.ClientID, header,
		s.path.EndpointA.Chain.SenderAccount.GetAddress().String(),
	)
	s.NoError(err)
	_, err = s.path.EndpointA.Chain.SendMsgs(msg)
	s.NoError(err)
	// There should be one optimistic update for the header height
	_, found := s.hubApp().LightClientKeeper.GetConsensusStateValHash(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
	s.True(found)

	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		bds.BD[0].Height, uint64(len(bds.BD)), &bds,
	)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Error(err)
	// The optimistic update valhash should be removed as part of fraud handling
	_, found = s.hubApp().LightClientKeeper.GetConsensusStateValHash(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
	s.False(found)
	// Ensuring that the rollapp is now frozen as part of fraud handling
	rollapp, _ := s.hubApp().RollappKeeper.GetRollapp(s.hubCtx(), s.rollappChain().ChainID)
	s.True(rollapp.Frozen)
}
