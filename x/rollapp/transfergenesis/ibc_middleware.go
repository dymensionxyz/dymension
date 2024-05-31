package transfergenesis

import (
	"encoding/json"
	"errors"
	"fmt"

	utilsmemo "github.com/dymensionxyz/dymension/v3/utils/memo"

	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

var _ porttypes.Middleware = &IBCMiddleware{}

type IBCMiddleware struct {
	porttypes.Middleware // next one
	delayedackKeeper     delayedackkeeper.Keeper
	rollappKeeper        rollappkeeper.Keeper
}

func NewIBCMiddleware(next porttypes.Middleware, keeper delayedackkeeper.Keeper, raK rollappkeeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		Middleware:       next,
		delayedackKeeper: keeper,
		rollappKeeper:    raK,
	}
}

func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	// may modify packet
	err := im.handleGenesisTransfers(ctx, &packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return im.OnRecvPacket(ctx, packet, relayer)
}

type genesisTransferDenomMemo struct {
	Data struct {
		Denom banktypes.Metadata `json:"denom"`
	} `json:"genesis_transfer"`
}

func (im IBCMiddleware) handleGenesisTransfers(
	ctx sdk.Context,
	packet *channeltypes.Packet,
) error {
	if !im.delayedackKeeper.IsRollappsEnabled(ctx) {
		return nil
	}

	l := ctx.Logger().With(
		"middleware", "transferGenesis",
		"packet_source_port", packet.SourcePort,
		"packet_destination_port", packet.DestinationPort,
		"packet_sequence", packet.Sequence)

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, "fungible token packet")
	}

	memo := data.GetMemo()
	var wrappedDenom genesisTransferDenomMemo // wrapped for memo namespacing reasons
	err := json.Unmarshal([]byte(memo), &wrappedDenom)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, "memo")
	}

	denom := wrappedDenom.Data.Denom

	l.Info("got the special memo!") // TODO: fix

	chaID := "channel-0"
	raID := "rollappevm_1234-1"

	ra, ok := im.rollappKeeper.GetRollapp(ctx, raID)
	if !ok {
		panic(errors.New("must find rollapp"))
	}

	_ = ra

	err = im.rollappKeeper.MarkGenesisAsHappened(ctx, chaID, raID)
	if err != nil {
		err = fmt.Errorf("mark genesis: %w", err)
		l.Error("OnRecvPacket", "err", err)
		panic(err)
	}

	err = im.rollappKeeper.RegisterDenomMetadata(ctx, raID, chaID, denom)
	if err != nil {
		err = fmt.Errorf("register denom meta: %w", err)
		l.Error("OnRecvPacket", "err", err)
		panic(err)
	}

	l.Info("Registered denom meta data from genesis transfer.")

	newMemo := delayedacktypes.Memo{}
	newMemo.Data.SkipDelay = true

	newMemoStr, err := utilsmemo.Merge(memo, newMemo)
	if err != nil {
		return errorsmod.Wrap(err, "memo merge")
	}
	data.Memo = newMemoStr

	bz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrJSONMarshal, "to ibc packet")
	}
	packet.Data = bz

	/*
			TODO: where was I on Friday afternoon?
			I'll need to make delayed ack scrub out the skip memo, but only if it's an 'internal' skip memo, thereby knowing it came
			from this chain
			Actually, maybe instead of screwing with the memo, I can use the context?
			Anyways, next steps are to check it's all working
			Then redefine the memo to include the total number of genesis tx and the id
		Then write the code that will wait for all of them to finish and trigger the dispute period
	*/

	return nil
}
