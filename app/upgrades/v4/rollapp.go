package v4

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func migrateRollapps(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper) error {
	list := rollappkeeper.GetAllRollapps(ctx)
	for _, oldRollapp := range list {
		newRollapp := ConvertOldRollappToNew(oldRollapp)
		if err := newRollapp.ValidateBasic(); err != nil {
			return err
		}
		rollappkeeper.SetRollapp(ctx, newRollapp)
	}
	return nil
}

func ConvertOldRollappToNew(oldRollapp rollapptypes.Rollapp) rollapptypes.Rollapp {
	return rollapptypes.Rollapp{
		RollappId:    oldRollapp.RollappId,
		Owner:        oldRollapp.Owner,
		GenesisState: oldRollapp.GenesisState,
		ChannelId:    oldRollapp.ChannelId,
		Frozen:       oldRollapp.Frozen,
		Metadata: &rollapptypes.RollappMetadata{
			Website:     "",
			Description: "",
			LogoUrl:     "",
			Telegram:    "",
			X:           "",
			GenesisUrl:  "",
			DisplayName: "",
			Tagline:     "",
			ExplorerUrl: "",
			FeeDenom: &rollapptypes.DenomMetadata{
				Display:  "",
				Base:     "",
				Exponent: 0,
			},
		},
		GenesisInfo: rollapptypes.GenesisInfo{
			GenesisChecksum: "",
			Bech32Prefix:    "",
			NativeDenom: rollapptypes.DenomMetadata{
				Display:  "",
				Base:     "",
				Exponent: 0,
			},
			InitialSupply: sdk.Int{},
			Sealed:        true,
			GenesisAccounts: &rollapptypes.GenesisAccounts{
				Accounts: []rollapptypes.GenesisAccount{
					{
						Amount:  sdk.Int{},
						Address: "",
					},
				},
			},
		},
		InitialSequencer:      "",
		VmType:                0,
		Launched:              true,
		PreLaunchTime:         &time.Time{},
		LivenessEventHeight:   0,
		LastStateUpdateHeight: 0,
	}
}
