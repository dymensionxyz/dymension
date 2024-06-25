package transferinject

import (
	"errors"
	. "slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/transferinject/types"
)

type ICS4Wrapper struct {
	porttypes.ICS4Wrapper

	rollappKeeper types.RollappKeeper
	bankKeeper    types.BankKeeper
}

// NewICS4Wrapper creates a new ICS4Wrapper.
// It intercepts outgoing IBC packets and adds token metadata to the memo if the rollapp doesn't have it.
// This is a solution for adding token metadata to fungible tokens transferred over IBC,
// targeted at rollapps that don't have the token metadata for the token being transferred.
// More info here: https://www.notion.so/dymension/ADR-x-IBC-Denom-Metadata-Transfer-From-Hub-to-Rollapp-d3791f524ac849a9a3eb44d17968a30b
func NewICS4Wrapper(
	next porttypes.ICS4Wrapper,
	rollappKeeper types.RollappKeeper,
	bankKeeper types.BankKeeper,
) *ICS4Wrapper {
	return &ICS4Wrapper{
		ICS4Wrapper:   next,
		rollappKeeper: rollappKeeper,
		bankKeeper:    bankKeeper,
	}
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (m *ICS4Wrapper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	srcPort string, srcChan string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	transfer, err := m.rollappKeeper.GetValidTransfer(ctx, data, srcPort, srcChan)
	if err != nil {
		return 0, errorsmod.Wrap(err, "transfer inject: get valid transfer")
	}

	if types.MemoAlreadyHasPacketMetadata(transfer.GetMemo()) {
		return 0, types.ErrMemoTransferInjectAlreadyExists
	}

	if
	// TODO: currently we check if receiving chain is a rollapp, consider that other chains also might want this feature
	// meaning, find a better way to check if the receiving chain supports this middleware
	!transfer.IsRollapp() || // proceed as normal
		transfertypes.ReceiverChainIsSource(srcPort, srcChan, transfer.Denom) {
		return m.ICS4Wrapper.SendPacket(ctx, chanCap, srcPort, srcChan, timeoutHeight, timeoutTimestamp, data)
	}

	// Check if the rollapp already contains the denom metadata by matching the base of the denom metadata.
	// At the first match, we assume that the rollapp already contains the metadata.
	// It would be technically possible to have a race condition where the denom metadata is added to the rollapp
	// from another packet before this packet is acknowledged.
	if Contains(transfer.Rollapp.RegisteredDenoms, transfer.GetDenom()) {
		return m.ICS4Wrapper.SendPacket(ctx, chanCap, srcPort, srcChan, timeoutHeight, timeoutTimestamp, data)
	}

	// get the denom metadata from the bank keeper, if it doesn't exist, move on to the next middleware in the chain
	denomMetadata, ok := m.bankKeeper.GetDenomMetaData(ctx, transfer.GetDenom())
	if !ok {
		return m.ICS4Wrapper.SendPacket(ctx, chanCap, srcPort, srcChan, timeoutHeight, timeoutTimestamp, data)
	}

	transfer.Memo, err = types.AddDenomMetadataToMemo(transfer.Memo, denomMetadata)
	if err != nil {
		if errors.Is(err, types.ErrMemoTransferInjectAlreadyExists) {
			err = errors.Join(err, gerrc.ErrPermissionDenied)
		}
		return 0, errorsmod.Wrap(err, "transfer inject: add denom metadata to memo")
	}

	data, err = types.ModuleCdc.MarshalJSON(&transfer.FungibleTokenPacketData)
	if err != nil {
		return 0, errorsmod.Wrap(errors.Join(err, errortypes.ErrJSONMarshal), "transfer inject: ics20 transfer packet data")
	}

	return m.ICS4Wrapper.SendPacket(ctx, chanCap, srcPort, srcChan, timeoutHeight, timeoutTimestamp, data)
}
