package ante

import (
	storetypes "cosmossdk.io/store/types"
	circuitante "cosmossdk.io/x/circuit/ante"
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ethante "github.com/evmos/ethermint/app/ante"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
)

type HandlerOptions struct {
	ExtensionOptionChecker ante.ExtensionOptionChecker
	FeegrantKeeper         FeegrantKeeper
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params types.Params) error
	TxFeeChecker           ante.TxFeeChecker

	AccountKeeper     AccountKeeper
	BankKeeper        BankKeeper
	IBCKeeper         *ibckeeper.Keeper
	FeeMarketKeeper   FeeMarketKeeper
	EvmKeeper         ethante.EVMKeeper
	TxFeesKeeper      *txfeeskeeper.Keeper
	MaxTxGasWanted    uint64
	RollappKeeper     rollappkeeper.Keeper
	LightClientKeeper *lightclientkeeper.Keeper
	CircuitKeeper     circuitante.CircuitBreaker
}

func (options HandlerOptions) validate() error {
	if options.AccountKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.FeeMarketKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "fee market keeper is required for AnteHandler")
	}
	if options.EvmKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "evm keeper is required for AnteHandler")
	}
	if options.TxFeesKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "tx fees keeper is required for AnteHandler")
	}
	if options.LightClientKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "light client keeper is required for AnteHandler")
	}
	if options.CircuitKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "circuit breaker keeper is required for AnteHandler")

	}
	return nil
}
