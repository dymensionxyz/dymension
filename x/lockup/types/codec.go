package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLockTokens{}, "dymensionxyz/dymension/lockup/LockTokens", nil)
	cdc.RegisterConcrete(&MsgBeginUnlockingAll{}, "dymensionxyz/dymension/lockup/BeginUnlockTokens", nil)
	cdc.RegisterConcrete(&MsgBeginUnlocking{}, "dymensionxyz/dymension/lockup/BeginUnlockPeriodLock", nil)
	cdc.RegisterConcrete(&MsgExtendLockup{}, "dymensionxyz/dymension/lockup/ExtendLockup", nil)
	cdc.RegisterConcrete(&MsgForceUnlock{}, "dymensionxyz/dymension/lockup/ForceUnlockTokens", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgLockTokens{},
		&MsgBeginUnlockingAll{},
		&MsgBeginUnlocking{},
		&MsgExtendLockup{},
		&MsgForceUnlock{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
