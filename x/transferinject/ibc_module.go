// Package transferinject module provides IBC middleware for sending and acknowledging IBC packets with injecting additional packet metadata to IBC packets.
package transferinject

import (
	"errors"
	. "slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"

	"github.com/dymensionxyz/dymension/v3/x/transferinject/types"
)

type IBCModule struct {
	porttypes.IBCModule

	rollappKeeper types.RollappKeeper
}

// NewIBCModule creates a new IBCModule.
// It intercepts acknowledged incoming IBC packets and adds token metadata that had just been registered on the rollapp itself,
// to the local rollapp record.
func NewIBCModule(
	ibc porttypes.IBCModule,
	rollappKeeper types.RollappKeeper,
) *IBCModule {
	return &IBCModule{
		IBCModule:     ibc,
		rollappKeeper: rollappKeeper,
	}
}

// OnAcknowledgementPacket adds the token metadata to the rollapp if it doesn't exist
func (m *IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errorsmod.Wrap(errors.Join(err, errortypes.ErrJSONUnmarshal), "ics20 transfer packet acknowledgement")
	}

	if !ack.Success() {
		return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	transfer, err := m.rollappKeeper.GetValidTransferFromSentPacket(ctx, packet)
	if err != nil {
		return errorsmod.Wrap(err, "get valid transfer from sent packet")
	}

	packetMetadata, err := types.ParsePacketMetadata(transfer.GetMemo())
	if err != nil {
		return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	if !transfer.IsRollapp() {
		return errorsmod.Wrap(errors.Join(err, errortypes.ErrInvalidRequest), "got a memo so should get rollapp, but didnt")
	}

	if !Contains(transfer.Rollapp.RegisteredDenoms, packetMetadata.DenomMetadata.Base) {
		// add the new token denom base to the list of rollapp's registered denoms
		transfer.Rollapp.RegisteredDenoms = append(transfer.Rollapp.RegisteredDenoms, packetMetadata.DenomMetadata.Base)

		m.rollappKeeper.SetRollapp(ctx, *transfer.Rollapp)
	}

	return m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}
