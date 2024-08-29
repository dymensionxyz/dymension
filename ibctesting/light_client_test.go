package ibctesting_test

import (
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"testing"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/stretchr/testify/suite"
)

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
	newTmConfig := ibctesting.TendermintConfig{
		TrustLevel:      types.ExpectedCanonicalClientParams.TrustLevel,
		TrustingPeriod:  types.ExpectedCanonicalClientParams.TrustingPeriod,
		UnbondingPeriod: types.ExpectedCanonicalClientParams.UnbondingPeriod,
		MaxClockDrift:   types.ExpectedCanonicalClientParams.MaxClockDrift,
	}
	endpointA := ibctesting.NewEndpoint(s.hubChain(), &newTmConfig, ibctesting.NewConnectionConfig(), ibctesting.NewChannelConfig())
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
	newTmConfig := ibctesting.TendermintConfig{
		TrustLevel:      types.ExpectedCanonicalClientParams.TrustLevel,
		TrustingPeriod:  types.ExpectedCanonicalClientParams.TrustingPeriod,
		UnbondingPeriod: types.ExpectedCanonicalClientParams.UnbondingPeriod,
		MaxClockDrift:   types.ExpectedCanonicalClientParams.MaxClockDrift,
	}
	endpointA := ibctesting.NewEndpoint(s.hubChain(), &newTmConfig, ibctesting.NewConnectionConfig(), ibctesting.NewChannelConfig())
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
	s.Equal("07-tendermint-0", canonClientID)
}
