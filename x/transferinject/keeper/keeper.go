package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	ctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	icstypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/v3/x/transferinject/types"
)

type Keeper struct {
	cdc codec.BinaryCodec
	icstypes.ICS4Wrapper
	getValue injValFunc
}

func NewTransferInject(
	cdc codec.BinaryCodec,
	icswrap icstypes.ICS4Wrapper,
	getValue injValFunc,
) *Keeper {
	return &Keeper{
		cdc:         cdc,
		ICS4Wrapper: icswrap,
		getValue:    getValue,
	}
}

type injValFunc func(ctx sdk.Context, in proto.Message, destinationPort string, destinationChannel string) (out proto.Message, err error)

func WithRollappDenomMetadata(delayedackKeeper delayedackkeeper.Keeper, bankKeeper types.BankKeeper) injValFunc {
	return func(ctx sdk.Context, in proto.Message, destinationPort string, destinationChannel string) (proto.Message, error) {
		// TODO: first check if rollapp has denom

		rollapp, err := delayedackKeeper.ExtractRollappFromChannel(ctx, destinationPort, destinationChannel)
		if err != nil {
			return nil, fmt.Errorf("cannot extract rollapp id from packet: %w", err)
		}

		if rollapp == nil {
			return in, nil
		}

		fungibleTokenPacketData, ok := in.(*transfertypes.FungibleTokenPacketData)
		if !ok {
			return nil, errorsmod.Wrapf(errortypes.ErrInvalidType, "invalid packet data type")
		}

		inDenom := fungibleTokenPacketData.Denom

		denomHash, err := DenomHash(inDenom)
		if err != nil {
			return nil, errorsmod.Wrapf(errortypes.ErrInvalidType, "denom hash not found")
		}

		// TODO: make sure token metadata in the rollapp is up to date
		for _, dm := range rollapp.TokenMetadata {
			if dm.Base == denomHash || dm.Base == inDenom { // TODO: probably won't work
				return in, nil
			}
		}

		denomMetaData, ok := bankKeeper.GetDenomMetaData(ctx, denomHash)
		if !ok {
			return nil, errorsmod.Wrapf(errortypes.ErrInvalidType, "denom metadata not found")
		}

		return &denomMetaData, nil
	}
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (t *Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	destinationPort string, destinationChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	// if the packet is a wrapped fungible token packet, or getValue func was not provided, just move on with it
	var wrappedPacket types.WrappedFungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(data, &wrappedPacket); err == nil || t.getValue == nil {
		return t.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
	}

	// otherwise, it's a normal packet, and we need to inject a custom value
	wrappedPacket.FungibleTokenPacketData = new(transfertypes.FungibleTokenPacketData)
	if err = types.ModuleCdc.UnmarshalJSON(data, wrappedPacket.FungibleTokenPacketData); err != nil {
		return 0, fmt.Errorf("cannot unmarshal transfer packet data: %w", err)
	}

	val, err := t.getValue(ctx, wrappedPacket.FungibleTokenPacketData, destinationPort, destinationChannel)
	if err != nil {
		return 0, fmt.Errorf("cannot get injected value: %w", err)
	}

	wrappedPacket.Data, err = ctypes.NewAnyWithValue(val)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidType, "cannot create injected value")
	}

	data, err = types.ModuleCdc.MarshalJSON(&wrappedPacket)
	if err != nil {
		return 0, errorsmod.Wrapf(errortypes.ErrInvalidType, "cannot marshal ICS-20 transfer packet data")
	}

	return t.ICS4Wrapper.SendPacket(ctx, chanCap, destinationPort, destinationChannel, timeoutHeight, timeoutTimestamp, data)
}

func DenomHash(trace string) (string, error) {
	if trace == "adym" {
		return trace, nil
	}

	denomTrace := transfertypes.ParseDenomTrace(trace)
	if err := denomTrace.Validate(); err != nil {
		return "", status.Error(codes.InvalidArgument, err.Error())
	}

	return "ibc/" + denomTrace.Hash().String(), nil
}
