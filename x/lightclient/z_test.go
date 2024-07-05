package lightclient

import (
	"testing"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientkeeper "github.com/cosmos/ibc-go/v6/modules/core/02-client/keeper"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/cosmos/ibc-go/v6/testing/simapp"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmprotoversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/types"
	tmversion "github.com/tendermint/tendermint/version"
	"pgregory.net/rapid"
)

const (
	chainID                      = "chainid"
	clientID                     = "clientid"
	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

var upgradePath = []string{"upgrade", "upgradedIBCState"}

type CreateClient struct {
	clientState    exported.ClientState
	consensusState exported.ConsensusState
}

func makeCreateClient() CreateClient {
	t := time.Now()
	root := commitmenttypes.NewMerkleRoot([]byte("hash"))
	valset := types.ValidatorSet{
		Validators: nil,
		Proposer:   nil,
	}
	valHash := valset.Hash()
	consState := ibctmtypes.NewConsensusState(t, root, valHash)
	clientState := ibctmtypes.NewClientState(
		chainID,
		ibctmtypes.DefaultTrustLevel,
		trustingPeriod,
		ubdPeriod,
		maxClockDrift,
		clienttypes.NewHeight(0, 42),
		commitmenttypes.GetSDKSpecs(),
		upgradePath, false, false,
	)
	return CreateClient{consensusState: consState, clientState: clientState}
}

type UpdateClient struct {
	clientID string
	header   exported.Header
}

func getTMHeader() *ibctmtypes.Header {
	var blockHeight int64
	var timestamp time.Time
	var commitID storetypes.CommitID
	var signers map[string]types.PrivValidator
	var tmValSet *types.ValidatorSet
	var nextVals *types.ValidatorSet
	var tmTrustedVals *types.ValidatorSet
	var appHeader tmproto.Header

	vsetHash := tmValSet.Hash()
	nextValHash := nextVals.Hash()

	tmHeader := types.Header{
		Version:            tmprotoversion.Consensus{Block: tmversion.BlockProtocol, App: 2},
		ChainID:            chainID,
		Height:             blockHeight,
		Time:               timestamp,
		LastBlockID:        ibctesting.MakeBlockID(make([]byte, tmhash.Size), 10_000, make([]byte, tmhash.Size)),
		LastCommitHash:     commitID.Hash,
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: nextValHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            appHeader.AppHash,
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    tmValSet.Proposer.Address, //nolint:staticcheck
	}

	hhash := tmHeader.Hash()
	blockID := ibctesting.MakeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))
	voteSet := types.NewVoteSet(chainID, blockHeight, 1, tmproto.PrecommitType, tmValSet)

	// MakeCommit expects a signer array in the same order as the validator array.
	// Thus we iterate over the ordered validator set and construct a signer array
	// from the signer map in the same order.
	var signerArr []types.PrivValidator     //nolint:prealloc // using prealloc here would be needlessly complex
	for _, v := range tmValSet.Validators { //nolint:staticcheck // need to check for nil validator set
		signerArr = append(signerArr, signers[v.Address.String()])
	}
	commit, err := types.MakeCommit(blockID, blockHeight, 1, voteSet, signerArr, timestamp)

	signedHeader := &tmproto.SignedHeader{
		Header: tmHeader.ToProto(),
		Commit: commit.ToProto(),
	}
	valset := types.ValidatorSet{
		Validators: nil,
		Proposer:   nil,
	}
	trustedHeight := clienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: 0,
	}
	trustedValidator := types.ValidatorSet{
		Validators: nil,
		Proposer:   nil,
	}

	valsetProto, _ := valset.ToProto()
	trustedValidatorProto, _ := trustedValidator.ToProto()
	return &ibctmtypes.Header{
		ValidatorSet:      valsetProto,
		SignedHeader:      signedHeader,
		TrustedHeight:     trustedHeight,
		TrustedValidators: trustedValidatorProto,
	}
}

func makeUpdateClient() UpdateClient {
	header := getTMHeader()
	return UpdateClient{clientID: clientID, header: header}
}

/*
TODO: what am I testing?
I just want to see if the whole thing basically works?
*/
type Model struct {
	app *simapp.SimApp
}

func (m *Model) ctx() sdk.Context {
	ctx := m.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: chainID, Time: time.Now()})
	return ctx
}

func (m *Model) createClient(x CreateClient) {
	m.clientKeeper().CreateClient(m.ctx(), x.clientState, x.consensusState)
}

func (m *Model) updateClient(x UpdateClient) {
	m.clientKeeper().UpdateClient(m.ctx(), x.clientID, x.header)
}

func (m *Model) clientKeeper() clientkeeper.Keeper {
	return m.app.IBCKeeper.ClientKeeper
}

func TestFoo(t *testing.T) {
	app := simapp.Setup(false)
	m := &Model{app: app}
	_ = m
}

// go test ./x/lightclient/... -v -run=TestRapid -rapid.checks=10000 -rapid.steps=50
func TestRapid(t *testing.T) {
	f := func(r *rapid.T) {
		ops := map[string]func(*rapid.T){
			"a": func(t *rapid.T) {
			},
			"b": func(t *rapid.T) {
			},
		}
		r.Repeat(ops)
	}

	rapid.Check(t, f)
}
