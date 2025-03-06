package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	"github.com/dymensionxyz/dymension/v3/x/common/types"
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
		// reject MsgEthereumTxs and disable the Msg types that cannot be included on an authz.MsgExec msgs field
		NewRejectMessagesDecorator().WithPredicate(
			BlockTypeUrls(
				1,
				// Only blanket rejects depth greater than zero because we have our own custom logic for depth 0
				// Note that there is never a genuine reason to pass both ibc update client and misbehaviour submission through gov or auth,
				// it's always done by relayers directly.
				sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}))).
			WithPredicate(BlockTypeUrls(
				0,
				sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
				sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
				sdk.MsgTypeURL(&vestingtypes.MsgCreatePeriodicVestingAccount{}),
				sdk.MsgTypeURL(&vestingtypes.MsgCreatePermanentLockedAccount{}))),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
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
		types.NewIBCProofHeightDecorator(),
		lightclientkeeper.NewIBCMessagesDecorator(*options.LightClientKeeper, options.IBCKeeper.ClientKeeper, options.IBCKeeper.ChannelKeeper, options.RollappKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		ethante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...)
}
