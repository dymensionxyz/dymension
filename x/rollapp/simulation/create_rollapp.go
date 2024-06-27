package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/dymensionxyz/dymension/v3/simulation"
	simulationtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func SimulateMsgCreateRollapp(ak simulationtypes.AccountKeeper, bk simulationtypes.BankKeeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// choose creator and rollappId
		simAccount, rollappNumId := simtypes.RandomAcc(r, accs)
		rollappId := "rollapp" + fmt.Sprint(rollappNumId)

		// check if we already created it
		bAlreadyExists := false
		for _, item := range simulation.GlobalRollappList {
			if item.RollappId == rollappId {
				bAlreadyExists = true
			}
		}

		permissionedAddresses := []string{}
		bPermissioned := r.Int()%2 == 0
		bFailDuplicateSequencer := false
		if bPermissioned {
			for i := 0; i < r.Intn(len(accs)); i++ {
				seqAccount, _ := simtypes.RandomAcc(r, accs)
				for _, item := range permissionedAddresses {
					if item == seqAccount.Address.String() {
						bFailDuplicateSequencer = true
					}
				}
				permissionedAddresses = append(permissionedAddresses, seqAccount.Address.String())
			}
		}

		// calculate maxSequencers and whether or not to fail the transaction
		bFailMaxSequencers := r.Int()%2 == 0
		maxSequencers := uint64(r.Intn(100)) + 1
		if bFailMaxSequencers {
			maxSequencers = 0
		}

		msg := &types.MsgCreateRollapp{
			Creator:               simAccount.Address.String(),
			RollappId:             rollappId,
			MaxSequencers:         maxSequencers,
			PermissionedAddresses: permissionedAddresses,
		}

		bExpectedError := bFailMaxSequencers || bFailDuplicateSequencer || bAlreadyExists

		if !bExpectedError {
			simulation.GlobalRollappList = append(simulation.GlobalRollappList, simulationtypes.SimRollapp{
				RollappId:             rollappId,
				MaxSequencers:         maxSequencers,
				PermissionedAddresses: permissionedAddresses,
				Sequencers:            []int{},
				LastHeight:            0,
				LastCreationHeight:    0,
			})
		}

		return simulation.GenAndDeliverMsgWithRandFees(msg, msg.Type(), types.ModuleName, r, app, &ctx, &simAccount, bk, ak, nil, bExpectedError)
	}
}
