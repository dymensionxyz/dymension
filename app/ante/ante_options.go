package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ethante "github.com/evmos/ethermint/app/ante"

	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
)

type HandlerOptions struct {
	Cdc                    codec.Codec
	AccountKeeper          *authkeeper.AccountKeeper
	BankKeeper             bankkeeper.Keeper
	IBCKeeper              *ibckeeper.Keeper
	FeeMarketKeeper        ethante.FeeMarketKeeper
	EvmKeeper              ethante.EVMKeeper
	FeegrantKeeper         ante.FeegrantKeeper
	TxFeesKeeper           *txfeeskeeper.Keeper
	SignModeHandler        authsigning.SignModeHandler
	MaxTxGasWanted         uint64
	ExtensionOptionChecker ante.ExtensionOptionChecker
	RollappKeeper          rollappkeeper.Keeper
	LightClientKeeper      *lightclientkeeper.Keeper
	GroupKeeper            *groupkeeper.Keeper
}

func (options HandlerOptions) validate() error {
	if options.Cdc == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "codec is required for AnteHandler")
	}
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
	if options.GroupKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "group keeper is required for AnteHandler")
	}
	return nil
}
