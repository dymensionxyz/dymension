package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	sharedtypes "github.com/dymensionxyz/dymension/shared/types"
	"github.com/dymensionxyz/dymension/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

func SimulateMsgCreateRollapp(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// choose creator and rollappId
		simAccount, rollappNumId := simtypes.RandomAcc(r, accs)
		rollappId := "rollapp" + fmt.Sprint(rollappNumId)

		// check if we already created it
		bAlreadyExists := false
		for _, item := range globalRollappIdList {
			if item == rollappId {
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
		// calculate maxWithholdingBlocks and whether or not to fail the transaction
		bFailMaxWithholdingBlocks := r.Int()%2 == 0
		maxWithholdingBlocks := uint64(r.Intn(len(accs))) + 1
		if bFailMaxWithholdingBlocks {
			maxWithholdingBlocks = 0
		}
		msg := &types.MsgCreateRollapp{
			Creator:              simAccount.Address.String(),
			RollappId:            rollappId,
			CodeStamp:            "",
			GenesisPath:          "",
			MaxWithholdingBlocks: maxWithholdingBlocks,
			MaxSequencers:        maxSequencers,
			PermissionedAddresses: sharedtypes.Sequencers{
				Addresses: permissionedAddresses,
			},
		}

		// fmt.Printf("SimulateMsgCreateRollapp: RollappId(%s) bFailMaxSequencers(%t) bFailMaxWithholdingBlocks(%t) bFailDuplicateSequencer(%t)\n",
		// 	msg.RollappId, bFailMaxSequencers, bFailMaxWithholdingBlocks, bFailDuplicateSequencer)
		bExpectedError := bFailMaxSequencers || bFailMaxWithholdingBlocks || bFailDuplicateSequencer || bAlreadyExists

		if !bExpectedError {
			globalRollappIdList = append(globalRollappIdList, msg.RollappId)
		}

		return GenAndDeliverMsgWithRandFees(msg, msg.Type(), r, app, &ctx, &simAccount, &bk, &ak, nil, bExpectedError)
	}
}
