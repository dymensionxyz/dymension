package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	ibcante "github.com/cosmos/ibc-go/v6/modules/core/ante"
	txfeesante "github.com/osmosis-labs/osmosis/v15/x/txfees/ante"

	evmos_cosmosante "github.com/evmos/evmos/v12/app/ante/cosmos"
	evmos_evmante "github.com/evmos/evmos/v12/app/ante/evm"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
)

// newEVMAnteHandler creates the default ante handler for Ethereum transactions
func newEVMAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		evmos_evmante.NewEthSetUpContextDecorator(options.EvmKeeper),

		//TODO: need to allow universal fees for Eth as well
		// Check eth effective gas price against the node's minimal-gas-prices config
		evmos_evmante.NewEthMempoolFeeDecorator(options.EvmKeeper), // Check eth effective gas price against minimal-gas-prices
		// Check eth effective gas price against the global MinGasPrice
		evmos_evmante.NewEthMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper), // Check eth effective gas price against the global MinGasPrice

		evmos_evmante.NewEthValidateBasicDecorator(options.EvmKeeper),
		evmos_evmante.NewEthSigVerificationDecorator(options.EvmKeeper),
		evmos_evmante.NewEthAccountVerificationDecorator(options.AccountKeeper, options.EvmKeeper),
		evmos_evmante.NewCanTransferDecorator(options.EvmKeeper),
		evmos_evmante.NewEthGasConsumeDecorator(options.BankKeeper, options.DistributionKeeper, options.EvmKeeper, options.StakingKeeper, options.MaxTxGasWanted),
		evmos_evmante.NewEthIncrementSenderSequenceDecorator(options.AccountKeeper), // innermost AnteDecorator.
		evmos_evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
		evmos_evmante.NewEthEmitEventDecorator(options.EvmKeeper), // emit eth tx hash and index at the very last ante handler.
	)
}

// newLegacyCosmosAnteHandlerEip712 creates an AnteHandler to process legacy EIP-712
// transactions, as defined by the presence of an ExtensionOptionsWeb3Tx extension.
func newLegacyCosmosAnteHandlerEip712(options HandlerOptions) sdk.AnteHandler {
	mempoolFeeDecorator := txfeesante.NewMempoolFeeDecorator(*options.TxFeesKeeper)
	deductFeeDecorator := txfeesante.NewDeductFeeDecorator(*options.TxFeesKeeper, options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper)

	return sdk.ChainAnteDecorators(
		evmos_cosmosante.RejectMessagesDecorator{}, // reject MsgEthereumTxs
		evmos_cosmosante.NewAuthzLimiterDecorator(
			// disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
		),

		ante.NewSetUpContextDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),

		// Use Mempool Fee Decorator from our txfees module instead of default one from auth
		mempoolFeeDecorator,
		deductFeeDecorator,

		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		// Note: signature verification uses EIP instead of the cosmos signature validator
		//nolint: staticcheck
		evmos_cosmosante.NewLegacyEip712SigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		evmos_evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}

func newCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	mempoolFeeDecorator := txfeesante.NewMempoolFeeDecorator(*options.TxFeesKeeper)
	deductFeeDecorator := txfeesante.NewDeductFeeDecorator(*options.TxFeesKeeper, options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper)

	return sdk.ChainAnteDecorators(
		evmos_cosmosante.RejectMessagesDecorator{}, // reject MsgEthereumTxs
		evmos_cosmosante.NewAuthzLimiterDecorator(
			// disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
		),
		ante.NewSetUpContextDecorator(),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		// Use Mempool Fee Decorator from our txfees module instead of default one from auth
		mempoolFeeDecorator,
		deductFeeDecorator,
		deductFeeDecorator,
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		evmos_evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)
}
