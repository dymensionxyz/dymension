package apptesting

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/rand"

	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"

	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var (
	alice = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
	bond  = sequencertypes.DefaultParams().MinBond
)

type KeeperTestHelper struct {
	suite.Suite
	App *app.App
	Ctx sdk.Context
}

func (suite *KeeperTestHelper) CreateDefaultRollapp() string {
	return suite.CreateRollappWithName(rand.Str(8))
}

func (suite *KeeperTestHelper) CreateRollappWithName(name string) string {
	rollapp := rollapptypes.Rollapp{
		RollappId:     name,
		Creator:       alice,
		Version:       0,
		MaxSequencers: 5,
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)
	return rollapp.GetRollappId()
}

func (suite *KeeperTestHelper) CreateDefaultSequencer(ctx sdk.Context, rollappId string) string {
	pubkey1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pubkey1.Address())
	pkAny1, err := codectypes.NewAnyWithValue(pubkey1)
	suite.Require().Nil(err)

	// fund account
	err = bankutil.FundAccount(suite.App.BankKeeper, ctx, addr1, sdk.NewCoins(bond))
	suite.Require().Nil(err)

	sequencerMsg1 := sequencertypes.MsgCreateSequencer{
		Creator:      addr1.String(),
		DymintPubKey: pkAny1,
		Bond:         bond,
		RollappId:    rollappId,
		Description:  sequencertypes.Description{},
	}

	msgServer := sequencerkeeper.NewMsgServerImpl(suite.App.SequencerKeeper)
	_, err = msgServer.CreateSequencer(ctx, &sequencerMsg1)
	suite.Require().Nil(err)
	return addr1.String()
}

func (suite *KeeperTestHelper) PostStateUpdate(ctx sdk.Context, rollappId, seqAddr string, startHeight, numOfBlocks uint64) (lastHeight uint64, err error) {
	var bds rollapptypes.BlockDescriptors
	bds.BD = make([]rollapptypes.BlockDescriptor, numOfBlocks)
	for k := 0; k < int(numOfBlocks); k++ {
		bds.BD[k] = rollapptypes.BlockDescriptor{Height: startHeight + uint64(k)}
	}

	updateState := rollapptypes.MsgUpdateState{
		Creator:     seqAddr,
		RollappId:   rollappId,
		StartHeight: startHeight,
		NumBlocks:   numOfBlocks,
		DAPath:      "",
		Version:     0,
		BDs:         bds,
	}
	msgServer := rollappkeeper.NewMsgServerImpl(suite.App.RollappKeeper)
	_, err = msgServer.UpdateState(ctx, &updateState)
	return startHeight + numOfBlocks, err
}
