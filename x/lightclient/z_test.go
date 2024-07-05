package lightclient

import (
	"testing"
	"time"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/cosmos/ibc-go/v6/testing/simapp"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	"pgregory.net/rapid"
)

func getTMHeader() *ibctmtypes.Header {
	valset := types.ValidatorSet{
		Validators: nil,
		Proposer:   nil,
	}
	return &ibctmtypes.Header{}
}

type CreateClient struct {
	clientState    exported.ClientState
	consensusState exported.ConsensusState
}

const (
	chainID                        = "gaia"
	chainIDRevision0               = "gaia-revision-0"
	chainIDRevision1               = "gaia-revision-1"
	clientID                       = "gaiamainnet"
	trustingPeriod   time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod        time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift    time.Duration = time.Second * 10
)

var upgradePath = []string{"upgrade", "upgradedIBCState"}

func getCreateClient() CreateClient {
	t := time.Now()
	root := commitmenttypes.NewMerkleRoot([]byte("hash"))
	valset := types.ValidatorSet{
		Validators: nil,
		Proposer:   nil,
	}
	valHash := valset.Hash()
	consState := ibctmtypes.NewConsensusState(t, root, valHash)
	clientState := ibctmtypes.NewClientState(
		"chainid",
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

func getUpdateClient() UpdateClient {
	clientID := "clientid"
	header := getTMHeader()
	return UpdateClient{clientID: clientID, header: header}
}

func TestFoo(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "chainid", Time: time.Now()})
	k := app.IBCKeeper.ClientKeeper

	k.CreateClient(ctx)

	k.UpdateClient()
}

/*
TODO: what am I testing?
I just want to see if the whole thing basically works?
*/
type Model struct{}

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
