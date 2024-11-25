package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/dymensionxyz/dymension/v3/simulation"
	simulationtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func SimulateMsgCreateRollapp(ak simulationtypes.AccountKeeper, bk simulationtypes.BankKeeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// choose creator and rollappId
		simAccount, rollappNumId := simtypes.RandomAcc(r, accs)
		rollappId := fmt.Sprintf("rollapp_%d-1", rollappNumId)

		// check if we already created it
		bAlreadyExists := false
		for _, item := range simulation.GlobalRollappList {
			if item.RollappId == rollappId {
				bAlreadyExists = true
			}
		}

		// fund the creator with coins to cover the rollapp alias fee
		rollappAliasFee := sdk.NewCoins(sdk.NewCoin("adym", commontypes.DYM.MulRaw(5)))
		err := bk.MintCoins(ctx, minttypes.ModuleName, rollappAliasFee)
		if err != nil {
			return simtypes.OperationMsg{}, nil, fmt.Errorf("SimulateMsgCreateRollapp: MintCoins: %w", err)
		}
		err = bk.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, simAccount.Address, rollappAliasFee)
		if err != nil {
			return simtypes.OperationMsg{}, nil, fmt.Errorf("SimulateMsgCreateRollapp: SendCoinsFromModuleToAccount: %w", err)
		}

		msg := &types.MsgCreateRollapp{
			Creator:          simAccount.Address.String(),
			RollappId:        rollappId,
			InitialSequencer: sample.AccAddress(),
			Alias:            "rollapplongalias",
			Metadata: &types.RollappMetadata{
				Website:     "https://dymension.xyz",
				Description: "Sample description",
				LogoUrl:     "https://dymension.xyz/logo.png",
				Telegram:    "https://t.me/rolly",
				X:           "https://x.dymension.xyz",
				GenesisUrl:  "https://genesis-dymension.xyz",
				DisplayName: "RollApp",
				Tagline:     "",
				ExplorerUrl: "https://dymension-explorer.xyz",
				FeeDenom: &types.DenomMetadata{
					Display:  "DYMLONG",
					Base:     "udymlong",
					Exponent: 18,
				},
			},
			GenesisInfo: &types.GenesisInfo{
				GenesisChecksum: "checksum",
				Bech32Prefix:    "bech",
				NativeDenom: types.DenomMetadata{
					Display:  "DYMLONG",
					Base:     "udymlong",
					Exponent: 18,
				},
				InitialSupply: sdk.NewInt(1000000000),
			},
			VmType: types.Rollapp_EVM,
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
