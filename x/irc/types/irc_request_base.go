package types

import (
	"encoding/binary"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
)

var _ binary.ByteOrder

var (
	_ codectypes.UnpackInterfacesMessage = IRCRequest{}
	_ IRCRequestI                        = &IRCRequest{}
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
	msgType := getMsgTypeFromMsg(msg)

	if msgType == Unspecified {
		return nil, ErrInvalidMsgType
	}

	return &IRCRequest{
		ReqId:       reqId,
		Message:     any,
		MessageType: getMsgTypeFromMsg(msg),
	}, nil
}

func getMsgTypeFromMsg(msg sdk.Msg) MsgType {
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

// GetMsg implements IRCRequestI
func (m IRCRequest) GetMsg() sdk.Msg {
	if m.Message == nil {
		return nil
	}
	msgVal := m.Message.GetCachedValue()
	if msgVal == nil {
		return nil
	}
	msg := msgVal.(sdk.Msg)
	return msg
	// switch m.MessageType {
	// case CreateClient:
	// 	return msg.(*clienttypes.MsgCreateClient)
	// case UpdateClient:
	// 	return msg.(*clienttypes.MsgUpdateClient)
	// case UpgradeClient:
	// 	return msg.(*clienttypes.MsgUpgradeClient)
	// case SubmitMisbehaviour:
	// 	return msg.(*clienttypes.MsgSubmitMisbehaviour)
	// case ConnectionOpenInit:
	// 	return msg.(*connectiontypes.MsgConnectionOpenInit)
	// case ConnectionOpenTry:
	// 	return msg.(*connectiontypes.MsgConnectionOpenTry)
	// case ConnectionOpenAck:
	// 	return msg.(*connectiontypes.MsgConnectionOpenAck)
	// case ConnectionOpenConfirm:
	// 	return msg.(*connectiontypes.MsgConnectionOpenConfirm)
	// case ChannelOpenInit:
	// 	return msg.(*channeltypes.MsgChannelOpenInit)
	// case ChannelOpenTry:
	// 	return msg.(*channeltypes.MsgChannelOpenTry)
	// case ChannelOpenAck:
	// 	return msg.(*channeltypes.MsgChannelOpenAck)
	// case ChannelOpenConfirm:
	// 	return msg.(*channeltypes.MsgChannelOpenConfirm)
	// case ChannelCloseInit:
	// 	return msg.(*channeltypes.MsgChannelCloseInit)
	// case ChannelCloseConfirm:
	// 	return msg.(*channeltypes.MsgChannelCloseConfirm)
	// case RecvPacket:
	// 	return msg.(*channeltypes.MsgRecvPacket)
	// case Timeout:
	// 	return msg.(*channeltypes.MsgTimeout)
	// case TimeoutOnClose:
	// 	return msg.(*channeltypes.MsgTimeoutOnClose)
	// case Acknowledgement:
	// 	return msg.(*channeltypes.MsgAcknowledgement)
	// default:
	// 	return nil
	// }
}

// Handler implements IRCRequestI
func (m IRCRequest) Handler() error {
	panic("unimplemented")
}
