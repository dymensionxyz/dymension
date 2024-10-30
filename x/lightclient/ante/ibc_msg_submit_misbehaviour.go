package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (i IBCMessagesDecorator) HandleMsgSubmitMisbehaviour(ctx sdk.Context, msg *ibcclienttypes.MsgSubmitMisbehaviour) error {
	_, ok := i.k.GetRollappForClientID(ctx, msg.ClientId)
	if ok {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "cannot submit misbehavour for a canonical client")
	}
	return nil
}
