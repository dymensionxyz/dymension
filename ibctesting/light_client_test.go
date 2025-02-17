package ibctesting_test

import (
	"slices"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/sdk-utils/utils/utest"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/stretchr/testify/suite"
)

var canonicalClientConfig = ibctesting.TendermintConfig{
	TrustLevel:      types.DefaultExpectedCanonicalClientParams().TrustLevel,
	TrustingPeriod:  types.DefaultExpectedCanonicalClientParams().TrustingPeriod,
	UnbondingPeriod: types.DefaultExpectedCanonicalClientParams().UnbondingPeriod,
	MaxClockDrift:   types.DefaultExpectedCanonicalClientParams().MaxClockDrift,
}

type lightClientSuite struct {
	utilSuite
	path *ibctesting.Path
}

func TestLightClientSuite(t *testing.T) {
	suite.Run(t, new(lightClientSuite))
}

// TestSetCanonicalClient_ParamsMismatch tests that a client cannot be set as a canonical client
// when the trust requirements do not match
func (s *lightClientSuite) TestSetCanonicalClient_ParamsMismatch() {
	s.createRollapp(false, nil)
	s.registerSequencer()

	currentHeader := s.rollappChain().CurrentHeader
	startHeight := uint64(currentHeader.Height)
	bd := rollapptypes.BlockDescriptor{Height: startHeight, StateRoot: currentHeader.AppHash, Timestamp: currentHeader.Time}

	// create default clients, which should have wrong params and can't be set as canonical client
	s.path = s.newTransferPath(s.hubChain(), s.rollappChain())
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
		2, // revision
		&rollapptypes.BlockDescriptors{BD: []rollapptypes.BlockDescriptor{bd, bdNext}},
	)
	_, err := s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Require().NoError(err)

	setCanonMsg := &types.MsgSetCanonicalClient{
		Signer: s.hubChain().SenderAccount.GetAddress().String(), ClientId: s.path.EndpointA.ClientID,
	}
	_, err = s.lightclientMsgServer().SetCanonicalClient(s.hubCtx(), setCanonMsg)
	utest.IsErr(s.Require(), err, lightclientkeeper.ErrParamsMismatch)

	_, found := s.hubApp().LightClientKeeper.GetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID)
	s.False(found)
}

func (s *lightClientSuite) TestSetCanonicalClient_ConsStateMismatch() {
	s.createRollapp(false, nil)
	s.registerSequencer()

	h := s.rollappChain().CurrentHeader
	startHeight := uint64(h.Height)
	bd := rollapptypes.BlockDescriptor{Height: uint64(h.Height), StateRoot: h.AppHash, Timestamp: h.Time}

	s.createCompatibleClient()

	h = s.rollappChain().CurrentHeader
	bdNext := rollapptypes.BlockDescriptor{Height: uint64(h.Height), StateRoot: h.AppHash, Timestamp: h.Time}

	setCanonMsg := &types.MsgSetCanonicalClient{
		Signer: s.hubChain().SenderAccount.GetAddress().String(), ClientId: s.path.EndpointA.ClientID,
	}

	// Update the rollapp state so we could attempt to set the canonical client
	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		startHeight,
		2,
		2, // revision
		&rollapptypes.BlockDescriptors{BD: []rollapptypes.BlockDescriptor{bd, bdNext}},
	)
	_, err := s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Require().NoError(err)

	c, _ := s.hubChain().App.GetIBCKeeper().ClientKeeper.GetClientState(s.hubCtx(), s.path.EndpointA.ClientID)
	tmClient, _ := c.(*ibctm.ClientState)
	cs, _ := s.hubChain().App.GetIBCKeeper().ClientKeeper.GetClientConsensusState(s.hubCtx(), s.path.EndpointA.ClientID, tmClient.GetLatestHeight())
	tmConsState, _ := cs.(*ibctm.ConsensusState)
	slices.Reverse(tmConsState.NextValidatorsHash) // make it wrong
	s.hubChain().App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(s.hubCtx(), s.path.EndpointA.ClientID, tmClient.GetLatestHeight(), tmConsState)

	_, err = s.lightclientMsgServer().SetCanonicalClient(s.hubCtx(), setCanonMsg)
	utest.IsErr(s.Require(), err, lightclientkeeper.ErrMismatch)
}

func (s *lightClientSuite) TestSetCanonicalClient_FailsIncompatibleState() {
	s.createRollapp(false, nil)
	s.registerSequencer()

	startHeight := uint64(s.rollappChain().CurrentHeader.Height)

	s.createCompatibleClient()

	h := s.rollappChain().CurrentHeader
	// first bd will have wrong state root in regard to the ibc headers
	bds := []rollapptypes.BlockDescriptor{
		{Height: startHeight, StateRoot: h.AppHash, Timestamp: h.Time},
		{Height: uint64(h.Height), StateRoot: h.AppHash, Timestamp: h.Time},
	}

	setCanonMsg := &types.MsgSetCanonicalClient{
		Signer: s.hubChain().SenderAccount.GetAddress().String(), ClientId: s.path.EndpointA.ClientID,
	}

	// Update the rollapp state so we could attempt to set the canonical client
	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		startHeight,
		2,
		2, // revision
		&rollapptypes.BlockDescriptors{bds},
	)
	_, err := s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Require().NoError(err)

	_, err = s.lightclientMsgServer().SetCanonicalClient(s.hubCtx(), setCanonMsg)
	utest.IsErr(s.Require(), err, lightclientkeeper.ErrMismatch)
}

// TestSetCanonicalClient_Succeeds tests that a client can be set as a canonical client
func (s *lightClientSuite) TestSetCanonicalClient_Succeeds() {
	s.createRollapp(false, nil)
	s.registerSequencer()

	currentHeader := s.rollappChain().CurrentHeader
	startHeight := uint64(currentHeader.Height)
	bd := rollapptypes.BlockDescriptor{Height: startHeight, StateRoot: currentHeader.AppHash, Timestamp: currentHeader.Time}

	s.createCompatibleClient()

	currentHeader = s.rollappChain().CurrentHeader
	bdNext := rollapptypes.BlockDescriptor{Height: uint64(currentHeader.Height), StateRoot: currentHeader.AppHash, Timestamp: currentHeader.Time}

	// no state update, it should fail
	setCanonMsg := &types.MsgSetCanonicalClient{
		Signer: s.hubChain().SenderAccount.GetAddress().String(), ClientId: s.path.EndpointA.ClientID,
	}
	_, err := s.lightclientMsgServer().SetCanonicalClient(s.hubCtx(), setCanonMsg)
	s.Require().Error(err)

	// FIXME: we have wrapper for this:
	// Update the rollapp state - this will trigger the check for prospective canonical client
	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		startHeight,
		2,
		2, // revision
		&rollapptypes.BlockDescriptors{BD: []rollapptypes.BlockDescriptor{bd, bdNext}},
	)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Require().NoError(err)

	_, err = s.lightclientMsgServer().SetCanonicalClient(s.hubCtx(), setCanonMsg)
	s.Require().NoError(err)

	canonClientID, found := s.hubApp().LightClientKeeper.GetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID)
	s.Require().True(found)
	s.Equal(s.path.EndpointA.ClientID, canonClientID)
}

func (s *lightClientSuite) TestUpdateClientAcceptedOptimistically() {
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
	seqAddr, err := s.hubApp().LightClientKeeper.GetSigner(s.hubCtx(), s.path.EndpointA.ClientID, s.path.EndpointA.GetClientState().GetLatestHeight().GetRevisionHeight())
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
	s.NoError(err)

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
		bds.BD[0].Height,
		uint64(len(bds.BD)),
		2, // revision
		&bds,
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
	// There should not be any optimistic updates as the roots were verified
	_, err = s.hubApp().LightClientKeeper.GetSigner(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
	s.Error(err)
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
	s.NoError(err)

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
		bds.BD[0].Height, uint64(len(bds.BD)), 2, &bds,
	)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.NoError(err)

	msg, err := clienttypes.NewMsgUpdateClient(
		s.path.EndpointA.ClientID, header,
		s.path.EndpointA.Chain.SenderAccount.GetAddress().String(),
	)
	s.NoError(err)

	// As there was incompatible stateinfo found, should prevent light client update.
	_, err = s.hubChain().SendMsgs([]sdk.Msg{msg}...)
	s.Require().ErrorContains(err, types.ErrTimestampMismatch.Error())
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
	s.NoError(err)

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
	_, err = s.hubApp().LightClientKeeper.GetSigner(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
	s.NoError(err)

	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		bds.BD[0].Height, uint64(len(bds.BD)), 2, &bds,
	)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.NoError(err)
	// The optimistic update valhash should be removed as the state has been confirmed to be compatible
	_, err = s.hubApp().LightClientKeeper.GetSigner(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
	s.Error(err)
	// Ensuring that the stateinfo is now upto date as well
	state, found := s.hubApp().RollappKeeper.GetLatestStateInfo(s.hubCtx(), s.rollappChain().ChainID)
	s.True(found)
	s.True(state.ContainsHeight(uint64(header.Header.Height)))
}

// TestAfterUpdateState_OptimisticUpdateExists_NotCompatible tests that a state info update is rejected in case the state is not compatible
// with the light client headers
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
	s.NoError(err)

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
	_, err = s.hubApp().LightClientKeeper.GetSigner(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
	s.NoError(err)

	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		bds.BD[0].Height, uint64(len(bds.BD)), 2, &bds,
	)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Error(err)
}

// Test the rollback flow for a light client
// - do some client updates
// - trigger rollback
// - validate rollback:
//   - check if the client is frozen
//   - validate client updates are blocked
//   - validate future consensus states are cleared
//
// - resolve hard fork
//   - validate client is unfrozen and hard fork is resolved
//   - validate the client is updated
//
// - validate client updates are allowed
func (s *lightClientSuite) TestAfterUpdateState_Rollback() {
	s.createRollapp(false, nil)
	s.registerSequencer()
	s.path = s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.SetupConnections(s.path)
	s.hubApp().LightClientKeeper.SetCanonicalClient(s.hubCtx(), s.rollappChain().ChainID, s.path.EndpointA.ClientID)
	s.coordinator.CreateChannels(s.path)

	bds := rollapptypes.BlockDescriptors{}
	signerHeights := []int64{}

	for i := 0; i < 20; i++ {
		s.coordinator.CommitBlock(s.hubChain(), s.rollappChain())

		lastHeader := s.rollappChain().LastHeader
		bd := rollapptypes.BlockDescriptor{Height: uint64(lastHeader.Header.Height), StateRoot: lastHeader.Header.AppHash, Timestamp: lastHeader.Header.Time}
		bds.BD = append(bds.BD, bd)

		if i%4 == 0 {
			header, err := s.path.EndpointA.Chain.ConstructUpdateTMClientHeader(s.path.EndpointA.Counterparty.Chain, s.path.EndpointA.ClientID)
			s.NoError(err)
			msg, err := clienttypes.NewMsgUpdateClient(
				s.path.EndpointA.ClientID, header,
				s.path.EndpointA.Chain.SenderAccount.GetAddress().String(),
			)
			s.NoError(err)
			_, err = s.path.EndpointA.Chain.SendMsgs(msg)
			s.NoError(err)

			// save signers
			_, err = s.hubApp().LightClientKeeper.GetSigner(s.hubCtx(), s.path.EndpointA.ClientID, uint64(header.Header.Height))
			s.NoError(err)
			signerHeights = append(signerHeights, header.Header.Height)
		}

	}

	// get number of consensus states before rollback
	csBeforeRollback := s.hubApp().IBCKeeper.ClientKeeper.GetAllConsensusStates(s.hubCtx())[0].ConsensusStates

	// Trigger rollback / simulate fork
	nRolledBack := uint64(5)
	lastValidHeight := uint64(s.rollappChain().LastHeader.Header.Height) - nRolledBack
	newRevisionHeight := lastValidHeight + 1
	ra := s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), s.rollappChain().ChainID)
	newRevision := ra.LatestRevision().Number + 1
	ra.Revisions = append(ra.Revisions, rollapptypes.Revision{StartHeight: newRevisionHeight, Number: newRevision})
	s.hubApp().RollappKeeper.SetRollapp(s.hubCtx(), ra)
	err := s.hubApp().LightClientKeeper.RollbackCanonicalClient(s.hubCtx(), s.rollappChain().ChainID, lastValidHeight)
	s.Require().NoError(err)

	clientState, found := s.hubApp().IBCKeeper.ClientKeeper.GetClientState(s.hubCtx(), s.path.EndpointA.ClientID)
	s.True(found)
	tmClientState, ok := clientState.(*ibctm.ClientState)
	s.True(ok)

	// Check if the client is frozen
	s.True(!tmClientState.FrozenHeight.IsZero(), "Client should be frozen after rollback")

	// Validate future consensus states are cleared
	csAfterRollback := s.hubApp().IBCKeeper.ClientKeeper.GetAllConsensusStates(s.hubCtx())[0].ConsensusStates
	s.Require().Less(len(csAfterRollback), len(csBeforeRollback), "Consensus states should be cleared after rollback")
	for height := uint64(0); height <= uint64(s.rollappChain().LastHeader.Header.Height); height++ {
		_, found := s.hubApp().IBCKeeper.ClientKeeper.GetClientConsensusState(s.hubCtx(), s.path.EndpointA.ClientID, clienttypes.NewHeight(1, height))
		if height >= newRevisionHeight {
			s.False(found, "Consensus state should be cleared for height %d", height)
		}
	}

	// validate signers are removed
	cnt := 0
	for _, height := range signerHeights {
		_, err := s.hubApp().LightClientKeeper.GetSigner(s.hubCtx(), s.path.EndpointA.ClientID, uint64(height))
		if height >= int64(lastValidHeight) {
			s.Error(err, "Signer should be removed for height %d", height)
		} else {
			s.NoError(err, "Signer should not be removed for height %d", height)
			cnt++
		}
	}
	s.Require().Less(cnt, len(signerHeights), "Signers should be removed after rollback")

	// a bit of a hack to make sure the ibc go testing framework can update, since we can't get inside to pass a revision
	ra = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), s.rollappChain().ChainID)
	ra.Revisions[len(ra.Revisions)-1].Number = 2
	s.hubApp().RollappKeeper.SetRollapp(s.hubCtx(), ra)
	// update should fail as client is frozen
	s.Error(s.path.EndpointA.UpdateClient())

	// back the revision
	ra.Revisions[len(ra.Revisions)-1].Number = newRevision
	s.hubApp().RollappKeeper.SetRollapp(s.hubCtx(), ra)

	// submit a state info update to resolve the hard fork
	bds.BD = bds.BD[len(bds.BD)-int(nRolledBack):]
	blockDescriptors := &rollapptypes.BlockDescriptors{BD: bds.BD}
	msgUpdateState := rollapptypes.NewMsgUpdateState(
		s.hubChain().SenderAccount.GetAddress().String(),
		rollappChainID(),
		"mock-da-path",
		bds.BD[0].Height,
		uint64(len(bds.BD)),
		3,
		blockDescriptors,
	)
	_, err = s.rollappMsgServer().UpdateState(s.hubCtx(), msgUpdateState)
	s.Require().NoError(err, "update state")

	// Test resolve hard fork
	clientState, found = s.hubApp().IBCKeeper.ClientKeeper.GetClientState(s.hubCtx(), s.path.EndpointA.ClientID)
	s.True(found)
	// Verify that the client is unfrozen and hard fork is resolved
	s.True(clientState.(*ibctm.ClientState).FrozenHeight.IsZero(), "Client should be unfrozen after hard fork resolution")
	// Verify that the client is updated with the height of the first block descriptor
	s.Require().Equal(bds.BD[0].Height, clientState.GetLatestHeight().GetRevisionHeight())
	_, ok = s.hubApp().IBCKeeper.ClientKeeper.GetLatestClientConsensusState(s.hubCtx(), s.path.EndpointA.ClientID)
	s.True(ok)

	// validate client updates are no longer blocked
	s.coordinator.CommitBlock(s.rollappChain())

	// a bit of a hack to make sure the ibc go testing framework can update, since we can't get inside to pass a revision
	ra = s.hubApp().RollappKeeper.MustGetRollapp(s.hubCtx(), s.rollappChain().ChainID)
	ra.Revisions[len(ra.Revisions)-1].Number = 2
	s.hubApp().RollappKeeper.SetRollapp(s.hubCtx(), ra)

	s.NoError(s.path.EndpointA.UpdateClient())
}

func (s *lightClientSuite) createCompatibleClient() {
	// create a custom tm client which matches the trust requirements of a canonical client
	endpointA := ibctesting.NewEndpoint(s.hubChain(), &canonicalClientConfig, ibctesting.NewConnectionConfig(), ibctesting.NewChannelConfig())
	endpointB := ibctesting.NewEndpoint(s.rollappChain(), ibctesting.NewTendermintConfig(), ibctesting.NewConnectionConfig(), ibctesting.NewChannelConfig())
	endpointA.Counterparty = endpointB
	endpointB.Counterparty = endpointA
	s.path = &ibctesting.Path{EndpointA: endpointA, EndpointB: endpointB}

	// Creating the tm client - this will take us to the next block
	s.NoError(s.path.EndpointA.CreateClient())
}
