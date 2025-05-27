package ante

import (
	circuitante "cosmossdk.io/x/circuit/ante"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	proofheightante "github.com/dymensionxyz/dymension/v3/x/delayedack/ante"
	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	ethante "github.com/evmos/ethermint/app/ante"
	txfeesante "github.com/osmosis-labs/osmosis/v15/x/txfees/ante"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

func newCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	mempoolFeeDecorator := txfeesante.NewMempoolFeeDecorator(*options.TxFeesKeeper, options.FeeMarketKeeper)
	deductFeeDecorator := txfeesante.NewDeductFeeDecorator(*options.TxFeesKeeper, options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper)

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		circuitante.NewCircuitBreakerDecorator(options.CircuitKeeper),

		// reject tx with extension options
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),

		// reject MsgEthereumTxs and disable the Msg types that cannot be included on an authz.MsgExec msgs field
		NewRejectMessagesDecorator().
			WithPredicate(BlockTypeUrls(
				0,
				sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
				sdk.MsgTypeURL(&ibcclienttypes.MsgSubmitMisbehaviour{}), // deprecated. not suppose to be used
				sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
				sdk.MsgTypeURL(&vestingtypes.MsgCreatePeriodicVestingAccount{}),
				sdk.MsgTypeURL(&vestingtypes.MsgCreatePermanentLockedAccount{}))),

		// Use Mempool Fee TransferEnabledDecorator from our txfees module instead of default one from auth
		mempoolFeeDecorator,
		deductFeeDecorator,

		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, ethante.DefaultSigVerificationGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),

		// decorator that runs our custom logic for all IBC messages, even wrapped msgs
		NewInnerDecorator(
			proofheightante.NewIBCProofHeightDecorator().InnerCallback,
			lightclientkeeper.NewIBCMessagesDecorator(*options.LightClientKeeper, options.IBCKeeper.ClientKeeper, options.IBCKeeper.ChannelKeeper, options.RollappKeeper).InnerCallback,
		),
		// TODO: make this supported as inner msg
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),

		ethante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...)
}
