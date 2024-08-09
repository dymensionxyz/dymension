package keeper_test

import (
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUnbondingStatusChange() {
	rollappId, pk1 := suite.CreateDefaultRollapp()

	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk1)
	seqAddrs := make([]string, 2)
	pk2, pk3 := ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey()
	seqAddrs[0] = suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk2)
	seqAddrs[1] = suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk3)
	// sort the  non proposer sequencers by address
	sort.Strings(seqAddrs)
	addr2 := seqAddrs[0]
	addr3 := seqAddrs[1]

	/* ----------------------------- unbond proposer ---------------------------- */
	sequencer, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Require().True(found)
	suite.Equal(types.Bonded, sequencer.Status)
	suite.True(sequencer.Proposer)

	sequencer2, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Require().True(found)
	suite.Equal(types.Bonded, sequencer2.Status)
	suite.False(sequencer2.Proposer)

	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check proposer rotation
	sequencer2, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Equal(types.Bonded, sequencer2.Status)
	suite.True(sequencer2.Proposer)

	// check sequencer operating status
	sequencer, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Equal(types.Unbonding, sequencer.Status)
	suite.False(sequencer.Proposer)

	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer.UnbondTime.Add(10*time.Second))

	sequencer, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Equal(types.Unbonded, sequencer.Status)

	/* ------------------------- unbond non proposer sequencer ------------------------ */
	sequencer3, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr3)
	suite.Require().True(found)
	suite.Equal(types.Bonded, sequencer3.Status)
	suite.False(sequencer3.Proposer)

	unbondMsg = types.MsgUnbond{Creator: addr3}
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check sequencer operating status
	sequencer3, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr3)
	suite.Require().True(found)
	suite.Equal(types.Unbonding, sequencer3.Status)

	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer3.UnbondTime.Add(10*time.Second))

	sequencer3, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr3)
	suite.Require().True(found)
	suite.Equal(types.Unbonded, sequencer3.Status)
}

func (suite *SequencerTestSuite) TestUnbondingNotBondedSequencer() {
	rollappId, pk1 := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk1)

	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// already unbonding, we expect error
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().Error(err)

	sequencer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer.UnbondTime.Add(10*time.Second))

	// already unbonded, we expect error
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().Error(err)
}
