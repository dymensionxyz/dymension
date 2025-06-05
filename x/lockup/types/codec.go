package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLockTokens{}, "dymensionxyz/dymension/lockup/LockTokens", nil)
	cdc.RegisterConcrete(&MsgBeginUnlocking{}, "dymensionxyz/dymension/lockup/BeginUnlockPeriodLock", nil)
	cdc.RegisterConcrete(&MsgExtendLockup{}, "dymensionxyz/dymension/lockup/ExtendLockup", nil)
	cdc.RegisterConcrete(&MsgForceUnlock{}, "dymensionxyz/dymension/lockup/ForceUnlockTokens", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "lockup/UpdateParams", nil)
	cdc.RegisterConcrete(Params{}, "lockup/Params", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgLockTokens{},
		&MsgBeginUnlocking{},
		&MsgExtendLockup{},
		&MsgForceUnlock{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
