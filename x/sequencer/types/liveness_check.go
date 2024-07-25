package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type LivenessSlashAndJailArgs struct {
	HHub                      int64
	HNoticeExpired            *int64
	HUpdate                   int64
	HubBlockTime              time.Duration
	SlashTimeNoUpdate         time.Duration
	SlashTimeNoTerminalUpdate time.Duration
	SlashInterval             time.Duration
	SlashMultiplier           sdk.Dec
	JailTime                  time.Duration
	Balance                   sdk.Coins
	MinBond                   sdk.Coins
}

func (a LivenessSlashAndJailArgs) Calculate() (slashAmt sdk.Coins, jail bool) {
	// TODO:
	return sdk.Coins{}, false
}

type LivenessSlashAndJailResult struct {
	Slashed                    sdk.Coins
	Jailed                     bool
	TimeUntilNextSlashPossible time.Time
	FundsReceived              sdk.Coins
}

type LivenessSlashAndJailFundsRecipient struct {
	Multiplier math.LegacyDec // multiplier for slashed funds to send
	Addr       sdk.AccAddress // recipient of reward
}
