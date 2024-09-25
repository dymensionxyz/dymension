package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/dymensionxyz/dymension/v3/simulation"
	simulationtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
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

		msg := &types.MsgCreateRollapp{
			Creator:          simAccount.Address.String(),
			RollappId:        rollappId,
			InitialSequencer: sample.AccAddress(),
			Alias:            "rollapp",
			Metadata: &types.RollappMetadata{
				Website:     "https://dymension.xyz",
				Description: "Sample description",
				LogoUrl:     "https://dymension.xyz/logo.png",
				Telegram:    "https://t.me/rolly",
				X:           "https://x.dymension.xyz",
			},
			GenesisInfo: &types.GenesisInfo{
				GenesisChecksum: "checksum",
				Bech32Prefix:    "bech",
				NativeDenom: types.DenomMetadata{
					Display:  "DYM",
					Base:     "udym",
					Exponent: 6,
				},
				InitialSupply: sdk.NewInt(1000000000),
			},
		}

		if !bAlreadyExists {
			simulation.GlobalRollappList = append(simulation.GlobalRollappList, simulationtypes.SimRollapp{
				RollappId:          rollappId,
				Sequencers:         []int{},
				LastHeight:         0,
				LastCreationHeight: 0,
			})
		}

		return simulation.GenAndDeliverMsgWithRandFees(msg, msg.Type(), types.ModuleName, r, app, &ctx, &simAccount, bk, ak, nil, bAlreadyExists)
	}
}
