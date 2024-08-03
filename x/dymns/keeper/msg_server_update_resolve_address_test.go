package keeper_test

import (
	cryptorand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
var nonHostChainBech32InputSet = []string{
	"dym1fl48vsnmsdzcv8",                         // host-chain prefix but invalid bech32 format
	"dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38xuuuu", // host-chain prefix but invalid bech32 checksum
	testAddr(func() uint64 {
		n, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(math.MaxInt64))
		return n.Uint64()
	}() + 9471274174).bech32C("another"),
	"4BDtRc8Ym9wGFyEBzDWMSZ7iuUcNJ1ssiRkU6LjQgHURD4PGAMsZnzxAz2SGmNhinLxPF111N41bTHQBiu6QTmaZwKngDWrH",
	"t1Rv4exT7bqhZqi2j7xz8bUHDMxwosrjADU",
	"zs1z7rejlpsa98s2rrrfkwmaxu53e4ue0ulcrw0h4x5g8jl04tak0d3mm47vdtahatqrlkngh9sly",
	"zcU1Cd6zYyZCd2VJF8yKgmzjxdiiU1rgTTjEwoN1CGUWCziPkUTXUjXmX7TMqdMNsTfuiGN1jQoVN4kGxUR4sAPN4XZ7pxb",
	"XpLM8qBMd7CqukVzKXkQWuQJmgrAFb87Qr",
	"0x7f533b5fbf6ef86c3b7df76cc27fc67744a9a760",
	"2UEQTE5QDNXPI7M3TU44G6SYKLFWLPQO7EBZM7K7MHMQQMFI4QJPLHQFHM",
	"ALGO-2UEQTE5QDNXPI7M3TU44G6SYKLFWLPQO7EBZM7K7MHMQQMFI4QJPLHQFHM",
	"0.0.123",
	"0.0.0",
	"0.0.123-vfmkw",
	"LMHEFMwRsQ3nHDfb9zZqynLHxjuJ2hgyyW",
	"MC2JYMPVWaxqUb9qUkUbjtUwoNMo1tPaLF",
	"ltc1qhzjptwpym9afcdjhs7jcz6fd0jma0l0rc0e5yr",
	"ltc1qzvcgmntglcuv4smv3lzj6k8szcvsrmvk0phrr9wfq8w493r096ssm2fgsw",
	"qrvax3jgtwqssnkpctlqdl0rq7rjn0l0hgny8pt0hp",
	"bitcoincash:qrvax3jgtwqssnkpctlqdl0rq7rjn0l0hgny8pt0hp",
	"D7wbmbjBWG5HPkT6d4gh6SdQPp6z25vcF2",
	"0xBe588061d20fe359E69D78824EC45EA98C87069A",
	"NVeu7XqbZ6WiL1prhChC1jMWgicuWtneDP",
	"ALuhj3QNoxvAnMZsA2oKP5UxYsBmRwjwHL",
	"tz1YWK1gDPQx9N1Jh4JnmVre7xN6xhGGM4uC",
	"tz3T8djchG5FDwt7H6wEUU3sRFJwonYPqMJe",
	"KT1S5hgipNSTFehZo7v81gq6fcLChbRwptqy",
	"rpshnaf39wBUDNEGHJKLM4PQRST7VWXYZ2bcdeCg65jkm8oFqi1tuvAxyz",
	"XV5sbjUmgPpvXv4ixFWZ5ptAYZ6PD28Sq49uo34VyjnmK5H",
	"7EcDhSYGxXyscszYEp35KHN8vvw3svAuLKTzXwCFLtV",
	"414450cf8c8b6a8229b7f628e36b3a658e84441b6f",
	"TGCRkw1Vq759FBCrwxkZGgqZbRX1WkBHSu",
	"xdc64b3b0a417775cfb441ed064611bf79826649c0f",
	"0x64b3b0a417775cfb441ed064611bf79826649c0f",
	"GBH4TZYZ4IRCPO44CBOLFUHULU2WGALXTAVESQA6432MBJMABBB4GIYI",
	"jed*stellar.org",
	"maria@gmail.com*stellar.org",
	"bc1qeklep85ntjz4605drds6aww9u0qr46qzrv5xswd35uhjuj8ahfcqgf6hak",
	"bc1pxwww0ct9ue7e8tdnlmug5m2tamfn7q06sahstg39ys4c9f3340qqxrdu9k",
	"bc1prwgcpptoxrpfl5go81wpd5qlsig5yt4g7urb45e",
	"bc1qwqdg6squsna38e46795at95yu9atm8azzmyvckulcc7kytlcckxswvvzej",
	"0x3cA8ac240F6ebeA8684b3E629A8e8C1f0E3bC0Ff",
	"X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5",
	"Ae2tdPwUPEZFSi1cTyL1ZL6bgixhc2vSy5heg6Zg9uP7PpumkAJ82Qprt8b",
	"DdzFFzCqrhsfZHjaBunVySZBU8i9Zom7Gujham6Jz8scCcAdkDmEbD9XSdXKdBiPoa1fjgL4ksGjQXD8ZkSNHGJfT25ieA9rWNCSA5qc",
	"addr1q8gg2r3vf9zggn48g7m8vx62rwf6warcs4k7ej8mdzmqmesj30jz7psduyk6n4n2qrud2xlv9fgj53n6ds3t8cs4fvzs05yzmz",
	"1a1LcBX6hGPKg5aQ6DXZpAHCCzWjckhea4sz3P1PvL3oc4F",
	"HNZata7iMYWmk5RvZRTiAsSDhV8366zq2YGb3tLH5Upf74F",
	"5CdiCGvTEuzut954STAXRfL8Lazs3KCZa5LPpkPeqqJXdTHp",
	"0x192c3c7e5789b461fbf1c7f614ba5eed0b22efc507cda60a5e7fda8e046bcdce",
	"0x0380d46a00e427d89f35d78b4eacb4270bd5ecfd10b64662dcfe31eb117fc62c68",
	"04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f",
	"11111111111111111111BZbvjr",
	"1111111111111111111114oLvT2",
	"12higDjoCCNXSA95xZMWUdPvXNmkAduhWv",
	"342ftSRCvFHfCeFFBuz4xwbeqnDw6BGUey",
	"bc1q34aq5drpuwy3wgl9lhup9892qp6svr8ldzyy7c",
}

var nonBech32InputSet []string

func init() {
	for _, input := range nonHostChainBech32InputSet {
		if !dymnsutils.IsValidBech32AccountAddress(input, false) {
			nonBech32InputSet = append(nonBech32InputSet, input)
		}
	}
}

func Test_msgServer_UpdateResolveAddress(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	setupTest := func() (dymnskeeper.Keeper, rollappkeeper.Keeper, sdk.Context) {
		dk, _, rk, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now).WithChainID(chainId)

		return dk, rk, ctx
	}

	ownerAcc := testAddr(1)
	controllerAcc := testAddr(2)
	anotherAcc := testAddr(14)
	_32BytesAcc := testAddr(15)

	const recordName = "my-name"

	const rollAppId = "ra_9999-1"

	//goland:noinspection SpellCheckingInspection
	nonBech32NonHexUpperCaseA := strings.ToUpper("X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5")

	params.SetAddressPrefixes()

	type testSuite struct {
		t   *testing.T
		dk  dymnskeeper.Keeper
		rk  rollappkeeper.Keeper
		ctx sdk.Context
	}

	nts := func(t *testing.T, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper, ctx sdk.Context) testSuite {
		return testSuite{
			t:   t,
			dk:  dk,
			rk:  rk,
			ctx: ctx,
		}
	}

	requireConfiguredAddressMappedDymNames := func(ts testSuite, cfgAddr string, names ...string) {
		dymNames, err := ts.dk.GetDymNamesContainsConfiguredAddress(ts.ctx, cfgAddr)
		require.NoError(ts.t, err)
		require.Len(ts.t, dymNames, len(names))
		sort.Strings(names)
		sort.Slice(dymNames, func(i, j int) bool {
			return dymNames[i].Name < dymNames[j].Name
		})
		for i, name := range names {
			require.Equal(ts.t, name, dymNames[i].Name)
		}
	}

	requireConfiguredAddressMappedNoDymName := func(ts testSuite, cfgAddr string) {
		requireConfiguredAddressMappedDymNames(ts, cfgAddr)
	}

	requireFallbackAddressMappedDymNames := func(ts testSuite, fallbackAddr dymnstypes.FallbackAddress, names ...string) {
		dymNames, err := ts.dk.GetDymNamesContainsFallbackAddress(ts.ctx, fallbackAddr)
		require.NoError(ts.t, err)
		require.Len(ts.t, dymNames, len(names))
		sort.Strings(names)
		sort.Slice(dymNames, func(i, j int) bool {
			return dymNames[i].Name < dymNames[j].Name
		})
		for i, name := range names {
			require.Equal(ts.t, name, dymNames[i].Name)
		}
	}

	requireFallbackAddressMappedNoDymName := func(ts testSuite, fallbackAddr dymnstypes.FallbackAddress) {
		requireFallbackAddressMappedDymNames(ts, fallbackAddr)
	}

	tests := []struct {
		name               string
		dymName            *dymnstypes.DymName
		msg                *dymnstypes.MsgUpdateResolveAddress
		preTestFunc        func(ts testSuite)
		wantErr            bool
		wantErrContains    string
		wantDymName        *dymnstypes.DymName
		wantMinGasConsumed sdk.Gas
		postTestFunc       func(ts testSuite)
	}{
		{
			name: "fail - reject if message not pass validate basic",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			wantErr:         true,
			wantErrContains: gerrc.ErrInvalidArgument.Error(),
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name:    "fail - Dym-Name does not exists",
			dymName: nil,
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: fmt.Sprintf("Dym-Name: %s: not found", recordName),
			wantDymName:     nil,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if Dym-Name expired",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() - 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "Dym-Name is already expired",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() - 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if sender is neither owner nor controller",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: anotherAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "permission denied",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if sender is owner but not controller",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: ownerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "please use controller account",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if config is not valid",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "0x1",
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "config is invalid:",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if config is not valid. Only accept lowercase",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				SubName:    "SUB", // upper-case is not accepted
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "config is invalid:",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - can update",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - address on RollApp automatically lowercase",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, anotherAcc.bech32C("rol"))
				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())

				ts.rk.SetRollapp(ts.ctx, rollapptypes.Rollapp{
					RollappId: rollAppId,
					Creator:   anotherAcc.bech32(),
				})
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    rollAppId,
				ResolveTo:  strings.ToUpper(anotherAcc.bech32C("rol")), // upper-cased
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: rollAppId,
						Path:    "",
						Value:   anotherAcc.bech32C("rol"),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, anotherAcc.bech32C("rol"), recordName)
				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())
			},
		},
		{
			name: "pass - keep case-sensitive address on non-host/non-RollApp",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				require.Equal(t,
					nonBech32NonHexUpperCaseA, strings.ToUpper(nonBech32NonHexUpperCaseA),
					"bad setup, this address must be upper-cased, to be used in this testcase",
				)

				requireConfiguredAddressMappedNoDymName(ts, nonBech32NonHexUpperCaseA)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "another",
				ResolveTo:  nonBech32NonHexUpperCaseA, // this address is neither bech32 nor hex
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "another",
						Path:    "",
						Value:   nonBech32NonHexUpperCaseA,
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, nonBech32NonHexUpperCaseA, recordName)
			},
		},
		{
			name: "pass - add new record if not exists",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - override record if exists",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - remove record if new resolve to empty, single-config",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "a",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs:    nil,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - remove record if new resolve to empty, single-config, not match any",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "non-exists",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - remove record if new resolve to empty, multi-config, first",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "a",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
		},
		{
			name: "pass - remove record if new resolve to empty, multi-configs, last",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - remove record if new resolve to empty, multi-config, not any of existing",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  "",
				SubName:    "non-exists",
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "a",
						Value: ownerAcc.bech32(),
					},
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: controllerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
			},
		},
		{
			name: "pass - expiry not changed",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 99,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 99,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - chain-id automatically removed from record if is host chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    chainId,
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // empty
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - chain-id automatically removed from record if is host chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // originally empty
						Path:    "a",
						Value:   controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    chainId,
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // empty
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - chain-id recorded if is NOT host chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "blumbus_100-1",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - do not override record with different chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "blumbus_100-1",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - do not override record with different chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   controllerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   controllerAcc.bech32(),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "blumbus_100-1",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   controllerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_100-1",
						Path:    "a",
						Value:   ownerAcc.bech32(),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "pass - if input is 20 bytes, hex address, lower-case when persist",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, anotherAcc.checksumHex())
				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())

				require.NotEqual(ts.t, strings.ToLower(anotherAcc.hexStr()), anotherAcc.checksumHex())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "ethereum",
				ResolveTo:  anotherAcc.checksumHex(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "ethereum",
						Value:   strings.ToLower(anotherAcc.checksumHex()), // lower-cased
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				// should be able to search case-insensitive
				requireConfiguredAddressMappedDymNames(ts, anotherAcc.checksumHex(), recordName)
				requireConfiguredAddressMappedDymNames(ts, strings.ToLower(anotherAcc.checksumHex()), recordName)
				requireConfiguredAddressMappedDymNames(ts, "0x"+strings.ToUpper(anotherAcc.checksumHex()[2:]), recordName)

				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())
			},
		},
		{
			name: "pass - if input is 32 bytes, hex address, lower-case when persist",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, "0x"+strings.ToLower(_32BytesAcc.hexStr()[2:]))
				requireConfiguredAddressMappedNoDymName(ts, "0x"+strings.ToUpper(_32BytesAcc.hexStr()[2:]))
				requireFallbackAddressMappedNoDymName(ts, _32BytesAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "another",
				ResolveTo:  "0x" + strings.ToUpper(_32BytesAcc.hexStr()[2:]), // upper-cased
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "another",
						Value:   "0x" + strings.ToLower(_32BytesAcc.hexStr()[2:]), // lower-cased
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				// should be able to search case-insensitive
				lowerCased := "0x" + strings.ToLower(_32BytesAcc.hexStr()[2:])
				requireConfiguredAddressMappedDymNames(ts, lowerCased, recordName)
				upperCased := "0x" + strings.ToUpper(_32BytesAcc.hexStr()[2:])
				requireConfiguredAddressMappedDymNames(ts, upperCased, recordName)

				requireFallbackAddressMappedNoDymName(ts, _32BytesAcc.fallback())
			},
		},
		{
			name: "pass - if input is 20 bytes, WITHOUT 0x hex address, keep case-sensitive when persist",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, strings.ToUpper(anotherAcc.hexStr()[2:]))
				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "another",
				ResolveTo:  strings.ToUpper(anotherAcc.hexStr()[2:]), // removed 0x part and upper-cased
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "another",
						Value:   strings.ToUpper(anotherAcc.hexStr()[2:]), // keep as is
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, strings.ToUpper(anotherAcc.hexStr()[2:]), recordName)
				requireConfiguredAddressMappedNoDymName(ts, strings.ToLower(anotherAcc.hexStr()[2:]))

				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())

				// dont returns for similar address (+0x)
				requireConfiguredAddressMappedNoDymName(ts, anotherAcc.hexStr())
				requireConfiguredAddressMappedNoDymName(ts, anotherAcc.checksumHex())
			},
		},
		{
			name: "pass - if input is 32 bytes, WITHOUT 0x hex address, keep case-sensitive when persist",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, strings.ToUpper(_32BytesAcc.hexStr()[2:]))
				requireFallbackAddressMappedNoDymName(ts, _32BytesAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "another",
				ResolveTo:  strings.ToUpper(_32BytesAcc.hexStr()[2:]), // removed 0x part and upper-cased
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "another",
						Value:   strings.ToUpper(_32BytesAcc.hexStr()[2:]), // keep as is
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, strings.ToUpper(_32BytesAcc.hexStr()[2:]), recordName)
				requireConfiguredAddressMappedNoDymName(ts, strings.ToLower(_32BytesAcc.hexStr()[2:]))

				requireFallbackAddressMappedNoDymName(ts, _32BytesAcc.fallback())

				// dont returns for similar address (+0x)
				requireConfiguredAddressMappedNoDymName(ts, _32BytesAcc.hexStr())
				requireConfiguredAddressMappedNoDymName(ts, "0x"+strings.ToUpper(_32BytesAcc.hexStr()[2:]))
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 on host chain if target chain is host chain, case empty chain-id",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32C("nim"), // owner but with nim prefix
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 on host chain if target chain is host chain, case use chain-id in request",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    chainId,
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32C("nim"), // owner but with nim prefix
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32C("nim"))
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 on host chain if target chain is host chain, case dym prefix but valoper, not acc addr",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32Valoper(), // owner but with valoper prefix
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsumed: 1,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32Valoper())
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
			},
		},
		{
			name: "fail - reject if address is not corresponding bech32 if target chain is RollApp",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				ts.rk.SetRollapp(ts.ctx, rollapptypes.Rollapp{
					RollappId: rollAppId,
					Creator:   anotherAcc.bech32(),
				})
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    rollAppId,
				SubName:    "a",
				ResolveTo:  ownerAcc.hexStr(), // wrong format
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on RollApp",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsumed: 1,
			postTestFunc:       func(ts testSuite) {},
		},
		{
			name: "fail - reject if address is not corresponding bech32 if target chain is RollApp",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				registerRollApp(ts.t, ts.ctx, ts.rk, ts.dk, "nim_1122-1", "nim", "nim")
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "nim_1122-1",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32C("ma"), // wrong bech32 prefix
				Controller: controllerAcc.bech32(),
			},
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on RollApps",
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			wantMinGasConsumed: 1,
			postTestFunc:       func(ts testSuite) {},
		},
		{
			name: "pass - accept if address is corresponding bech32 if target chain is RollApp",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			},
			preTestFunc: func(ts testSuite) {
				ts.rk.SetRollapp(ts.ctx, rollapptypes.Rollapp{
					RollappId: "nim_1122-1",
					Creator:   anotherAcc.bech32(),
				})
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "nim_1122-1",
				SubName:    "a",
				ResolveTo:  ownerAcc.bech32C("nim"),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "a",
						Value:   ownerAcc.bech32C("nim"),
					},
				},
			},
			wantMinGasConsumed: 1,
			postTestFunc:       func(ts testSuite) {},
		},
		{
			name: "pass - reverse mapping record should be updated accordingly",
			dymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   controllerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "a",
						Value:   anotherAcc.bech32C("nim"),
					},
				},
			},
			preTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedNoDymName(ts, ownerAcc.bech32())
				requireConfiguredAddressMappedDymNames(ts, controllerAcc.bech32(), recordName)
				requireConfiguredAddressMappedDymNames(ts, anotherAcc.bech32C("nim"), recordName)
				requireFallbackAddressMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddressMappedDymNames(ts, controllerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())
			},
			msg: &dymnstypes.MsgUpdateResolveAddress{
				ChainId:    "",
				SubName:    "",
				ResolveTo:  ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
			},
			wantErr: false,
			wantDymName: &dymnstypes.DymName{
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerAcc.bech32(),
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "a",
						Value:   anotherAcc.bech32C("nim"),
					},
				},
			},
			wantMinGasConsumed: dymnstypes.OpGasConfig,
			postTestFunc: func(ts testSuite) {
				requireConfiguredAddressMappedDymNames(ts, ownerAcc.bech32(), recordName)
				requireConfiguredAddressMappedNoDymName(ts, controllerAcc.bech32())
				requireConfiguredAddressMappedDymNames(ts, anotherAcc.bech32C("nim"), recordName)
				requireFallbackAddressMappedDymNames(ts, ownerAcc.fallback(), recordName)
				requireFallbackAddressMappedNoDymName(ts, controllerAcc.fallback())
				requireFallbackAddressMappedNoDymName(ts, anotherAcc.fallback())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.preTestFunc)
			require.NotNil(t, tt.postTestFunc)

			dk, rk, ctx := setupTest()

			if tt.dymName != nil {
				if tt.dymName.Name == "" {
					tt.dymName.Name = recordName
				}
				err := dk.SetDymName(ctx, *tt.dymName)
				require.NoError(t, err)
				require.NoError(t, dk.AfterDymNameOwnerChanged(ctx, tt.dymName.Name))
				require.NoError(t, dk.AfterDymNameConfigChanged(ctx, tt.dymName.Name))
			}
			if tt.wantDymName != nil && tt.wantDymName.Name == "" {
				tt.wantDymName.Name = recordName
			}

			tt.preTestFunc(nts(t, dk, rk, ctx))

			tt.msg.Name = recordName
			resp, err := dymnskeeper.NewMsgServerImpl(dk).UpdateResolveAddress(ctx, tt.msg)
			laterDymName := dk.GetDymName(ctx, tt.msg.Name)

			defer func() {
				if tt.wantMinGasConsumed > 0 {
					require.GreaterOrEqual(t,
						ctx.GasMeter().GasConsumed(), tt.wantMinGasConsumed,
						"should consume at least %d gas", tt.wantMinGasConsumed,
					)
				}

				if !t.Failed() {
					tt.postTestFunc(nts(t, dk, rk, ctx))
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				require.Nil(t, resp)

				if tt.wantDymName != nil {
					require.Equal(t, *tt.wantDymName, *laterDymName)

					owned, err := dk.GetDymNamesOwnedBy(ctx, laterDymName.Owner)
					require.NoError(t, err)
					if laterDymName.ExpireAt >= now.Unix() {
						require.Len(t, owned, 1)
					} else {
						require.Empty(t, owned)
					}
				} else {
					require.Nil(t, laterDymName)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, laterDymName)
			require.Equal(t, *tt.wantDymName, *laterDymName)
		})
	}

	for _, input := range nonHostChainBech32InputSet {
		t.Run("non-bech32/non-hex on non-host/non-RollApp chain: "+input, func(t *testing.T) {
			dk, _, ctx := setupTest()
			ctx = ctx.WithBlockTime(now).WithChainID(chainId)

			const anotherChainId = "another"

			dymName := dymnstypes.DymName{
				Name:       "a",
				Owner:      ownerAcc.bech32(),
				Controller: controllerAcc.bech32(),
				ExpireAt:   now.Unix() + 1,
			}
			require.NoError(t, dk.SetDymName(ctx, dymName))

			resp, err := dymnskeeper.NewMsgServerImpl(dk).UpdateResolveAddress(ctx, &dymnstypes.MsgUpdateResolveAddress{
				Name:       dymName.Name,
				Controller: dymName.Controller,
				ChainId:    anotherChainId,
				SubName:    "",
				ResolveTo:  input,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)

			wantRecordedValue := input
			if dymnsutils.IsValidHexAddress(input) {
				// if input is hex, lower-case it regardless chain
				wantRecordedValue = strings.ToLower(input)
			}

			laterDymName := dk.GetDymName(ctx, dymName.Name)
			require.NotNil(t, laterDymName)
			require.Equal(t, dymnstypes.DymName{
				Name:       dymName.Name,
				Owner:      dymName.Owner,
				Controller: dymName.Controller,
				ExpireAt:   dymName.ExpireAt,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: anotherChainId,
						Path:    "",
						Value:   wantRecordedValue,
					},
				},
			}, *laterDymName)

			dymNameAddress := fmt.Sprintf("%s@%s", dymName.Name, anotherChainId)
			outputAddress, err := dk.ResolveByDymNameAddress(ctx, dymNameAddress)
			require.NoError(t, err)
			require.Equal(t, wantRecordedValue, outputAddress)

			list, err := dk.ReverseResolveDymNameAddress(ctx, input, anotherChainId)
			require.NoError(t, err)
			require.Len(t, list, 1)
			require.Equal(t, dymNameAddress, list[0].String())

			list, err = dk.ReverseResolveDymNameAddress(ctx, input, chainId)
			require.True(t, err != nil || len(list) == 0)
		})
	}
}

func Test_msgServer_UpdateResolveAddress_ReverseMapping(t *testing.T) {
	ownerAcc := testAddr(1)
	anotherAcc := testAddr(2)

	const chainId = "dymension_1100-1"
	const rollappChainId = "rollapp_1-1"
	const rollAppBech32 = "rol"
	const externalChainId = "awesome"
	const name = "my-name"
	const subName = "sub"

	params.SetAddressPrefixes()

	const (
		tcCfgAddr = iota
		tcFallbackAddr
		tcResolveAddr
		tcReverseResolveAddr
	)
	type tc struct {
		_type int
		input string
		want  any
	}
	testMapCfgAddrToDymName := func(input string, wantMapped bool) tc {
		return tc{_type: tcCfgAddr, input: input, want: wantMapped}
	}
	testMapFallbackAddrToDymName := func(input string, wantMapped bool) tc {
		return tc{_type: tcFallbackAddr, input: input, want: wantMapped}
	}
	testResolveAddr := func(input, want string) tc {
		return tc{_type: tcResolveAddr, input: input, want: want}
	}
	testReverseResolveAddr := func(input, want string) tc {
		return tc{_type: tcReverseResolveAddr, input: input, want: want}
	}

	type testStruct struct {
		name                   string
		inputResolveTo         string
		multipleInputResolveTo []string
		hostChain              bool
		rollapp                bool
		rollappWithBech32      bool
		externalChain          bool
		useSubName             bool
		wantReject             bool
		tests                  []tc
	}

	tests := []testStruct{
		{
			name:           "bech32 on host-chain, without sub-name",
			inputResolveTo: anotherAcc.bech32(),
			hostChain:      true,
			useSubName:     false,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.bech32(), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), true), // cuz host-chain and default config
				testResolveAddr(name+"@"+chainId, anotherAcc.bech32()),
				testReverseResolveAddr(anotherAcc.bech32(), name+"@"+chainId),
				testReverseResolveAddr(anotherAcc.hexStr(), name+"@"+chainId),
			},
		},
		{
			name:           "bech32 on host-chain, with sub-name",
			inputResolveTo: anotherAcc.bech32(),
			hostChain:      true,
			useSubName:     true,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.bech32(), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), false), // cuz sub-name, not default config
				testResolveAddr(subName+"."+name+"@"+chainId, anotherAcc.bech32()),
				testReverseResolveAddr(anotherAcc.bech32(), subName+"."+name+"@"+chainId),

				testReverseResolveAddr(anotherAcc.hexStr(), subName+"."+name+"@"+chainId),
				// reverse-resolve-able cuz it's host-chain or RollApp with bech32 configured
			},
		},
		{
			name:              "bech32 on RollApp, without sub-name, without bech32 prefix cfg",
			inputResolveTo:    anotherAcc.bech32(),
			rollapp:           true,
			rollappWithBech32: false,
			useSubName:        false,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.bech32(), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), false), // cuz not host-chain
				testResolveAddr(name+"@"+rollappChainId, anotherAcc.bech32()),
				testReverseResolveAddr(anotherAcc.bech32(), name+"@"+rollappChainId),
				testReverseResolveAddr(anotherAcc.hexStr(), ""),
			},
		},
		{
			name:              "bech32 on RollApp, with sub-name, without bech32 prefix cfg",
			inputResolveTo:    anotherAcc.bech32(),
			rollapp:           true,
			rollappWithBech32: false,
			useSubName:        true,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.bech32(), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), false), // cuz not host-chain
				testResolveAddr(subName+"."+name+"@"+rollappChainId, anotherAcc.bech32()),
				testReverseResolveAddr(anotherAcc.bech32(), subName+"."+name+"@"+rollappChainId),
				testReverseResolveAddr(anotherAcc.hexStr(), ""),
			},
		},
		{
			name:              "bech32 on RollApp, without sub-name, with bech32 prefix cfg",
			inputResolveTo:    anotherAcc.bech32C(rollAppBech32),
			rollapp:           true,
			rollappWithBech32: true,
			useSubName:        false,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.bech32C(rollAppBech32), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), false), // cuz not host-chain
				testResolveAddr(name+"@"+rollappChainId, anotherAcc.bech32C(rollAppBech32)),
				testReverseResolveAddr(anotherAcc.bech32C(rollAppBech32), name+"@"+rollappChainId),
				testReverseResolveAddr(anotherAcc.hexStr(), name+"@"+rollappChainId), // cuz it's RollApp with bech32 prefix configured
			},
		},
		{
			name:              "bech32 on RollApp with sub-name, with bech32 prefix cfg",
			inputResolveTo:    anotherAcc.bech32C(rollAppBech32),
			rollapp:           true,
			rollappWithBech32: true,
			useSubName:        true,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.bech32C(rollAppBech32), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), false), // cuz not host-chain
				testResolveAddr(subName+"."+name+"@"+rollappChainId, anotherAcc.bech32C(rollAppBech32)),
				testReverseResolveAddr(anotherAcc.bech32C(rollAppBech32), subName+"."+name+"@"+rollappChainId),
				testReverseResolveAddr(anotherAcc.hexStr(), subName+"."+name+"@"+rollappChainId), // cuz it's RollApp with bech32 prefix configured
			},
		},
		{
			name:           "bech32 on external-chain, without sub-name",
			inputResolveTo: anotherAcc.bech32(),
			externalChain:  true,
			useSubName:     false,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.bech32(), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), false), // cuz not host-chain
				testResolveAddr(name+"@"+externalChainId, anotherAcc.bech32()),
				testReverseResolveAddr(anotherAcc.bech32(), name+"@"+externalChainId),
				testReverseResolveAddr(anotherAcc.hexStr(), ""),
			},
		},
		{
			name:           "bech32 on external-chain, with sub-name",
			inputResolveTo: anotherAcc.bech32(),
			externalChain:  true,
			useSubName:     true,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.bech32(), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), false), // cuz not host-chain
				testResolveAddr(subName+"."+name+"@"+externalChainId, anotherAcc.bech32()),
				testReverseResolveAddr(anotherAcc.bech32(), subName+"."+name+"@"+externalChainId),
				testReverseResolveAddr(anotherAcc.hexStr(), ""),
			},
		},
		{
			name:                   "non-bech32 on host-chain, without sub-name",
			inputResolveTo:         anotherAcc.hexStr(),
			multipleInputResolveTo: nonHostChainBech32InputSet,
			hostChain:              true,
			useSubName:             false,
			wantReject:             true, // host-chain requires bech32 as input
		},
		{
			name:                   "non-bech32 on host-chain, with sub-name",
			inputResolveTo:         anotherAcc.hexStr(),
			multipleInputResolveTo: nonHostChainBech32InputSet,
			hostChain:              true,
			useSubName:             true,
			wantReject:             true, // host-chain, requires bech32 as input
		},
		{
			name:                   "non-bech32 on RollApp, without sub-name",
			inputResolveTo:         anotherAcc.hexStr(),
			multipleInputResolveTo: nonBech32InputSet,
			rollapp:                true,
			useSubName:             false,
			wantReject:             true, // RollApp requires bech32 as input
		},
		{
			name:                   "non-bech32 on RollApp, with sub-name",
			inputResolveTo:         anotherAcc.hexStr(),
			multipleInputResolveTo: nonBech32InputSet,
			rollapp:                true,
			useSubName:             true,
			wantReject:             true, // RollApp requires bech32 as input
		},
		{
			name:           "hex on external chain, without sub-name",
			inputResolveTo: anotherAcc.hexStr(),
			externalChain:  true,
			useSubName:     false,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.hexStr(), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), false), // cuz not host-chain, not default config
				testResolveAddr(name+"@"+externalChainId, anotherAcc.hexStr()),
				testReverseResolveAddr(anotherAcc.hexStr(), name+"@"+externalChainId),
				testReverseResolveAddr(anotherAcc.bech32(), ""), // cuz input is hex
			},
		},
		{
			name:           "hex on external chain, with sub-name",
			inputResolveTo: anotherAcc.hexStr(),
			externalChain:  true,
			useSubName:     true,
			tests: []tc{
				testMapCfgAddrToDymName(anotherAcc.hexStr(), true),
				testMapFallbackAddrToDymName(anotherAcc.hexStr(), false), // cuz not host-chain, not default config
				testResolveAddr(subName+"."+name+"@"+externalChainId, anotherAcc.hexStr()),
				testReverseResolveAddr(anotherAcc.hexStr(), subName+"."+name+"@"+externalChainId),
				testReverseResolveAddr(anotherAcc.bech32(), ""), // cuz input is hex
			},
		},
	}

	// build test cases from non-bech32 set
	for _, input := range nonHostChainBech32InputSet {
		if dymnsutils.IsValidHexAddress(input) {
			continue
		}
		tests = append(
			tests,
			testStruct{
				name:           fmt.Sprintf("non-bech32 on external chain, without sub-name: %s", input),
				inputResolveTo: input,
				externalChain:  true,
				useSubName:     false,
				tests: []tc{
					testMapCfgAddrToDymName(input, true),
					testResolveAddr(name+"@"+externalChainId, input),
					testReverseResolveAddr(input, name+"@"+externalChainId),
				},
			},
			testStruct{
				name:           fmt.Sprintf("non-bech32 on external chain, with sub-name: %s", input),
				inputResolveTo: input,
				externalChain:  true,
				useSubName:     true,
				tests: []tc{
					testMapCfgAddrToDymName(input, true),
					testResolveAddr(subName+"."+name+"@"+externalChainId, input),
					testReverseResolveAddr(input, subName+"."+name+"@"+externalChainId),
				},
			},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bti := func(b bool) int {
				if b {
					return 1
				}
				return 0
			}
			require.Equal(t,
				1, bti(tt.hostChain)+bti(tt.rollapp)+bti(tt.externalChain),
				"at least one and only one flag is allowed",
			)
			if len(tt.multipleInputResolveTo) > 0 {
				require.True(t, tt.wantReject, "multiple input resolve-to only be used with want-reject")
			}

			dk, _, rk, ctx := testkeeper.DymNSKeeper(t)
			ctx = ctx.WithChainID(chainId)

			dymName := dymnstypes.DymName{
				Name:       name,
				Owner:      ownerAcc.bech32(),
				Controller: ownerAcc.bech32(),
				ExpireAt:   ctx.BlockTime().Add(time.Second).Unix(),
			}
			setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

			if tt.rollapp {
				bech32Prefix := ""
				if tt.rollappWithBech32 {
					bech32Prefix = rollAppBech32
				}

				registerRollApp(t, ctx, rk, dk, rollappChainId, bech32Prefix, "")
			}

			var useContextChainId string
			if tt.hostChain {
				useContextChainId = chainId
			} else if tt.rollapp {
				useContextChainId = rollappChainId
			} else {
				useContextChainId = externalChainId
			}

			var useSubName string
			if tt.useSubName {
				useSubName = subName
			}

			msg := &dymnstypes.MsgUpdateResolveAddress{
				Name:       dymName.Name,
				Controller: ownerAcc.bech32(),
				ChainId:    useContextChainId,
				SubName:    useSubName,
				ResolveTo:  tt.inputResolveTo,
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).UpdateResolveAddress(ctx, msg)
			if tt.wantReject {
				require.Error(t, err)

				for _, input := range tt.multipleInputResolveTo {
					msg.ResolveTo = input
					_, err := dymnskeeper.NewMsgServerImpl(dk).UpdateResolveAddress(ctx, msg)
					require.Errorf(t, err, "input: %s", input)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			{
				// check Dym-Name record

				laterDymName := dk.GetDymName(ctx, dymName.Name)
				require.NotNil(t, laterDymName)

				wantDymName := dymName
				wantDymName.Configs = []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: msg.ChainId,
						Path:    msg.SubName,
						Value:   msg.ResolveTo,
					},
				}
				if tt.hostChain {
					wantDymName.Configs[0].ChainId = ""
				}
				require.Equal(t, wantDymName, *laterDymName)
			}

			require.NotEmpty(t, tt.tests)

			for _, tc := range tt.tests {
				switch tc._type {
				case tcCfgAddr:
					list, err := dk.GetDymNamesContainsConfiguredAddress(ctx, tc.input)
					require.NoError(t, err)
					if tc.want.(bool) {
						requireDymNameList(list, []string{dymName.Name}, t)
					} else {
						require.Empty(t, list)
					}
				case tcFallbackAddr:
					list, err := dk.GetDymNamesContainsFallbackAddress(ctx, dymnsutils.GetBytesFromHexAddress(tc.input))
					require.NoError(t, err)
					if tc.want.(bool) {
						requireDymNameList(list, []string{dymName.Name}, t)
					} else {
						require.Empty(t, list)
					}
				case tcResolveAddr:
					outputAddr, err := dk.ResolveByDymNameAddress(ctx, tc.input)
					if tc.want.(string) == "" {
						require.Error(t, err)
						require.Empty(t, outputAddr)
					} else {
						require.NoError(t, err)
						require.Equal(t, tc.want.(string), outputAddr)
					}
				case tcReverseResolveAddr:
					candidates, err := dk.ReverseResolveDymNameAddress(ctx, tc.input, useContextChainId)
					if tc.want.(string) == "" {
						require.NoError(t, err)
						require.Empty(t, candidates)
					} else {
						require.NoError(t, err)
						require.NotEmptyf(t, candidates, "want %s", tc.want.(string))
						require.Equal(t, tc.want.(string), candidates[0].String())
					}
				default:
					t.Fatalf("unknown test case type: %d", tc._type)
				}
			}
		})
	}
}
