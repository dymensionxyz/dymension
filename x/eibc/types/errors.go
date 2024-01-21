package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/eibc module sentinel errors
var (
	ErrInvalidDemandOrderPrice       = sdkerrors.Register(ModuleName, 1, "Price must be greater than 0")
	ErrInvalidDemandOrderFee         = sdkerrors.Register(ModuleName, 2, "Fee must be greater than 0 and less than or equal to the total amount")
	ErrInvalidOrderID                = sdkerrors.Register(ModuleName, 3, "Invalid order ID")
	ErrInvalidAmount                 = sdkerrors.Register(ModuleName, 4, "Invalid amount")
	ErrDemandOrderDoesNotExist       = sdkerrors.Register(ModuleName, 5, "Demand order does not exist")
	ErrDemandOrderInactive           = sdkerrors.Register(ModuleName, 6, "Demand order inactive")
	ErrFullfillerAddressDoesNotExist = sdkerrors.Register(ModuleName, 7, "Fullfiller address does not exist")
	ErrFullfillerInsufficientBalance = sdkerrors.Register(ModuleName, 8, "Fullfiller does not have enough balance")
	ErrInvalidRecipientAddress       = sdkerrors.Register(ModuleName, 9, "Invalid recipient address")
	ErrBlockedAddress                = sdkerrors.Register(ModuleName, 10, "Can't purchase demand order for recipient with blocked address")
)
