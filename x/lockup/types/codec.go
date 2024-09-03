package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
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

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterCodec(amino)
	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	sdk.RegisterLegacyAminoCodec(amino)
	RegisterCodec(authzcodec.Amino)

	amino.Seal()
}
