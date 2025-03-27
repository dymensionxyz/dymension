package types

import (
	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
)

func (h *HookEIBCtoHL) ValidateBasic() error {
	err := h.Recovery.ValidateBasic()
	if err != nil {
		return err
	}
	return nil
}

// sender is computed
func NewHookEIBCtoHL(
	recovery *Recovery,
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,

	gasLimit math.Int, // can be zero
	customHookId *hyperutil.HexAddress, // optional
	customHookMetadata string, // can be empty
) *HookEIBCtoHL {
	return &HookEIBCtoHL{
		Recovery: recovery,
		HyperlaneTransfer: &warptypes.MsgRemoteTransfer{
			TokenId:            tokenId,
			DestinationDomain:  destinationDomain,
			Recipient:          recipient,
			Amount:             amount,
			CustomHookId:       customHookId,
			GasLimit:           gasLimit,
			MaxFee:             maxFee,
			CustomHookMetadata: customHookMetadata,
		},
	}
}

func (h *HookHLtoIBC) ValidateBasic() error {
	err := h.Transfer.ValidateBasic()
	if err != nil {
		return err
	}
	err = h.Recovery.ValidateBasic()
	if err != nil {
		return err
	}
	return nil
}

// token is computed
// sender is computed
// timeout height not supported
// next memo should go together in the top level of the HL memo
func NewHookHLtoIBC(
	sourcePort string,
	sourceChannel string,
	token sdk.Coin,
	receiver string,
	timeoutTimestamp uint64,
) *HookHLtoIBC {
	return &HookHLtoIBC{
		Transfer: &ibctransfertypes.MsgTransfer{
			SourcePort:       sourcePort,
			SourceChannel:    sourceChannel,
			Token:            token,
			Receiver:         receiver,
			TimeoutTimestamp: timeoutTimestamp,
		},
	}
}

func (r *Recovery) ValidateBasic() error {
	_, err := r.AccAddr()
	if err != nil {
		return err
	}
	return nil
}

func NewRecovery(
	address string,
) *Recovery {
	return &Recovery{
		Address: address,
	}
}

func (r *Recovery) AccAddr() (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(r.Address)
	return addr, err
}

func (r *Recovery) MustAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(r.Address)
}
