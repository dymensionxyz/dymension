package types

import (
	"encoding/binary"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

var _ binary.ByteOrder

var (
	_ codectypes.UnpackInterfacesMessage = IRCRequest{}
)

var MsgTypeStr2MsgType = map[string]MsgType{}

// NewIRCRequest returns a new IRCRequest.
//
//nolint:interfacer
func NewIRCRequest(reqId uint64, msg sdk.Msg) (*IRCRequest, error) {
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}
	msgType := getMsgTypeFromNsg(msg)

	if msgType == Unspecified {
		return nil, ErrInvalidMsgType
	}

	return &IRCRequest{
		ReqId:       reqId,
		Message:     any,
		MessageType: getMsgTypeFromNsg(msg),
	}, nil
}

func getMsgTypeFromNsg(msg sdk.Msg) MsgType {
	msgTypeStr := fmt.Sprintf("%T", msg)
	msgType, exists := MsgTypeStr2MsgType[msgTypeStr]
	if exists {
		return msgType
	}
	switch msg.(type) {
	case *clienttypes.MsgCreateClient:
		MsgTypeStr2MsgType[msgTypeStr] = CreateClient
	case *clienttypes.MsgUpdateClient:
		MsgTypeStr2MsgType[msgTypeStr] = UpdateClient
	case *clienttypes.MsgUpgradeClient:
		MsgTypeStr2MsgType[msgTypeStr] = UpgradeClient
	case *clienttypes.MsgSubmitMisbehaviour:
		MsgTypeStr2MsgType[msgTypeStr] = SubmitMisbehaviour
	case *connectiontypes.MsgConnectionOpenInit:
		MsgTypeStr2MsgType[msgTypeStr] = ConnectionOpenInit
	case *connectiontypes.MsgConnectionOpenTry:
		MsgTypeStr2MsgType[msgTypeStr] = ConnectionOpenTry
	case *connectiontypes.MsgConnectionOpenAck:
		MsgTypeStr2MsgType[msgTypeStr] = ConnectionOpenAck
	case *connectiontypes.MsgConnectionOpenConfirm:
		MsgTypeStr2MsgType[msgTypeStr] = ConnectionOpenConfirm
	case *channeltypes.MsgChannelOpenInit:
		MsgTypeStr2MsgType[msgTypeStr] = ChannelOpenInit
	case *channeltypes.MsgChannelOpenTry:
		MsgTypeStr2MsgType[msgTypeStr] = ChannelOpenTry
	case *channeltypes.MsgChannelOpenAck:
		MsgTypeStr2MsgType[msgTypeStr] = ChannelOpenAck
	case *channeltypes.MsgChannelOpenConfirm:
		MsgTypeStr2MsgType[msgTypeStr] = ChannelOpenConfirm
	case *channeltypes.MsgChannelCloseInit:
		MsgTypeStr2MsgType[msgTypeStr] = ChannelCloseInit
	case *channeltypes.MsgChannelCloseConfirm:
		MsgTypeStr2MsgType[msgTypeStr] = ChannelCloseConfirm
	case *channeltypes.MsgRecvPacket:
		MsgTypeStr2MsgType[msgTypeStr] = RecvPacket
	case *channeltypes.MsgTimeout:
		MsgTypeStr2MsgType[msgTypeStr] = Timeout
	case *channeltypes.MsgTimeoutOnClose:
		MsgTypeStr2MsgType[msgTypeStr] = TimeoutOnClose
	case *channeltypes.MsgAcknowledgement:
		MsgTypeStr2MsgType[msgTypeStr] = Acknowledgement
	default:
		return Unspecified
	}

	return MsgTypeStr2MsgType[msgTypeStr]
}

func (m IRCRequest) UnpackInterfaces(ctx codectypes.AnyUnpacker) error {
	var msg sdk.Msg
	return ctx.UnpackAny(m.Message, &msg)
}
