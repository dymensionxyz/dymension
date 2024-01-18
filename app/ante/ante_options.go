package ante

import (
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
	evmante "github.com/evmos/evmos/v12/app/ante/evm"
	anteutils "github.com/evmos/evmos/v12/app/ante/utils"

	vestingtypes "github.com/evmos/evmos/v12/x/vesting/types"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type HandlerOptions struct {
	AccountKeeper          *authkeeper.AccountKeeper
	BankKeeper             bankkeeper.Keeper
	IBCKeeper              *ibckeeper.Keeper
	FeeMarketKeeper        evmante.FeeMarketKeeper
	StakingKeeper          vestingtypes.StakingKeeper
	DistributionKeeper     anteutils.DistributionKeeper
	EvmKeeper              evmante.EVMKeeper
	FeegrantKeeper         ante.FeegrantKeeper
	TxFeesKeeper           *txfeeskeeper.Keeper
	SignModeHandler        authsigning.SignModeHandler
	MaxTxGasWanted         uint64
	SigGasConsumer         func(meter sdk.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	ExtensionOptionChecker ante.ExtensionOptionChecker
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
	return nil
}
