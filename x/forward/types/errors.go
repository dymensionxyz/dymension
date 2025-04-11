package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type RefundFail struct {
	Addr      string
	Coins     sdk.Coin
	ErrCause  error
	ErrRefund error
}

func (e RefundFail) Unwrap() error {
	return gerrc.ErrInternal
}

func (e RefundFail) Error() string {
	return errorsmod.Wrapf(e.Unwrap(), "refund: addr: %s, coins: %s, errCause: %s, errRefund: %s", e.Addr, e.Coins, e.ErrCause, e.ErrRefund).Error()
}
