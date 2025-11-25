package simulation_test

import (
	"fmt"
	"time"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	usim "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/dymensionxyz/dymension/v3/app"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Helper function to marshal a module's genesis state and assign it to the overall genesis map.
func marshalAndSetGenesis(cdc codec.JSONCodec, genesis app.GenesisState, moduleName string, moduleGenesis interface{}) error {
	rawGenesis, err := cdc.MarshalJSON(moduleGenesis)
	if err != nil {
		return fmt.Errorf("failed to marshal %s genesis state: %w", moduleName, err)
	}
	genesis[moduleName] = rawGenesis
	return nil
}

// prepareGenesis sets up a modified GenesisState for simulation purposes.
// It initializes default parameters and overrides them for specific simulation requirements.
func prepareGenesis(cdc codec.JSONCodec) (app.GenesisState, error) {
	newApp := app.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, usim.EmptyAppOptions{}, baseapp.SetChainID(SimulationAppChainID))
	genesis := newApp.DefaultGenesis()
	var err error

	// --- Government (Gov) Params ---
	govGenesis := govtypesv1.DefaultGenesisState()
	// Set a non-zero minimum deposit and a short voting period for faster simulation results.
	govGenesis.Params.MinDeposit[0].Amount = math.NewInt(10000000000)
	govGenesis.Params.MinDeposit[0].Denom = "adym"
	govVotingPeriod := time.Minute
	govGenesis.Params.VotingPeriod = &govVotingPeriod
	if err = marshalAndSetGenesis(cdc, genesis, govtypes.ModuleName, govGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Rollapp Params ---
	rollappGenesis := rollapptypes.DefaultGenesis()
	// Set a short dispute period for simulation
	rollappGenesis.Params.DisputePeriodInBlocks = 50
	if err = marshalAndSetGenesis(cdc, genesis, rollapptypes.ModuleName, rollappGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Sequencer Params ---
	sequencerGenesis := sequencertypes.DefaultGenesis()
	// Set a short notice period
	sequencerGenesis.Params.NoticePeriod = time.Minute
	if err = marshalAndSetGenesis(cdc, genesis, sequencertypes.ModuleName, sequencerGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Auth Params ---
	authGenesis := auth.DefaultGenesisState()
	// Adjust the transaction size cost
	authGenesis.Params.TxSizeCostPerByte = 100
	if err = marshalAndSetGenesis(cdc, genesis, auth.ModuleName, authGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Slashing Params ---
	slashingGenesis := slashing.DefaultGenesisState()
	slashingGenesis.Params.SignedBlocksWindow = 10000
	slashingGenesis.Params.MinSignedPerWindow = math.LegacyMustNewDecFromStr("0.800000000000000000")
	slashingGenesis.Params.DowntimeJailDuration = 2 * time.Minute
	// Disable downtime slash to keep validators active during simulation
	slashingGenesis.Params.SlashFractionDowntime = math.LegacyZeroDec()
	if err = marshalAndSetGenesis(cdc, genesis, slashing.ModuleName, slashingGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Staking Params ---
	stakingGenesis := stakingtypes.DefaultGenesisState()
	stakingGenesis.Params.BondDenom = "adym"
	if err = marshalAndSetGenesis(cdc, genesis, stakingtypes.ModuleName, stakingGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Mint Params ---
	mintGenesis := minttypes.DefaultGenesisState()
	mintGenesis.Params.MintDenom = "adym"
	if err = marshalAndSetGenesis(cdc, genesis, minttypes.ModuleName, mintGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- EVM Params (Ethermint) ---
	evmGenesis := evmtypes.DefaultGenesisState()
	evmGenesis.Params.EvmDenom = "adym"
	// Disable contract creation for simplicity in simulation
	evmGenesis.Params.EnableCreate = false
	if err = marshalAndSetGenesis(cdc, genesis, evmtypes.ModuleName, evmGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Feemarket Params (EIP-1559) ---
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	// Disable base fee enforcement for simpler transaction processing
	feemarketGenesis.Params.NoBaseFee = true
	if err = marshalAndSetGenesis(cdc, genesis, feemarkettypes.ModuleName, feemarketGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Dymns Params (Dymension Name Service) ---
	dymnsGenesis := dymnstypes.DefaultGenesis()
	// Set a short duration for sell orders
	dymnsGenesis.Params.Misc.SellOrderDuration = 2 * time.Minute
	if err = marshalAndSetGenesis(cdc, genesis, dymnstypes.ModuleName, dymnsGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Bank Params (Denom Metadata) ---
	bankGenesis := banktypes.DefaultGenesisState()
	// Define metadata for the base asset (adym/DYM)
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
	if err = marshalAndSetGenesis(cdc, genesis, banktypes.ModuleName, bankGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Crisis Params ---
	crisisGenesis := crisistypes.DefaultGenesisState()
	crisisGenesis.ConstantFee.Denom = "adym"
	if err = marshalAndSetGenesis(cdc, genesis, crisistypes.ModuleName, crisisGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Txfees Params (Osmosis) ---
	txfeesGenesis := txfeestypes.DefaultGenesis()
	txfeesGenesis.Basedenom = "adym"
	txfeesGenesis.Params.EpochIdentifier = "minute"
	if err = marshalAndSetGenesis(cdc, genesis, txfeestypes.ModuleName, txfeesGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Gamm Params (Osmosis Liquidity Pools) ---
	gammGenesis := gammtypes.DefaultGenesis()
	gammGenesis.Params.PoolCreationFee[0].Denom = "adym"
	gammGenesis.Params.EnableGlobalPoolFees = true
	if err = marshalAndSetGenesis(cdc, genesis, gammtypes.ModuleName, gammGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Incentives Params (Osmosis) ---
	incentivesGenesis := incentivestypes.DefaultGenesis()
	incentivesGenesis.Params.DistrEpochIdentifier = "minute"
	// Set a short lock duration for incentives simulation
	incentivesGenesis.LockableDurations = []time.Duration{time.Minute}
	if err = marshalAndSetGenesis(cdc, genesis, incentivestypes.ModuleName, incentivesGenesis); err != nil {
		return app.GenesisState{}, err
	}

	// --- Epochs Params (Osmosis) ---
	epochsGenesis := epochstypes.DefaultGenesis()
	// Define the "minute" epoch for use by incentives/txfees
	epochsGenesis.Epochs = append(epochsGenesis.Epochs, epochstypes.EpochInfo{
		Identifier:              "minute",
		StartTime:               time.Time{},
		Duration:                time.Minute,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
	})
	if err = marshalAndSetGenesis(cdc, genesis, epochstypes.ModuleName, epochsGenesis); err != nil {
		return app.GenesisState{}, err
	}

	return genesis, nil
}
