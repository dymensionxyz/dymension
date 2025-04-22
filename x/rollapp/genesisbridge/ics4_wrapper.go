package genesisbridge

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type ChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error) // implemented by ibc channel keeper
}

type ICS4Wrapper struct {
	porttypes.ICS4Wrapper
	rollappK              RollappKeeperMinimal
	getChannelClientState ChannelKeeper
}

func NewICS4Wrapper(
	next porttypes.ICS4Wrapper,
	rollappKeeper RollappKeeperMinimal,
	getChannelClientState ChannelKeeper,
) *ICS4Wrapper {
	return &ICS4Wrapper{
		ICS4Wrapper:           next,
		rollappK:              rollappKeeper,
		getChannelClientState: getChannelClientState,
	}
}

func (w *ICS4Wrapper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	if err := w.transferAllowed(ctx, sourcePort, sourceChannel); err != nil {
		return 0, errorsmod.Wrap(err, "transfer allowed")
	}
	return w.ICS4Wrapper.SendPacket(
		ctx,
		chanCap,
		sourcePort,
		sourceChannel,
		timeoutHeight,
		timeoutTimestamp,
		data,
	)
}

func (w *ICS4Wrapper) transferAllowed(ctx sdk.Context, sourcePort string, sourceChannel string) error {
	ra, err := w.rollappK.GetRollappByPortChan(ctx, sourcePort, sourceChannel)
	if err != nil {
		if errorsmod.IsOf(err, types.ErrRollappNotFound) {
			// Two cases
			// 1. Non rollapp - Transfers are allowed
			// 2. It is for rollapp but the light client of this transfer is not canonical and will never
			//    be marked canonical: we can't set a canonical client if it's already have channels, so this transfer corresponds to a not-relevant channel.
			return nil
		}
		return errorsmod.Wrap(err, "rollapp by port chan")
	}

	if !w.rollappK.IsCanonicalChannel(ctx, ra.RollappId, sourcePort, sourceChannel) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "non canonical channel %s for rollapp %s", sourceChannel, ra.RollappId)
	}

	return nil
}
