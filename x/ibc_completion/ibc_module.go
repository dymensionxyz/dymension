package ibc_completion

import (
	"encoding/json"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	denomutils "github.com/dymensionxyz/dymension/v3/utils/denom"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

const (
	ModuleName = "ibc_completion"
)

var _ porttypes.IBCModule = &IBCModule{}

type IBCModule struct {
	porttypes.IBCModule
	rollappK RolKeeper
	dackK    DackKeeper
}

func NewIBCModule(
	next porttypes.IBCModule,
	rollappKeeper RolKeeper,
	dackKeeper DackKeeper,
) IBCModule {
	return IBCModule{
		IBCModule: next,
		rollappK:  rollappKeeper,
		dackK:     dackKeeper,
	}
}

type RolKeeper interface {
	GetValidTransfer(ctx sdk.Context, data []byte, destPort string, destChannel string) (rollapptypes.TransferData, error)
}

type DackKeeper interface {
	ValidateCompletionHook(info commontypes.CompletionHookCall) error
	RunCompletionHook(ctx sdk.Context, fundsSrc sdk.AccAddress, budget sdk.Coin, call commontypes.CompletionHookCall) error
}

func (m IBCModule) logger(
	ctx sdk.Context,
	packet channeltypes.Packet,
	method string,
) log.Logger {
	return ctx.Logger().With(
		"module", ModuleName,
		"packet_source_port", packet.SourcePort,
		"packet_destination_port", packet.DestinationPort,
		"packet_sequence", packet.Sequence,
		"method", method,
	)
}

type Memo struct {
	// can be nil
	OnCompletionHook []byte `json:"on_completion,omitempty"`
	// can be nil
	PFM []byte `json:"forward,omitempty"`
}

const (
	pfmKey = "forward"
)

// for non-rollapp packets only, process any completion hooks
func (m IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := m.logger(ctx, packet, "OnRecvPacket")
	transfer, err := m.rollappK.GetValidTransfer(ctx, packet.GetData(), packet.GetDestPort(), packet.GetDestChannel())
	if err != nil {
		l.Error("Get valid transfer.", "err", err)
		err = errorsmod.Wrapf(err, "%s: get valid transfer", ModuleName)
		return uevent.NewErrorAcknowledgement(ctx, err)
	}

	if transfer.IsRollapp() {
		return m.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	memoBz := []byte(transfer.Memo)
	if memoHasConflictingMiddleware(memoBz) {
		return m.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	var memo Memo
	err = json.Unmarshal(memoBz, &memo)
	if err != nil {
		return m.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	if len(memo.OnCompletionHook) == 0 {
		return m.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	var hook commontypes.CompletionHookCall
	err = proto.Unmarshal(memo.OnCompletionHook, &hook)
	if err != nil {
		return uevent.NewErrorAcknowledgement(ctx, fmt.Errorf("unmarshal completion hook: %w", err))
	}
	if err := hook.ValidateBasic(); err != nil {
		return uevent.NewErrorAcknowledgement(ctx, fmt.Errorf("val basic completion hook: %w", err))
	}
	if err := m.dackK.ValidateCompletionHook(hook); err != nil {
		return uevent.NewErrorAcknowledgement(ctx, fmt.Errorf("full validate completion hook: %w", err))
	}

	// first need to complete the inbound transfer so that the funds are available
	// (that's why we cant allow PFM or other middlewares which conflict)

	amt, ok := math.NewIntFromString(transfer.Amount)
	if !ok {
		return uevent.NewErrorAcknowledgement(ctx, errors.New("invalid amount string"))
	}
	denom := denomutils.GetIncomingTransferDenom(packet, transfer.FungibleTokenPacketData)
	fundsSrc, err := sdk.AccAddressFromBech32(transfer.Receiver)
	if err != nil {
		return uevent.NewErrorAcknowledgement(ctx, fmt.Errorf("invalid recipient address: %w", err))
	}
	budget := sdk.NewCoin(denom, amt)

	ack := m.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if ack != nil {
		return ack
	}

	err = m.dackK.RunCompletionHook(ctx, fundsSrc, budget, hook)
	if err != nil {
		return uevent.NewErrorAcknowledgement(ctx, fmt.Errorf("run completion hook: %w", err))
	}

	return nil
}

func memoHasConflictingMiddleware(memoBz []byte) bool {
	d := make(map[string]interface{})
	err := json.Unmarshal(memoBz, &d)
	return err != nil || d[pfmKey] != nil
}
