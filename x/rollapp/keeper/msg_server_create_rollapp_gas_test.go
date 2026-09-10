package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// Gas for a tx must be a pure function of the tx bytes and the committed state.
// Simulating one unchanged MsgCreateRollapp against one unchanged committed state
// must therefore return the same figure every time, and delivering it on two
// independently built chains must agree.
const createRollappGasRuns = 20

func TestCreateRollappGasIsDeterministic(t *testing.T) {
	t.Run("simulate", func(t *testing.T) {
		a, txBytes := setupCreateRollappGasFixture(t)

		gas := make([]uint64, createRollappGasRuns)
		for i := range gas {
			info, res, err := a.Simulate(txBytes)
			require.NoError(t, err)
			require.NotNil(t, res, "run %d: simulation returned no result", i)
			gas[i] = info.GasUsed
		}

		for i, g := range gas {
			require.Equal(t, gas[0], g, "run %d consumed %d gas, first run consumed %d (all runs: %v)", i, g, gas[0], gas)
		}
		t.Logf("%d simulations against identical committed state: %d gas each", createRollappGasRuns, gas[0])

		// A simulation must not leave anything behind, or the next one would price differently.
		check := a.NewContextLegacy(true, cmtproto.Header{Height: 1, ChainID: apptesting.TestChainID})
		_, found := a.RollappKeeper.GetRollapp(check, createRollappGasRollappID)
		require.False(t, found, "simulation leaked the rollapp into the check state")
		require.Zero(t, a.AccountKeeper.GetAccount(check, createRollappGasCreator()).GetSequence(),
			"simulation leaked the signer's sequence increment into the check state")
	})

	t.Run("deliver", func(t *testing.T) {
		a1, txBytes1 := setupCreateRollappGasFixture(t)
		a2, txBytes2 := setupCreateRollappGasFixture(t)
		require.Equal(t, txBytes1, txBytes2, "fixtures must produce byte-identical txs to be comparable")

		require.Equal(t, deliverCreateRollapp(t, a1, txBytes1), deliverCreateRollapp(t, a2, txBytes2))
	})

	// Guards the assertions above against passing vacuously: the measurement has to
	// move when the tx does.
	t.Run("gas tracks the payload", func(t *testing.T) {
		a, txBytes := setupCreateRollappGasFixture(t)
		base, _, err := a.Simulate(txBytes)
		require.NoError(t, err)

		check := a.NewContextLegacy(true, cmtproto.Header{Height: 1, ChainID: apptesting.TestChainID})
		acc := a.AccountKeeper.GetAccount(check, createRollappGasCreator())
		larger := signTx(t, a.TxConfig(), createRollappGasMsg("1234567890abcdefg1234567890abcdefg1234567890abcdefg"),
			createRollappGasPrivKey(), acc.GetAccountNumber(), acc.GetSequence())

		got, _, err := a.Simulate(larger)
		require.NoError(t, err)
		require.NotEqual(t, base.GasUsed, got.GasUsed)
	})
}

func deliverCreateRollapp(t *testing.T, a *app.App, txBytes []byte) uint64 {
	t.Helper()

	res, err := a.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.Equal(t, uint32(0), res.TxResults[0].Code, res.TxResults[0].Log)
	return uint64(res.TxResults[0].GasUsed)
}

const (
	createRollappGasRollappID = "rollappevm_1234-1"
	createRollappGasAlias     = "icxuv"
)

func createRollappGasPrivKey() *ethsecp256k1.PrivKey {
	priv := &ethsecp256k1.PrivKey{Key: make([]byte, 32)}
	priv.Key[31] = 1
	return priv
}

func createRollappGasCreator() sdk.AccAddress {
	return sdk.AccAddress(createRollappGasPrivKey().PubKey().Address())
}

func createRollappGasMsg(genesisChecksum string) *types.MsgCreateRollapp {
	return &types.MsgCreateRollapp{
		Creator:          createRollappGasCreator().String(),
		RollappId:        createRollappGasRollappID,
		InitialSequencer: "*",
		MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin,
		Alias:            createRollappGasAlias,
		VmType:           types.Rollapp_EVM,
		GenesisInfo: &types.GenesisInfo{
			Bech32Prefix:    "ethm",
			GenesisChecksum: genesisChecksum,
			InitialSupply:   math.NewInt(1000),
			NativeDenom: types.DenomMetadata{
				Display:  "RAX",
				Base:     "urax",
				Exponent: 18,
			},
		},
	}
}

// setupCreateRollappGasFixture returns a chain whose state is committed, plus a signed
// MsgCreateRollapp tx ready to be simulated or delivered against that state.
func setupCreateRollappGasFixture(t *testing.T) (*app.App, []byte) {
	t.Helper()

	a := apptesting.Setup(t)
	ctx := a.NewContextLegacy(false, cmtproto.Header{Height: 1, ChainID: apptesting.TestChainID})

	msg := createRollappGasMsg("1234567890abcdefg")
	apptesting.FundForAliasRegistration(a, ctx, msg.Alias, msg.Creator)

	_, err := a.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = a.Commit()
	require.NoError(t, err)

	committed := a.NewContextLegacy(true, cmtproto.Header{Height: 1, ChainID: apptesting.TestChainID})
	acc := a.AccountKeeper.GetAccount(committed, createRollappGasCreator())
	require.NotNil(t, acc)

	return a, signTx(t, a.TxConfig(), msg, createRollappGasPrivKey(), acc.GetAccountNumber(), acc.GetSequence())
}

func signTx(t *testing.T, txConfig client.TxConfig, msg sdk.Msg, priv cryptotypes.PrivKey, accNum, seq uint64) []byte {
	t.Helper()

	b := txConfig.NewTxBuilder()
	require.NoError(t, b.SetMsgs(msg))
	b.SetGasLimit(200000)

	sigData := &signingtypes.SingleSignatureData{SignMode: signingtypes.SignMode_SIGN_MODE_DIRECT}
	sig := signingtypes.SignatureV2{PubKey: priv.PubKey(), Data: sigData, Sequence: seq}
	require.NoError(t, b.SetSignatures(sig))

	signBytes, err := authsigning.GetSignBytesAdapter(
		t.Context(), txConfig.SignModeHandler(), signingtypes.SignMode_SIGN_MODE_DIRECT,
		authsigning.SignerData{
			Address:       sdk.AccAddress(priv.PubKey().Address()).String(),
			ChainID:       apptesting.TestChainID,
			AccountNumber: accNum,
			Sequence:      seq,
			PubKey:        priv.PubKey(),
		}, b.GetTx())
	require.NoError(t, err)

	sigData.Signature, err = priv.Sign(signBytes)
	require.NoError(t, err)
	require.NoError(t, b.SetSignatures(sig))

	txBytes, err := txConfig.TxEncoder()(b.GetTx())
	require.NoError(t, err)
	return txBytes
}
