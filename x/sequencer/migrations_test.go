package sequencer_test

import (
	"testing"
	"time"

	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	seq "github.com/dymensionxyz/dymension/v3/x/sequencer"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func TestMigrate2to3(t *testing.T) {
	app := apptesting.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, cometbftproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})

	// create legacy subspace
	testValue := 555555 * time.Second // random value for testing
	params := types.DefaultParams()
	params.UnbondingTime = testValue

	seqSubspace, ok := app.AppKeepers.ParamsKeeper.GetSubspace(types.ModuleName)
	if !ok {
		t.Fatalf("sequencer subspace not found")
	}

	// set KeyTable if it has not already been set
	if !seqSubspace.HasKeyTable() {
		seqSubspace = seqSubspace.WithKeyTable(types.ParamKeyTable())
	}
	seqSubspace.SetParamSet(ctx, &params)

	// Create a migrator instance
	migrator := seq.NewMigrator(app.SequencerKeeper, seqSubspace)

	// Call the Migrate2to3 function
	err := migrator.Migrate2to3(ctx)

	// Check if there was an error
	if err != nil {
		t.Errorf("Migrate2to3 returned an error: %s", err)
	}

	// Check if the value was migrated correctly
	params = app.SequencerKeeper.GetParams(ctx)
	if params.UnbondingTime != testValue {
		t.Errorf("UnbondingTime not migrated correctly: got %v, expected %v", params.UnbondingTime, testValue)
	}
	return
}
