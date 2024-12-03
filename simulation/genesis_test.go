package simulation_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govtypes2 "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypes1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	"github.com/dymensionxyz/dymension/v3/app"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func prepareGenesis(cdc codec.JSONCodec) (app.GenesisState, error) {
	genesis := app.NewDefaultGenesisState(cdc)

	// Modify gov params
	govGenesis := govtypes1.DefaultGenesisState()
	govGenesis.Params.MinDeposit[0].Amount = math.NewInt(10000000000)
	govGenesis.Params.MinDeposit[0].Denom = "adym"
	govVotingPeriod := time.Minute
	govGenesis.Params.VotingPeriod = &govVotingPeriod
	govRawGenesis, err := cdc.MarshalJSON(govGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal gov genesis state: %w", err)
	}
	genesis[govtypes2.ModuleName] = govRawGenesis

	// Modify rollapp params
	rollappGenesis := rollapptypes.DefaultGenesis()
	rollappGenesis.Params.DisputePeriodInBlocks = 50
	rollappRawGenesis, err := cdc.MarshalJSON(rollappGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal rollapp genesis state: %w", err)
	}
	genesis[rollapptypes.ModuleName] = rollappRawGenesis

	// Modify sequencer params
	sequencerGenesis := sequencertypes.DefaultGenesis()
	sequencerGenesis.Params.NoticePeriod = time.Minute
	sequencerRawGenesis, err := cdc.MarshalJSON(sequencerGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal sequencer genesis state: %w", err)
	}
	genesis[sequencertypes.ModuleName] = sequencerRawGenesis

	// Modify auth params
	authGenesis := auth.DefaultGenesisState()
	authGenesis.Params.TxSizeCostPerByte = 100
	authRawGenesis, err := cdc.MarshalJSON(authGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}
	genesis[auth.ModuleName] = authRawGenesis

	// Modify slashing params
	slashingGenesis := slashing.DefaultGenesisState()
	slashingGenesis.Params.SignedBlocksWindow = 10000
	slashingGenesis.Params.MinSignedPerWindow = math.LegacyMustNewDecFromStr("0.800000000000000000")
	slashingGenesis.Params.DowntimeJailDuration = 2 * time.Minute
	slashingGenesis.Params.SlashFractionDowntime = math.LegacyZeroDec()
	slashingRawGenesis, err := cdc.MarshalJSON(slashingGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal slashing genesis state: %w", err)
	}
	genesis[slashing.ModuleName] = slashingRawGenesis

	// Modify staking params
	stakingGenesis := stakingtypes.DefaultGenesisState()
	stakingGenesis.Params.BondDenom = "adym"
	stakingRawGenesis, err := cdc.MarshalJSON(stakingGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal staking genesis state: %w", err)
	}
	genesis[stakingtypes.ModuleName] = stakingRawGenesis

	// Modify mint params
	mintGenesis := minttypes.DefaultGenesisState()
	mintGenesis.Params.MintDenom = "adym"
	mintRawGenesis, err := cdc.MarshalJSON(mintGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal mint genesis state: %w", err)
	}
	genesis[minttypes.ModuleName] = mintRawGenesis

	// Modify evm params
	evmGenesis := evmtypes.DefaultGenesisState()
	evmGenesis.Params.EvmDenom = "adym"
	evmGenesis.Params.EnableCreate = false
	evmRawGenesis, err := cdc.MarshalJSON(evmGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal evm genesis state: %w", err)
	}
	genesis[evmtypes.ModuleName] = evmRawGenesis

	// Modify feemarket params
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.NoBaseFee = true
	feemarketRawGenesis, err := cdc.MarshalJSON(feemarketGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal feemarket genesis state: %w", err)
	}
	genesis[feemarkettypes.ModuleName] = feemarketRawGenesis

	// Modify dymns params
	dymnsGenesis := dymnstypes.DefaultGenesis()
	dymnsGenesis.Params.Misc.SellOrderDuration = 2 * time.Minute
	dymnsRawGenesis, err := cdc.MarshalJSON(dymnsGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal dymns genesis state: %w", err)
	}
	genesis[dymnstypes.ModuleName] = dymnsRawGenesis

	// Modify bank denom metadata
	bankGenesis := banktypes.DefaultGenesisState()
	bankGenesis.DenomMetadata = []banktypes.Metadata{
		{
			Base: "adym",
			DenomUnits: []*banktypes.DenomUnit{
				{Denom: "adym", Exponent: 0},
				{Denom: "DYM", Exponent: 18},
			},
			Description: "Denom metadata for DYM (adym)",
			Display:     "DYM",
			Name:        "DYM",
			Symbol:      "DYM",
		},
	}
	bankRawGenesis, err := cdc.MarshalJSON(bankGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}
	genesis[banktypes.ModuleName] = bankRawGenesis

	// Modify misc params
	crisisGenesis := crisistypes.DefaultGenesisState()
	crisisGenesis.ConstantFee.Denom = "adym"
	crisisRawGenesis, err := cdc.MarshalJSON(crisisGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal crisis genesis state: %w", err)
	}
	genesis[crisistypes.ModuleName] = crisisRawGenesis

	txfeesGenesis := txfeestypes.DefaultGenesis()
	txfeesGenesis.Basedenom = "adym"
	txfeesGenesis.Params.EpochIdentifier = "minute"
	txfeesRawGenesis, err := cdc.MarshalJSON(txfeesGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal txfees genesis state: %w", err)
	}
	genesis[txfeestypes.ModuleName] = txfeesRawGenesis

	gammGenesis := gammtypes.DefaultGenesis()
	gammGenesis.Params.PoolCreationFee[0].Denom = "adym"
	gammGenesis.Params.EnableGlobalPoolFees = true
	gammRawGenesis, err := cdc.MarshalJSON(gammGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal gamm genesis state: %w", err)
	}
	genesis[gammtypes.ModuleName] = gammRawGenesis

	// Modify incentives params
	incentivesGenesis := incentivestypes.DefaultGenesis()
	incentivesGenesis.Params.DistrEpochIdentifier = "minute"
	incentivesGenesis.LockableDurations = []time.Duration{time.Minute}
	incentivesRawGenesis, err := cdc.MarshalJSON(incentivesGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal incentives genesis state: %w", err)
	}
	genesis[incentivestypes.ModuleName] = incentivesRawGenesis

	// Modify epochs params
	epochsGenesis := epochstypes.DefaultGenesis()
	epochsGenesis.Epochs = append(epochsGenesis.Epochs, epochstypes.EpochInfo{
		Identifier:              "minute",
		StartTime:               time.Time{},
		Duration:                time.Minute,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
	})
	epochsRawGenesis, err := cdc.MarshalJSON(epochsGenesis)
	if err != nil {
		return app.GenesisState{}, fmt.Errorf("failed to marshal epochs genesis state: %w", err)
	}
	genesis[epochstypes.ModuleName] = epochsRawGenesis

	return genesis, nil
}
