package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/dymensionxyz/dymension/v3/simulation"
	simulationtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// TODO: revive simulation tests!
func SimulateMsgCreateSequencer(ak simulationtypes.AccountKeeper, bk simulationtypes.BankKeeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// choose creator and rollappId
		creatorAccount, _ := simtypes.RandomAcc(r, accs)
		seqAddress := creatorAccount.Address.String()
		pkAny, err := codectypes.NewAnyWithValue(creatorAccount.PubKey)
		if err != nil {
			panic(err)
		}

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
			Creator:      seqAddress,
			DymintPubKey: pkAny,
			RollappId:    rollappId,
			Metadata: types.SequencerMetadata{
				Moniker:        "sequencer",
				Details:        "Sample description",
				P2PSeeds:       nil,
				Rpcs:           []string{"https://rpc-dym.com:443"},
				EvmRpcs:        []string{"https://rpc-dym-evm.com:443"},
				RestApiUrls:    []string{"https://rest-dym.com:443"},
				ExplorerUrl:    "https://dymension-explorer.xyz",
				GenesisUrls:    []string{"https://genesis-dym.com"},
				ContactDetails: nil,
				ExtraData:      nil,
				Snapshots:      nil,
				GasPrice:       nil,
			},
			Bond:                sdk.NewCoin("adym", commontypes.DYM.MulRaw(100)),
			RewardAddr:          seqAddress,
			WhitelistedRelayers: []string{},
		}

		bExpectedError := bFailNoRollapp

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

		bExpectedError = bExpectedError || bAlreadyExists

		if !bExpectedError {
			sequencer := simulationtypes.SimSequencer{
				Account:      creatorAccount,
				Creator:      msg.Creator,
				RollappIndex: rollappIndex,
			}
			simulation.GlobalSequencerAddressesList = append(simulation.GlobalSequencerAddressesList, sequencer)
			simulation.GlobalRollappList[rollappIndex].Sequencers = append(rollapp.Sequencers, len(simulation.GlobalSequencerAddressesList)-1)
		}

		return simulation.GenAndDeliverMsgWithRandFees(msg, msg.Type(), types.ModuleName, r, app, &ctx, &creatorAccount, bk, ak, nil, bExpectedError)
	}
}
