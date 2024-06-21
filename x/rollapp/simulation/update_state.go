package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/dymensionxyz/dymension/v3/simulation"
	simulationtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func SimulateMsgUpdateState(
	ak simulationtypes.AccountKeeper,
	bk simulationtypes.BankKeeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if len(simulation.GlobalSequencerAddressesList) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateState, "No sequencers"), nil, nil
		}
		sequencer, sequencerIndex := simulation.RandomSequencer(r, simulation.GlobalSequencerAddressesList)

		rollappIndex := sequencer.RollappIndex
		rollapp := simulation.GlobalRollappList[rollappIndex]
		bNotActive := rollapp.Sequencers[0] != sequencerIndex

		bStateWasUpdatedInThisHeight := false
		if rollapp.LastCreationHeight == uint64(ctx.BlockHeight()) {
			bStateWasUpdatedInThisHeight = true
		}

		// decide whether or not to send to wrong rollapp
		bWrongRollapp := r.Int()%5 == 0
		if bWrongRollapp {
			rollapp, rollappIndex = simulation.RandomRollapp(r, simulation.GlobalRollappList)
			if sequencer.RollappIndex == rollappIndex {
				bWrongRollapp = false
			}
		}

		// calc numBlocks
		numBlocks := uint64(r.Intn(2000))
		bNoBds := numBlocks == 0

		// decide start height
		bWrongStartHeight := r.Int()%5 == 0
		startHeight := rollapp.LastHeight + 1
		if bWrongStartHeight {
			randStartheight := uint64(r.Intn(20000000))
			if startHeight == randStartheight {
				startHeight -= 1
			} else {
				startHeight = randStartheight
			}
		}

		bds := types.BlockDescriptors{}
		for i := uint64(0); i < numBlocks; i++ {
			bds.BD = append(bds.BD, types.BlockDescriptor{
				Height:    startHeight + i,
				StateRoot: make([]byte, 32),
			})
		}

		// create message
		msg := &types.MsgUpdateState{
			Creator:     sequencer.Account.Address.String(),
			RollappId:   rollapp.RollappId,
			StartHeight: startHeight,
			NumBlocks:   numBlocks,
			DAPath:      "",
			Version:     0,
			BDs:         bds,
		}

		bExpectedError := bNotActive || bWrongRollapp || bNoBds || bWrongStartHeight || bStateWasUpdatedInThisHeight

		// update rollapp
		if !bExpectedError {
			simulation.GlobalRollappList[rollappIndex].LastHeight += numBlocks
			simulation.GlobalRollappList[rollappIndex].LastCreationHeight = uint64(ctx.BlockHeight())
		}

		return simulation.GenAndDeliverMsgWithRandFees(msg, msg.Type(), types.ModuleName, r, app, &ctx, &sequencer.Account, bk, ak, nil, bExpectedError)
	}
}
