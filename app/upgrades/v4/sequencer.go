package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func migrateSequencers(ctx sdk.Context, sequencerKeeper *sequencerkeeper.Keeper) {
	list := sequencerKeeper.AllSequencers(ctx)
	for _, oldSequencer := range list {
		newSequencer := ConvertOldSequencerToNew(oldSequencer)
		sequencerKeeper.SetSequencer(ctx, newSequencer)

		// TODO: huh?
		if oldSequencer.Proposer {
			sequencerKeeper.SetProposer(ctx, oldSequencer.RollappId, oldSequencer.Address)
			seqs := sequencerKeeper.RollappPotentialProposers(ctx, oldSequencer.RollappId)
			successor, err := sequencerkeeper.ProposerChoiceAlgo(seqs)
			if err != nil {
				continue
			}
			sequencerKeeper.SetSuccessor(ctx, oldSequencer.RollappId, successor.Address)
		}
	}
}

var defaultGasPrice, _ = sdk.NewIntFromString("10000000000")

func ConvertOldSequencerToNew(old sequencertypes.Sequencer) sequencertypes.Sequencer {
	return sequencertypes.Sequencer{
		Address:      old.Address,
		DymintPubKey: old.DymintPubKey,
		RollappId:    old.RollappId,
		Status:       old.Status,
		Tokens:       old.Tokens,
		Metadata: sequencertypes.SequencerMetadata{
			Moniker:     old.Metadata.Moniker,
			Details:     old.Metadata.Details,
			P2PSeeds:    []string{},
			Rpcs:        []string{},
			EvmRpcs:     []string{},
			RestApiUrls: []string{},
			ExplorerUrl: "",
			GenesisUrls: []string{},
			ContactDetails: &sequencertypes.ContactDetails{
				Website:  "",
				Telegram: "",
				X:        "",
			},
			ExtraData: []byte{},
			Snapshots: []*sequencertypes.SnapshotInfo{
				{
					SnapshotUrl: "",
					Height:      0,
					Checksum:    "",
				},
			},
			GasPrice: &defaultGasPrice,
		},
	}
}
