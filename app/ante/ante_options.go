package ante

import (
	circuitante "cosmossdk.io/x/circuit/ante"
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ethante "github.com/evmos/ethermint/app/ante"

	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
)

// max depth of nested messages
// Depth 0 is top level message
// Depth 1 or more is wrapped in something
const maxInnerDepth = 6

type HandlerOptions struct {
	ExtensionOptionChecker ante.ExtensionOptionChecker
	FeegrantKeeper         FeegrantKeeper
	SignModeHandler        *txsigning.HandlerMap

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
