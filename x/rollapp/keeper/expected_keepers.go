package keeper

import (
	"context"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// TODO: move to types/expected_keepers.go
type ChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}

type SequencerKeeper interface {
	GetProposer(ctx sdk.Context, rollappId string) types.Sequencer
	GetSuccessor(ctx sdk.Context, rollapp string) types.Sequencer
	SlashLiveness(ctx sdk.Context, rollappID string) error
	PunishSequencer(ctx sdk.Context, seqAddr string, rewardee *sdk.AccAddress) error
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
}

type CanonicalLightClientKeeper interface {
	GetRollappForClientID(ctx sdk.Context, clientID string) (string, bool)
	GetCanonicalClient(ctx sdk.Context, rollappId string) (string, bool)
}

type TransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (transfertypes.DenomTrace, bool)
}
