package irctesting

import (
	"time"

	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

var (
	hash32 = []byte("12345678901234567890123456789012")
)

// DymintTestChainClient is a wrapper for baseTestChainClient
// it is used to intersept the NextBlock() function
// in order to track after the app.Commit() and get an AppHash.
// The AppHash is used for generating BD
type DymintTestChainClient struct {
	baseTestChainClient ibctesting.TestChainClientI
	baseTestChain       *ibctesting.TestChain
	bds                 *rollapptypes.BlockDescriptors
}

func (dymintC *DymintTestChainClient) GetContext() sdk.Context {
	return dymintC.baseTestChainClient.GetContext()
}
func (dymintC *DymintTestChainClient) NextBlock() {
	// NextBlock sets the last header to the current header and increments the current header to be
	// at the next block height.
	// This function must only be called after app.Commit() occurs
	// The app.Commit() calculate the new AppHash which gets into the BlockHeader
	dymintC.baseTestChainClient.NextBlock()
	// app.Commit() resets the deliver state (which holds the changes of the multistore as created in the block)
	// NextBlock opens a new block and calls to BeginBlock, which initializes the deliver state
	// so we can access the last committed block only after NextBlock() and not between app.Commit() & NextBlock()
	// otherwise, GetContext() returns nil
	bd := rollapptypes.BlockDescriptor{
		Height:                 uint64(dymintC.baseTestChain.GetContext().BlockHeader().Height),
		StateRoot:              dymintC.baseTestChain.GetContext().BlockHeader().AppHash,
		IntermediateStatesRoot: hash32,
	}
	// add new BD to list
	dymintC.bds.BD = append(dymintC.bds.BD, bd)
}
func (dymintC *DymintTestChainClient) BeginBlock() {
	dymintC.baseTestChainClient.BeginBlock()
}
func (dymintC *DymintTestChainClient) UpdateCurrentHeaderTime(t time.Time) {
	dymintC.baseTestChainClient.UpdateCurrentHeaderTime(t)
}
func (dymintC *DymintTestChainClient) ClientConfigToState(ClientConfig ibctesting.ClientConfig) exported.ClientState {
	return dymintC.baseTestChainClient.ClientConfigToState(ClientConfig)
}
func (dymintC *DymintTestChainClient) GetConsensusState() exported.ConsensusState {
	return dymintC.baseTestChainClient.GetConsensusState()
}
func (dymintC *DymintTestChainClient) NewConfig() ibctesting.ClientConfig {
	return dymintC.baseTestChainClient.NewConfig()
}
func (dymintC *DymintTestChainClient) GetSelfClientType() string {
	return dymintC.baseTestChainClient.GetSelfClientType()
}
func (dymintC *DymintTestChainClient) GetLastHeader() interface{} {
	return dymintC.baseTestChainClient.GetLastHeader()
}
