package v5

import (
	storetypes "cosmossdk.io/store/types"
	circuittypes "cosmossdk.io/x/circuit/types"

	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	hyperwarptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
)

const (
	UpgradeName = "v5"
)

var Upgrade = upgrades.Upgrade{
	Name:          UpgradeName,
	CreateHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			hypercoretypes.ModuleName,
			hyperwarptypes.ModuleName,
			circuittypes.ModuleName,
			ratelimittypes.ModuleName,
		},
	},
}

var CircuitBreakPermissioned = []string{
	"dym1fljlkdrsp5qpthxcltahjagzkzw6vgqgpzunv8",
	"dym1k5rmr6j8fce7el3hr0t3ujya7nqttzh8kyqp56",
}

const (
	// Noble USDC. Channel ID is derived from
	// dymd q ibc-transfer denom-trace B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4 --node https://rpc-dymension.mzonder.com/
	NobleUSDCDenom     = "ibc/B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4"
	NobleUSDCChannelID = "channel-6"

	// Kava USDT. Channel ID is derived from
	// dymd q ibc-transfer denom-trace B72B5B3F7AD44783584921DC33354BCE07C8EB0A7F0349247C3DAD38C3B6E6A5 --node https://rpc-dymension.mzonder.com/
	KavaUSDTDenom     = "ibc/B72B5B3F7AD44783584921DC33354BCE07C8EB0A7F0349247C3DAD38C3B6E6A5"
	KavaUSDTChannelID = "channel-3"
)

var IBCChannels = []ratelimittypes.Path{
	{
		Denom:     NobleUSDCDenom,
		ChannelId: NobleUSDCChannelID,
	},
	{
		Denom:     KavaUSDTDenom,
		ChannelId: KavaUSDTChannelID,
	},
}
