package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSetCanonicalClient{}, "lightclient/SetCanonicalClient", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetCanonicalClient{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	registry.RegisterInterface(
		"ibc.core.client.v1.ClientState",
		(*exported.ClientState)(nil),
	)
}
