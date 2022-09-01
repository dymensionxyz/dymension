package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/dymensionxyz/dymension/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/x/sequencer/types"
	"github.com/dymensionxyz/dymension/x/simulation"
	simulationtypes "github.com/dymensionxyz/dymension/x/simulation/types"
)

func SimulateMsgCreateSequencer(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// choose creator and rollappId
		creatorAccount, _ := simtypes.RandomAcc(r, accs)
		seqAccount, _ := simtypes.RandomAcc(r, accs)
		seqAddress := seqAccount.Address.String()

		// choose rollappID and whether or not to fail the transaction
		rollappId := "NoSuchRollapp"
		rollappIndex := -1
		bFailNoRollapp := r.Int()%5 == 0 || len(simulation.GlobalRollappList) == 0
		var rollapp simulationtypes.SimRollapp
		if !bFailNoRollapp {
			rollapp, rollappIndex = simulation.RandomRollapp(r, simulation.GlobalRollappList)
			rollappId = rollapp.RollappId
		}

		msg := &types.MsgCreateSequencer{
			Creator:          creatorAccount.Address.String(),
			SequencerAddress: seqAddress,
			Pubkey:           nil,
			RollappId:        rollappId,
			Description:      types.Description{},
		}

		bNotPermissioned := false
		if !bFailNoRollapp && len(rollapp.PermissionedAddresses) > 0 {
			// check whether or not to fail the transaction because of permissioned sequencer
			bNotPermissioned = true
			for _, item := range rollapp.PermissionedAddresses {
				if item == seqAddress {
					bNotPermissioned = false
					break
				}
			}
		}

		bExpectedError := bFailNoRollapp || bNotPermissioned

		// count how many sequencers already attached to this rollapp
		rollappSeqNum := uint64(0)
		bAlreadyExists := false
		if !bExpectedError {
			for _, item := range simulation.GlobalSequencerAddressesList {
				// check how many sequencers already attached to this rollapp
				if item.RollappIndex == rollappIndex {
					rollappSeqNum += 1
				}
				// check if we already created it
				if item.Account.Address.String() == seqAddress {
					bAlreadyExists = true
				}
			}
		}

		bMaxSequencersFailure := rollapp.MaxSequencers <= rollappSeqNum

		bExpectedError = bExpectedError || bAlreadyExists || bMaxSequencersFailure

		if !bExpectedError {
			sequencer := simulationtypes.SimSequencer{
				Account:      seqAccount,
				Creator:      msg.Creator,
				RollappIndex: rollappIndex,
			}
			simulation.GlobalSequencerAddressesList = append(simulation.GlobalSequencerAddressesList, sequencer)
			simulation.GlobalRollappList[rollappIndex].Sequencers = append(rollapp.Sequencers, len(simulation.GlobalSequencerAddressesList)-1)
		}

		return simulation.GenAndDeliverMsgWithRandFees(msg, msg.Type(), types.ModuleName, r, app, &ctx, &creatorAccount, bk, ak, nil, bExpectedError)

	}
}
