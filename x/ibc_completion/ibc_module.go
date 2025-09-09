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

	"github.com/dymensionxyz/dymension/v3/x/ibc_completion/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

const (
	ModuleName = "ibc_completion"
	pfmKey     = "forward"
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

// for non-rollapp packets only, process any completion hooks
func (m IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := m.logger(ctx, packet, "OnRecvPacket")
	port, channel := commontypes.PacketHubPortChan(commontypes.RollappPacket_ON_RECV, packet)
	transfer, err := m.rollappK.GetValidTransfer(ctx, packet.GetData(), port, channel)
	if err != nil {
		l.Error("Get valid transfer.", "err", err)
		err = errorsmod.Wrapf(err, "%s: get valid transfer", ModuleName)
		return uevent.NewErrorAcknowledgement(ctx, err)
	}

	h, err := m.getCompletionHookToRun(ctx, packet, transfer)
	if errorsmod.IsOf(err, ErrRoutedThroughEIBC) {
		return m.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}
	if errorsmod.IsOf(err, ErrMemoHasConflictingMiddleware) {
		return m.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}
	if err != nil {
		l.Error("Get completion hook to run.", "err", err)
		err = errorsmod.Wrapf(err, "%s: get completion hook to run", ModuleName)
		return uevent.NewErrorAcknowledgement(ctx, err)
	}

	ack := m.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	err = m.dackK.RunCompletionHook(ctx, h.fundsSrc, h.budget, h.hook)
	if err != nil {
		return uevent.NewErrorAcknowledgement(ctx, fmt.Errorf("run completion hook: %w", err))
	}

	return ack
}

type completionHookRunnable struct {
	hook     commontypes.CompletionHookCall
	fundsSrc sdk.AccAddress
	budget   sdk.Coin
}

var (
	ErrRoutedThroughEIBC            = errors.New("routed through EIBC")
	ErrMemoHasConflictingMiddleware = errors.New("memo has conflicting middleware")
)

func (m IBCModule) getCompletionHookToRun(ctx sdk.Context, packet channeltypes.Packet, transfer rollapptypes.TransferData) (completionHookRunnable, error) {
	routedThroughEIBC := !commontypes.WasNotDelayed(ctx)

	if routedThroughEIBC {
		return completionHookRunnable{}, ErrRoutedThroughEIBC
	}

	hook, err := getCompletionHookFromMemo(packet.GetData())
	if err != nil {
		return completionHookRunnable{}, err
	}

	if err := hook.ValidateBasic(); err != nil {
		return completionHookRunnable{}, fmt.Errorf("val basic completion hook: %w", err)
	}
	if err := m.dackK.ValidateCompletionHook(hook); err != nil {
		return completionHookRunnable{}, fmt.Errorf("full validate completion hook: %w", err)
	}

	// first need to complete the inbound transfer so that the funds are available
	// (that's why we cant allow PFM or other middlewares which conflict)

	amt, ok := math.NewIntFromString(transfer.Amount)
	if !ok {
		return completionHookRunnable{}, errors.New("invalid amount string")
	}
	denom := denomutils.GetIncomingTransferDenom(packet, transfer.FungibleTokenPacketData)
	fundsSrc, err := sdk.AccAddressFromBech32(transfer.Receiver)
	if err != nil {
		return completionHookRunnable{}, fmt.Errorf("invalid recipient address: %w", err)
	}
	budget := sdk.NewCoin(denom, amt)
	return completionHookRunnable{
		hook:     hook,
		fundsSrc: fundsSrc,
		budget:   budget,
	}, nil
}

func getCompletionHookFromMemo(memoBz []byte) (commontypes.CompletionHookCall, error) {
	if memoHasConflictingMiddleware(memoBz) {
		return commontypes.CompletionHookCall{}, ErrMemoHasConflictingMiddleware
	}
	var memo types.Memo
	err := json.Unmarshal(memoBz, &memo)
	if err != nil {
		return commontypes.CompletionHookCall{}, errors.New("invalid memo")
	}
	if len(memo.OnCompletionHook) == 0 {
		return commontypes.CompletionHookCall{}, errors.New("no completion hook in memo")
	}
	var hook commontypes.CompletionHookCall
	err = proto.Unmarshal(memo.OnCompletionHook, &hook)
	if err != nil {
		return commontypes.CompletionHookCall{}, errors.New("invalid completion hook")
	}
	return hook, nil
}

func memoHasConflictingMiddleware(memoBz []byte) bool {
	d := make(map[string]interface{})
	err := json.Unmarshal(memoBz, &d)
	return err != nil || d[pfmKey] != nil
}
