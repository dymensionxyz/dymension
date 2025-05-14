package types

import (
	"encoding/hex"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgFulfillOrder{}
	_ sdk.Msg = &MsgFulfillOrderAuthorized{}
	_ sdk.Msg = &MsgUpdateDemandOrder{}
	_ sdk.Msg = &MsgTryFulfillOnDemand{}
	_ sdk.Msg = &MsgCreateOnDemandLP{}
	_ sdk.Msg = &MsgDeleteOnDemandLP{}
)

func NewMsgFulfillOrder(fulfillerAddress, orderId, expectedFee string) *MsgFulfillOrder {
	return &MsgFulfillOrder{
		FulfillerAddress: fulfillerAddress,
		OrderId:          orderId,
		ExpectedFee:      expectedFee,
	}
}

func (msg *MsgFulfillOrder) ValidateBasic() error {
	err := validateCommon(msg.OrderId, msg.ExpectedFee, msg.FulfillerAddress)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error()) // TODO: join
	}
	return nil
}

func (msg *MsgFulfillOrder) GetFulfillerBech32Address() []byte {
	return sdk.MustAccAddressFromBech32(msg.FulfillerAddress)
}

func NewMsgFulfillOrderAuthorized(
	orderId,
	rollappId,
	granterAddress,
	operatorFeeAddress,
	expectedFee string,
	price sdk.Coins,
	amount math.Int,
	fulfillerFeePart math.LegacyDec,
	settlementValidated bool,
) *MsgFulfillOrderAuthorized {
	return &MsgFulfillOrderAuthorized{
		OrderId:             orderId,
		RollappId:           rollappId,
		LpAddress:           granterAddress,
		OperatorFeeAddress:  operatorFeeAddress,
		ExpectedFee:         expectedFee,
		Price:               price,
		Amount:              amount,
		OperatorFeeShare:    fulfillerFeePart,
		SettlementValidated: settlementValidated,
	}
}

func (msg *MsgFulfillOrderAuthorized) ValidateBasic() error {
	if err := validateRollappID(msg.RollappId); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid rollapp id")
	}

	if err := validateCommon(msg.OrderId, msg.ExpectedFee, msg.OperatorFeeAddress, msg.LpAddress); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if !msg.Price.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "price is invalid")
	}

	if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount is invalid")
	}

	if msg.OperatorFeeShare.IsNil() || msg.OperatorFeeShare.IsNegative() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "operator fee share cannot be empty or negative")
	}

	if msg.OperatorFeeShare.GT(math.LegacyOneDec()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "operator fee share cannot be greater than 1")
	}

	return nil
}

func (msg *MsgFulfillOrderAuthorized) GetLPBech32Address() []byte {
	return sdk.MustAccAddressFromBech32(msg.LpAddress)
}

func (msg *MsgFulfillOrderAuthorized) GetOperatorFeeBech32Address() []byte {
	return sdk.MustAccAddressFromBech32(msg.OperatorFeeAddress)
}

func NewMsgUpdateDemandOrder(ownerAddr, orderId, newFee string) *MsgUpdateDemandOrder {
	return &MsgUpdateDemandOrder{
		OrderId:      orderId,
		OwnerAddress: ownerAddr,
		NewFee:       newFee,
	}
}

func (m *MsgUpdateDemandOrder) ValidateBasic() error {
	err := validateCommon(m.OrderId, m.NewFee, m.OwnerAddress)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return nil
}

func (m *MsgUpdateDemandOrder) GetSignerAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(m.OwnerAddress)
}

func isValidOrderId(orderId string) bool {
	hashBytes, err := hex.DecodeString(orderId)
	if err != nil {
		// The string is not a valid hexadecimal string
		return false
	}
	// SHA-256 hashes are 32 bytes long
	return len(hashBytes) == 32
}

func validateCommon(orderId, fee string, address ...string) error {
	if !isValidOrderId(orderId) {
		return fmt.Errorf("%w: %s", ErrInvalidOrderID, orderId)
	}

	for _, addr := range address {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return err
		}
	}

	feeInt, ok := math.NewIntFromString(fee)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("parse fee: %s", fee))
	}

	if feeInt.IsNegative() {
		return ErrNegativeFee
	}

	return nil
}

func validateRollappID(rollappID string) error {
	if rollappID == "" {
		return rollapptypes.ErrInvalidRollappID
	}
	_, err := rollapptypes.NewChainID(rollappID)
	return err
}

/////////////////////

func (m *MsgTryFulfillOnDemand) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		return err
	}
	if m.OrderId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty order id")
	}
	return nil
}

func (m *MsgCreateOnDemandLP) ValidateBasic() error {
	if m.Lp == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty lp")
	}
	_, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		return err
	}
	if m.Signer != m.Lp.FundsAddr {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "signer does not match lp address")
	}
	return m.Lp.Validate()
}

func (m *MsgCreateOnDemandLP) MustAcc() sdk.AccAddress {
	a, err := sdk.AccAddressFromBech32(m.Lp.FundsAddr)
	if err != nil {
		panic(err)
	}
	return a
}

func (m *MsgDeleteOnDemandLP) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Signer)
	return err
}

func (m *MsgDeleteOnDemandLP) MustAcc() sdk.AccAddress {
	a, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return a
}

func (m MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "authority must be a valid bech32 address: %s", m.Authority),
		)
	}

	err = m.NewParams.ValidateBasic()
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidRequest,
			errorsmod.Wrapf(err, "failed to validate params"),
		)
	}

	return nil
}
