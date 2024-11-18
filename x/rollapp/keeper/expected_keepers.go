package keeper

import (
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type ChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}

type SequencerKeeper interface {
	SlashLiveness(ctx sdk.Context, rollappID string) error
	PunishSequencer(ctx sdk.Context, seqAddr string, rewardee *sdk.AccAddress) error
	GetProposer(ctx sdk.Context, rollappId string) types.Sequencer
	GetSuccessor(ctx sdk.Context, rollapp string) types.Sequencer
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

type CanonicalLightClientKeeper interface {
	GetRollappForClientID(ctx sdk.Context, clientID string) (string, bool)
}

type TransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (transfertypes.DenomTrace, bool)
}
