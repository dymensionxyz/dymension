package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/edwards"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var (
	_ sdk.Msg                            = &MsgCreateSequencer{}
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateSequencer)(nil)
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateSequencer) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.DymintPubKey, &pubKey)
}

/* --------------------------- MsgCreateSequencer --------------------------- */

func NewMsgCreateSequencer(
	creator string,
	pubkey cryptotypes.PubKey,
	rollappId string,
	metadata *SequencerMetadata,
	bond sdk.Coin,
	rewardAddr string,
	whitelistedRelayers []string,
) (*MsgCreateSequencer, error) {
	if metadata == nil {
		return nil, gerrc.ErrInvalidArgument
	}
	var pkAny *codectypes.Any
	if pubkey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubkey); err != nil {
			return nil, err
		}
	}

	return &MsgCreateSequencer{
		Creator:             creator,
		DymintPubKey:        pkAny,
		RollappId:           rollappId,
		Metadata:            *metadata,
		Bond:                bond,
		RewardAddr:          rewardAddr,
		WhitelistedRelayers: whitelistedRelayers,
	}, nil
}

func (msg *MsgCreateSequencer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAddr, "invalid creator address (%s)", err)
	}

	// public key also checked by the application logic
	if msg.DymintPubKey == nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "sequencer pubkey is required")
	}

	// check it is a pubkey
	if _, err = codectypes.NewAnyWithValue(msg.DymintPubKey); err != nil {
		return errorsmod.Wrapf(ErrInvalidPubKey, "invalid sequencer pubkey(%s)", err)
	}

	// cast to cryptotypes.PubKey type
	pk, ok := msg.DymintPubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return errorsmod.WithType(ErrInvalidPubKey, pk)
	}

	_, err = edwards.ParsePubKey(edwards.Edwards(), pk.Bytes())
	// err means the pubkey validation failed
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidPubKey, "%s", err)
	}

	if err = msg.Metadata.Validate(); err != nil {
		return errorsmod.Wrap(ErrInvalidMetadata, err.Error())
	}

	if !msg.Bond.IsValid() || msg.Bond.IsZero() {
		return errorsmod.Wrapf(ErrInvalidCoins, "invalid bond amount: %s", msg.Bond.String())
	}

	if msg.RewardAddr != "" {
		_, err = sdk.AccAddressFromBech32(msg.RewardAddr)
		if err != nil {
			return errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "get reward addr from bech32")
		}
	}

	err = ValidateWhitelistedRelayers(msg.WhitelistedRelayers)
	if err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "validate whitelisted relayers")
	}

	return nil
}

func (msg *MsgCreateSequencer) VMSpecificValidate(vmType types.Rollapp_VMType) error {
	if vmType == types.Rollapp_EVM {
		if err := validateURLs(msg.Metadata.EvmRpcs); err != nil {
			return errorsmod.Wrap(err, "invalid evm rpcs URLs")
		}
	}
	return nil
}
