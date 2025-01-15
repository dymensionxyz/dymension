package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ethante "github.com/evmos/ethermint/app/ante"

	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
)

// FeeMarketKeeper defines the expected keeper interface used on the AnteHandler
type FeeMarketKeeper interface {
	ethante.FeeMarketKeeper
	GetMinGasPrice(ctx sdk.Context) (minGasPrice math.LegacyDec)
}

type HandlerOptions struct {
	ante.HandlerOptions
	// AccountKeeper          *authkeeper.AccountKeeper
	// BankKeeper             bankkeeper.Keeper
	// FeegrantKeeper         ante.FeegrantKeeper
	// ExtensionOptionChecker ante.ExtensionOptionChecker
	// SignModeHandler        authsigning.SignModeHandler
	IBCKeeper         *ibckeeper.Keeper
	FeeMarketKeeper   FeeMarketKeeper
	EvmKeeper         ethante.EVMKeeper
	TxFeesKeeper      *txfeeskeeper.Keeper
	MaxTxGasWanted    uint64
	RollappKeeper     rollappkeeper.Keeper
	LightClientKeeper *lightclientkeeper.Keeper
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
	return nil
}
