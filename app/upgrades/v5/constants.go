package v5

import (
	storetypes "cosmossdk.io/store/types"
	circuittypes "cosmossdk.io/x/circuit/types"
	bridgingfeetypes "github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"

	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	hyperwarptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	kastypes "github.com/dymensionxyz/dymension/v3/x/kas/types"
	otcbuybacktypes "github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
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
			kastypes.ModuleName,
			circuittypes.ModuleName,
			ratelimittypes.ModuleName,
			otcbuybacktypes.ModuleName,
			bridgingfeetypes.ModuleName,
		},
	},
}

var CircuitBreakPermissionedAddrs = []string{
	"dym1fljlkdrsp5qpthxcltahjagzkzw6vgqgpzunv8",
	"dym1k5rmr6j8fce7el3hr0t3ujya7nqttzh8kyqp56",
	"dym15vcpft3dwh08c5gr604pmvaukt9lyv4jn6e03t",
}

func CircuitBreakPermissioned(ctx sdk.Context) []string {
	return CircuitBreakPermissionedAddrs
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

const (
	NobleUSDCDenomTN     = "ibc/6490A7EAB61059BFC1CDDEB05917DD70BDF3A611654162A1A47DB930D40D8AF4"
	NobleUSDCChannelIDTN = "channel-3"
)

func NobleUsdcDenom(ctx sdk.Context) string {
	if IsTestnet(ctx) {
		return NobleUSDCDenomTN
	}
	return NobleUSDCDenom
}

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

func IsTestnet(ctx sdk.Context) bool {
	return ctx.ChainID() == "blumbus_111-1"
}

func IbcChannels(ctx sdk.Context) []ratelimittypes.Path {
	if IsTestnet(ctx) {
		return []ratelimittypes.Path{
			{
				Denom:     NobleUSDCDenomTN,
				ChannelId: NobleUSDCChannelIDTN,
			},
		}
	}
	return IBCChannels
}
