package keeper

import (
	errorsmod "cosmossdk.io/errors"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	types "github.com/dymensionxyz/dymension/v3/x/forward/types"
	transfer "github.com/dymensionxyz/dymension/v3/x/transfer"
)

var _ transfer.CompletionHookInstance = rollappToHubCompletion{}

func (k Keeper) Hook() rollappToHubCompletion {
	return rollappToHubCompletion{
		Keeper: &k,
	}
}

type rollappToHubCompletion struct {
	*Keeper
}

func (h rollappToHubCompletion) ValidateData(data []byte) error {
	var d types.HookEIBCtoHL
	err := proto.Unmarshal(data, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal")
	}
	if err := d.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "validate")
	}
	return nil
}

// TODO: rename method
// at this point funds have not been sent from the fulfiller/eibc LP/funds provider to the recipient (or anywhere else)
func (h rollappToHubCompletion) Run(ctx sdk.Context, order *eibctypes.DemandOrder, fundsSource sdk.AccAddress,
	newTransferRecipient sdk.AccAddress,
	fulfiller sdk.AccAddress, hookData []byte) error {

	budget := sdk.NewCoin(order.Denom(), order.PriceAmount())
	h.refundOnError(ctx, func() error {
		var d types.HookEIBCtoHL
		err := proto.Unmarshal(hookData, &d)
		if err != nil {
			return errorsmod.Wrap(err, "unmarshal")
		}

		return h.forwardToHyperlane(ctx, fundsSource, budget, d)
	}, nil, "", order.GetRecipientBech32Address(), budget)
	return nil
}

func (k Keeper) forwardToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer, maxBudget sdk.Coin, memo []byte) error {

	// Case analysis: if the IBC transfer fails due to error ack or timeout...
	//
	// Case a, finalize:
	//  - Packet is passed to the rest of the transfer stack
	//  - Ultimately ends up in refundPacketToken func of transfer keeper. It will send coins to back to this module.
	// Case b, fulfill:
	//  - Funds are sent from eibc fulfiller to this module.
	//    (new recipient becomes whatever fulfiller specified, but that's not interesting here)
	//
	// So in either case the funds are given back to this module, but we can't easily hook onto this occurring (in finalize case)
	// The cleanest solution is to have

	// TODO:, it occurs to me the simplest thing to do is:
	// - make the sender the recovery addr
	// - then in ibc case, this guy automatically gets the tokens back

	m := ibctransfertypes.NewMsgTransfer(
		transfer.SourcePort,
		transfer.SourceChannel,
		maxBudget,
		warptypes.ModuleName,
		transfer.Receiver,
		ibcclienttypes.Height{}, // ignore, removed in ibc v2 also
		transfer.TimeoutTimestamp,
		string(memo),
	)

	_, err := k.transferK.Transfer(ctx, m)

	return err
}
