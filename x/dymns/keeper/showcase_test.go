package keeper_test

import (
	"sort"
	"strings"
	"testing"
	"time"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

/**
This file contains showcases of how the DymNS config,
 resolve from Dym-Name-Address to address (forward-resolve),
 reverse-resolve from address to Dym-Name-Address,
 and how alias working in DymNS module.
*/

//goland:noinspection SpellCheckingInspection
func TestKeeper_NewRegistration(t *testing.T) {
	/**
	This show how Dym-Name record look like in store, after new registration
	And basic resolve Dym-Name-Address & reverse-resolve address to the Dym-Name-Address
	*/

	sc := setupShowcase(t)

	require.Equal(t, "dymension", sc.ctx.ChainID()) // our chain-id is "dymension"

	dymNameExpirationDate := sc.now.Add(365 * 24 * time.Hour)

	sc.
		newDymName(
			// name of the Dym-Name
			"my-name",
			// the owner of the Dym-Name
			"dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		).
		withExpiry(dymNameExpirationDate).
		save()

	t.Run("this show how the new Dym-Name record look like after new registration", func(t *testing.T) {
		sc.requireDymName("my-name").equals(
			dymnstypes.DymName{
				Name:       "my-name",
				Owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				Controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue", // default controller is the owner
				ExpireAt:   dymNameExpirationDate.Unix(),
				Configs:    nil, // default no config
				Contact:    "",  // default no contact
			},
		)
	})

	t.Run("this show how resolve Dym-Name-Address look like for brand-new Dym-Name", func(t *testing.T) {
		// resolve "my-name@dymension" to "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
		sc.
			requireResolveDymNameAddress("my-name@dymension").
			Equals("dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue")
	})

	t.Run("this show how reverse-resolve to a Dym-Name-Address look like for brand-new Dym-Name", func(t *testing.T) {
		// resolve "my-name@dymension" to "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
		// and reverse-resolve is to resolve from "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue" back to "my-name@dymension"

		sc.
			requireReverseResolve("dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue").forChainId("dymension").
			equals("my-name@dymension")

		// reverse lookup from 0x address
		ownerIn0xFormat := common.BytesToAddress(sdk.MustAccAddressFromBech32("dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue")).String()
		require.Equal(t, "0x4feA76427B8345861e80A3540a8a9D936FD39391", ownerIn0xFormat)
		sc.
			requireReverseResolve(ownerIn0xFormat).forChainId("dymension").
			equals("my-name@dymension")
		// reverse lookup from 0x address has some limitation, I'll provide more details at later parts
	})

	t.Run("this show how resolve address across all RollApps", func(t *testing.T) {
		dymName := sc.requireDymName("my-name").get()
		ownerAccAddr := sdk.MustAccAddressFromBech32(dymName.Owner)

		// register RollApps with their accounts bech32 prefix
		rollAppONE := sc.newRollApp("one_1-1").withBech32Prefix("one").save()
		rollAppTWO := sc.newRollApp("two_2-2").withBech32Prefix("two").save()
		rollAppWithoutBech32 := sc.newRollApp("nob_3-3").withoutBech32Prefix().save()

		// convert owner address to RollApp's bech32 prefix
		ownerWithONEPrefix := sdk.MustBech32ifyAddressBytes("one", ownerAccAddr)
		require.Equal(t, "one1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3668wjg", ownerWithONEPrefix)
		ownerWithTWOPrefix := sdk.MustBech32ifyAddressBytes("two", ownerAccAddr)
		require.Equal(t, "two1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3y4eefh", ownerWithTWOPrefix)

		sc.
			requireResolveDymNameAddress("my-name@" + rollAppONE.RollappId).
			Equals(ownerWithONEPrefix)

		sc.
			requireResolveDymNameAddress("my-name@" + rollAppTWO.RollappId).
			Equals(ownerWithTWOPrefix)

		// do not resolve for RollApp without bech32 prefix
		sc.
			requireResolveDymNameAddress("my-name@" + rollAppWithoutBech32.RollappId).
			NoResult()
	})

	t.Run("when Dym-Name is expired, resolution won't work in both ways", func(t *testing.T) {
		dymName := sc.requireDymName("my-name").get()
		dymName.ExpireAt = sc.now.Add(-1 * time.Hour).Unix()
		sc.requireDymName("my-name").update(*dymName)

		// resolve address don't work anymore
		sc.
			requireResolveDymNameAddress("my-name@dymension").
			NoResult()

		// reverse-resolve address don't work anymore
		sc.
			requireReverseResolve("dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue").forChainId("dymension").
			NoResult()
	})

	// Friendly notes for non-technical readers:
	// _________________________________________
	testAccount := sc.newTestAccount() // this creates a test account, which support procedure multiple address format
	// _________________________________________
	_ = testAccount.bech32() // this will output the address in bech32 format with "dym" prefix
	// look like this: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	// _________________________________________
	_ = testAccount.bech32C("rol") // this will output the address in bech32 format with "rol" prefix
	// look like this: "rol1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3n0r7hx"
	// _________________________________________
	_ = testAccount.hexStr() // this will output the address in 0x format
	// look like this: "0x4fea76427b8345861e80a3540a8a9d936fd39391"
	// _________________________________________
	_ = testAccount.checksumHex() // this will output the address in 0x format, with checksum
	// look like this: "0x4feA76427B8345861e80A3540a8a9D936FD39391" // similar hex but with mixed case for checksum
	// _________________________________________
	// 4 formats of the same address, but in different format
	// will be used many times later.
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_DefaultDymNameConfiguration(t *testing.T) {
	/**
	This adds more information about the default resolution
	and how things change after update the default resolution

	Default resolution is config where chain-id = host-chain, no sub-name.

	* limit to the default only *
	*/

	sc := setupShowcase(t)

	require.Equal(t, "dymension", sc.ctx.ChainID()) // our chain-id is "dymension"

	owner := "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	sc.
		newDymName(
			"my-name", // name of the Dym-Name
			owner,     // the owner of the Dym-Name
		).
		save()

	t.Run("default resolution", func(t *testing.T) {
		// by default, there is no configuration for configuration
		dymName := sc.requireDymName("my-name").get()
		require.Empty(t, dymName.Configs) // no config record

		// so when query Dym-Name-Address, it will resolve to the Owner
		sc.
			requireResolveDymNameAddress("my-name@dymension").
			Equals(owner)
		// so as reverse-resolve can find Dym-Name-Address from the Owner address
		sc.
			requireReverseResolve(owner).forChainId("dymension").
			equals("my-name@dymension")
		// reverse-resolve using 0x address
		sc.
			requireReverseResolve("0x4feA76427B8345861e80A3540a8a9D936FD39391"). // still owner, just in 0x format
			forChainId("dymension").
			equals("my-name@dymension")

		randomTestAccount := sc.newTestAccount()
		// reverse-resolve other addresses will return no result at this point
		sc.requireReverseResolve(randomTestAccount.bech32()).forChainId("dymension").NoResult()

		// why? Underthe hood, Dym-Name without any config equals with the one have a default config, which look like this
		dymName.Configs = []dymnstypes.DymNameConfig{
			{
				Type:    dymnstypes.DymNameConfigType_NAME,
				ChainId: "", // empty for host-chain
				Path:    "",
				Value:   owner,
			},
		}
		require.Len(t, dymName.Configs, 1) // ignore this line
	})

	t.Run("this is how things changed after we change the default resolution", func(t *testing.T) {
		dymName := sc.requireDymName("my-name").get()

		// create a new test account
		notTheOwner := sc.newTestAccount()

		dymName.Configs = []dymnstypes.DymNameConfig{
			{
				Type:    dymnstypes.DymNameConfigType_NAME,
				ChainId: "", // empty for host-chain
				Path:    "",
				Value:   notTheOwner.bech32(), // changed to another account
			},
		}

		sc.requireDymName("my-name").update(*dymName)

		// reverse-resolve no longer returns result for search of Dym-Name-Address using the Owner address
		sc.requireReverseResolve(owner).forChainId("dymension").NoResult()

		// Dym-Name-Address now resolves to the new account
		sc.
			requireResolveDymNameAddress("my-name@dymension").
			Equals(notTheOwner.bech32()) // look
		sc.
			requireResolveDymNameAddress("my-name@dymension").
			NotEquals(owner) // NO longer the owner

		// Reverse-resolve will return the Dym-Name-Address when lookup by the new account
		sc.
			requireReverseResolve(notTheOwner.bech32()).
			forChainId("dymension").
			equals("my-name@dymension")
		// and 0x
		sc.
			requireReverseResolve(notTheOwner.hexStr() /* 0x */).
			forChainId("dymension").
			equals("my-name@dymension")
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_DymNameConfiguration(t *testing.T) {
	/**
	This show how Dym-Name record look like in store, after update configuration resolution
	And resolve Dym-Name-Address & reverse-resolve address to the Dym-Name-Address in a more complex way
	*/

	sc := setupShowcase(t)

	require.Equal(t, "dymension", sc.ctx.ChainID()) // our chain-id is "dymension"

	owner := sc.newTestAccount() // account 1

	dymName := sc.
		newDymName(
			"my-name",
			owner.bech32(),
		).
		save()

	const rollAppId = "rollapp_1-1"
	const rollAppBech32Prefix = "rol"
	_ = sc.
		newRollApp(rollAppId).
		withBech32Prefix(rollAppBech32Prefix).
		save()

	// update the resolution configuration
	// vvvvvvv please PAY ATTENTION on this section vvvvvvv
	/// account 2
	anotherUser2 := sc.newTestAccount()
	updateDymName(dymName).resolveTo(anotherUser2.bech32()).onChain("dymension").withSubName("sub-name-host").add()
	sc.addLaterTest(func() {
		sc.requireResolveDymNameAddress("sub-name-host.my-name@dymension").Equals(anotherUser2.bech32())
		// on host-chain, configured address is case-insensitive
		sc.requireReverseResolve(
			anotherUser2.bech32(),
			swapCase(anotherUser2.bech32()),
		).forChainId("dymension").equals("sub-name-host.my-name@dymension")
		// able to reverse-lookup using 0x address on host-chain
		_0xAddr := anotherUser2.hexStr()
		sc.requireReverseResolve(
			_0xAddr,
			swapCase(_0xAddr),
		).forChainId("dymension").equals("sub-name-host.my-name@dymension")
	})
	/// account 3
	anotherUser3 := sc.newTestAccount()
	updateDymName(dymName).resolveTo(anotherUser3.bech32C("cosmos")).onChain("cosmoshub-4").add()
	sc.addLaterTest(func() {
		sc.requireResolveDymNameAddress("my-name@cosmoshub-4").Equals(anotherUser3.bech32C("cosmos"))
		sc.requireReverseResolve(anotherUser3.bech32C("cosmos")).forChainId("cosmoshub-4").equals("my-name@cosmoshub-4")
		// on non-host and non-RollApp, configured address is case-sensitive
		swappedCase := swapCase(anotherUser3.bech32C("cosmos"))
		sc.requireReverseResolve(swappedCase).forChainId("cosmoshub-4").NoResult()
		// not able to reverse-lookup using 0x address on non-host chain
		_0xAddr := anotherUser3.hexStr()
		sc.requireReverseResolve(_0xAddr).forChainId("cosmoshub-4").NoResult()
	})
	/// account 4
	anotherUser4 := sc.newTestAccount()
	updateDymName(dymName).resolveTo(anotherUser4.bech32C("cosmos")).onChain("cosmoshub-4").withSubName("sub-name").add()
	sc.addLaterTest(func() {
		sc.requireResolveDymNameAddress("sub-name.my-name@cosmoshub-4").Equals(anotherUser4.bech32C("cosmos"))
		sc.requireReverseResolve(anotherUser4.bech32C("cosmos")).forChainId("cosmoshub-4").equals("sub-name.my-name@cosmoshub-4")
	})
	/// account 5
	anotherUser5 := sc.newTestAccount()
	updateDymName(dymName).resolveTo(anotherUser5.bech32C(rollAppBech32Prefix)).onChain(rollAppId).add()
	sc.addLaterTest(func() {
		sc.requireResolveDymNameAddress("my-name@" + rollAppId).Equals(anotherUser5.bech32C(rollAppBech32Prefix))
		// on RollApp, configured address is case-insensitive
		sc.requireReverseResolve(
			anotherUser5.bech32C(rollAppBech32Prefix),
			swapCase(anotherUser5.bech32C(rollAppBech32Prefix)),
		).forChainId(rollAppId).equals("my-name@" + rollAppId)
		// able to reverse-lookup using 0x address on RollApp
		_0xAddr := anotherUser5.hexStr()
		sc.requireReverseResolve(
			_0xAddr,
			swapCase(_0xAddr),
		).forChainId(rollAppId).equals("my-name@" + rollAppId)
	})
	/// account 6
	anotherUser6 := sc.newTestAccount()
	updateDymName(dymName).resolveTo(anotherUser6.bech32C(rollAppBech32Prefix)).withSubName("sub-rol").onChain(rollAppId).add()
	sc.addLaterTest(func() {
		sc.requireResolveDymNameAddress("sub-rol.my-name@" + rollAppId).Equals(anotherUser6.bech32C(rollAppBech32Prefix))
		sc.requireReverseResolve(anotherUser6.bech32C(rollAppBech32Prefix)).forChainId(rollAppId).equals("sub-rol.my-name@" + rollAppId)
		_0xAddr := anotherUser6.hexStr()
		sc.requireReverseResolve(_0xAddr).forChainId(rollAppId).equals("sub-rol.my-name@" + rollAppId)
	})
	/// account 7 is a Bitcoin address
	anotherUser7 := "12higDjoCCNXSA95xZMWUdPvXNmkAduhWv"
	updateDymName(dymName).resolveTo(anotherUser7).onChain("bitcoin").add()
	sc.addLaterTest(func() {
		sc.requireResolveDymNameAddress("my-name@bitcoin").Equals(anotherUser7)
		sc.requireReverseResolve(anotherUser7).forChainId("bitcoin").equals("my-name@bitcoin")
		// on non-host and non-RollApp, configured address is case-sensitive
		swappedCase := swapCase(anotherUser7)
		sc.requireReverseResolve(swappedCase).forChainId("bitcoin").NoResult()
	})
	/// account 8 is an Ethereum checksum address (mixed case)
	anotherUser8WithChecksum := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"
	updateDymName(dymName).resolveTo(anotherUser8WithChecksum).onChain("ethereum").add()
	sc.addLaterTest(func() {
		/*
			Just above, we know that, on non-host and non-RollApp, configured address is case-sensitive
			But there is an exception: if the address is welformed a hex address, start with 0x,
			then treat the chain is case-insensitive address.
		*/

		lowercase := strings.ToLower(anotherUser8WithChecksum)
		uppercase := "0x" + strings.ToUpper(anotherUser8WithChecksum[2:])
		swappedCase := "0x" + swapCase(anotherUser8WithChecksum[2:])
		sc.requireResolveDymNameAddress("my-name@ethereum").Equals(lowercase)
		sc.requireReverseResolve(
			anotherUser8WithChecksum,
			lowercase,
			uppercase,
			swappedCase,
		).forChainId("ethereum").equals("my-name@ethereum")
	})
	// ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
	sc.requireDymName("my-name").update(*dymName)

	t.Run("this show how the new Dym-Name record look like after updated", func(t *testing.T) {
		sc.requireDymName("my-name").equals(
			dymnstypes.DymName{
				Name:       "my-name",
				Owner:      dymName.Owner,      // unchanged
				Controller: dymName.Controller, // unchanged
				ExpireAt:   dymName.ExpireAt,   // unchanged
				Contact:    "",
				Configs: []dymnstypes.DymNameConfig{
					// Note: there is no default configuration record here,
					// therefore my-name@dymension will resolve to the owner (account 1)

					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "", // chain-id will be set to empty if it is host-chain-id
						Path:    "sub-name-host",
						Value:   anotherUser2.bech32(), // account 2
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "cosmoshub-4",
						Path:    "",
						Value:   anotherUser3.bech32C("cosmos"), // account 3
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "cosmoshub-4",
						Path:    "sub-name",
						Value:   anotherUser4.bech32C("cosmos"), // account 4
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: rollAppId,
						Path:    "",
						Value:   anotherUser5.bech32C(rollAppBech32Prefix), // account 5
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: rollAppId,
						Path:    "sub-rol",
						Value:   anotherUser6.bech32C(rollAppBech32Prefix), // account 6
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "bitcoin",
						Path:    "",
						Value:   anotherUser7, // account 7
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "ethereum",
						Path:    "",
						Value:   strings.ToLower(anotherUser8WithChecksum), // account 8
						// the value is lowercased, even tho the original value is mixed-case
						// and the chain is neither host-chain nor RollApp
						// because it welformed as hex address, start with 0x
					},
				},
			},
		)
	})

	t.Run("resolve and reverse-resolve after updated", func(t *testing.T) {
		// now we run the pending tests
		sc.runPendingTests()
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_Alias(t *testing.T) {
	/**
	This show details about Alias: working with alias and how it affects the resolution
	*/

	sc := setupShowcase(t)

	require.Equal(t, "dymension", sc.ctx.ChainID()) // our chain-id is "dymension"

	// register some aliases
	sc.registerAlias("dym").forChainId("dymension")
	sc.registerAlias("btc").forChainId("bitcoin")
	sc.registerAlias("eth").forChainId("ethereum")

	_ = sc.newRollApp("one_1-1").withBech32Prefix("one").withAlias("one").save()

	owner := sc.newTestAccount()
	ethereumAddr := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"
	bitcoinAddr := "12higDjoCCNXSA95xZMWUdPvXNmkAduhWv"

	// register a Dym-Name
	dymName := sc.
		newDymName(
			"my-name",
			owner.bech32(),
		).
		save()
	// add some configuration
	updateDymName(dymName).resolveTo(bitcoinAddr).onChain("bitcoin").add()
	updateDymName(dymName).resolveTo(ethereumAddr).onChain("ethereum").add()
	sc.requireDymName("my-name").update(*dymName)

	t.Run("resolve address", func(t *testing.T) {
		// general resolution using chain-id
		sc.
			requireResolveDymNameAddress("my-name@dymension").
			Equals(owner.bech32())
		sc.
			requireResolveDymNameAddress("my-name@one_1-1").
			Equals(owner.bech32C("one"))
		sc.
			requireResolveDymNameAddress("my-name@bitcoin").
			Equals(bitcoinAddr)
		sc.
			requireResolveDymNameAddress("my-name@ethereum").
			Equals(strings.ToLower(ethereumAddr))

		// resolution using alias works
		sc.
			requireResolveDymNameAddress("my-name@dym").
			Equals(owner.bech32())
		sc.
			requireResolveDymNameAddress("my-name@one").
			Equals(owner.bech32C("one"))
		sc.
			requireResolveDymNameAddress("my-name@btc").
			Equals(bitcoinAddr)
		sc.
			requireResolveDymNameAddress("my-name@eth").
			Equals(strings.ToLower(ethereumAddr))
	})

	t.Run("reverse-resolve address", func(t *testing.T) {
		// if a chain has alias configured then reverse-resolve will use the alias instead of the chain-id
		sc.
			requireReverseResolve(owner.bech32()).
			forChainId("dymension").
			equals("my-name@dym") // alias "dym" is used instead of "dymension"
		sc.
			requireReverseResolve(owner.bech32C("one")).
			forChainId("one_1-1").
			equals("my-name@one") // alias "one" is used instead of "one_1-1"
		sc.
			requireReverseResolve(bitcoinAddr).
			forChainId("bitcoin").
			equals("my-name@btc") // alias "btc" is used instead of "bitcoin"
		sc.
			requireReverseResolve(ethereumAddr).
			forChainId("ethereum").
			equals("my-name@eth") // alias "eth" is used instead of "ethereum"
	})

	t.Run("when alias is defined in params, it has priority over RollApp", func(t *testing.T) {
		// this RollApp is uses the alias "cosmos" which supposed to belong to Cosmos Hub
		_ = sc.
			newRollApp("unexpected_2-2").
			withAlias("cosmos"). // NOOO...
			withBech32Prefix("unexpected").save()

		// add some resolution configuration for testing purpose
		updateDymName(dymName).resolveTo(owner.bech32C("cosmos")).onChain("cosmoshub-4").add()
		sc.requireDymName("my-name").update(*dymName)

		// now we see how it looks like
		sc.
			requireResolveDymNameAddress("my-name@cosmos").
			Equals(owner.bech32C("unexpected")) // that bad

		// to protect users, we take over the alias "cosmos" and give it to Cosmos Hub
		sc.registerAlias("cosmos").forChainId("cosmoshub-4") // <= this put into module params

		// now we see how it looks like, again
		sc.
			requireResolveDymNameAddress("my-name@cosmos").
			Equals(owner.bech32C("cosmos")) // that's it
	})

	t.Run("Reverse-resolve when alias is defined in params, it has priority over RollApp", func(t *testing.T) {
		// this RollApp is uses the alias "injective" which supposed to belong to Injective
		_ = sc.
			newRollApp("unintended_3-3").
			withAlias("injective"). // NOOO...
			withBech32Prefix("unintended").save()

		// add some resolution configuration for testing purpose
		updateDymName(dymName).resolveTo(owner.bech32C("unintended")).onChain("unintended_3-3").add()
		updateDymName(dymName).resolveTo(owner.bech32C("inj")).onChain("injective-1").add()
		sc.requireDymName("my-name").update(*dymName)
		sc.requireResolveDymNameAddress("my-name@injective").Equals(owner.bech32C("unintended")) // that bad

		sc.
			requireReverseResolve(owner.bech32C("unintended")).
			forChainId("unintended_3-3").
			equals("my-name@injective") // the alias is used in reverse-resolve
		// later, we want reverse-resolve won't use the alias for the RollApp
		// => "my-name@unintended_3-3"

		sc.
			requireReverseResolve(owner.bech32C("inj")).
			forChainId("injective-1").
			equals("my-name@injective-1") // no alias is used in reverse-resolve
		// later, we want reverse-resolve will use the alias for Injective
		// => "my-name@injective"

		// to protect users, we take over the alias "injective" and give it to Injective
		sc.registerAlias("injective").forChainId("injective-1") // <= this put into module params
		sc.requireResolveDymNameAddress("my-name@injective").Equals(owner.bech32C("inj"))

		// now test again
		sc.
			requireReverseResolve(owner.bech32C("unintended")).
			forChainId("unintended_3-3").
			equals("my-name@unintended_3-3") // RollApp no longer use the alias
		sc.
			requireReverseResolve(owner.bech32C("inj")).
			forChainId("injective-1").
			equals("my-name@injective") // alias is used in reverse-resolve
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_ResolveExtraFormat(t *testing.T) {
	/**
	This show additional details about Extra Format resolution for the Dym-Name-Address.
	Extra formats:
	- nim1...@dym
	- 0x.....@dym

	These formats are not the default resolution, but they are supported.
	It converts the address to another-chain-based bech32 format like:
	- nim1...@dym => dym1...
	- 0x.....@dym => dym1...
	- dym1...@nim => nim1...
	- 0x.....@nim => nim1...

	In this mode, no Dym-Name is required, and the resolution is based on the chain-id
	and the input address must be in the correct format and no Sub-Name.
	The result is expected to be the same as the input address, with prefix changed corresponding to the chain-id.
	*/

	sc := setupShowcase(t)

	require.Equal(t, "dymension", sc.ctx.ChainID()) // our chain-id is "dymension"
	sc.registerAlias("dym").forChainId("dymension")

	rollApp1 := sc.newRollApp("one_1-1").withBech32Prefix("one").withAlias("one").save()
	rollApp2 := sc.newRollApp("two_2-2").withBech32Prefix("two").withAlias("two").save()
	rollAppWithoutBech32 := sc.newRollApp("nob_3-3").withoutBech32Prefix().save()

	testAccount := sc.newTestAccount()

	t.Run("resolve hex to chain-based bech32", func(t *testing.T) {
		// 0x...@dymension
		sc.requireResolveDymNameAddress(testAccount.hexStr() + "@dymension").
			Equals(testAccount.bech32())

		// 0x...@one_1-1
		sc.requireResolveDymNameAddress(testAccount.hexStr() + "@" + rollApp1.RollappId).
			Equals(testAccount.bech32C("one"))

		// 0x...@two_2-2
		sc.requireResolveDymNameAddress(testAccount.hexStr() + "@" + rollApp2.RollappId).
			Equals(testAccount.bech32C("two"))

		// 0x...@nob_3-3
		sc.requireResolveDymNameAddress(testAccount.hexStr() + "@" + rollAppWithoutBech32.RollappId).
			NoResult() // won't work for RollApp without bech32 prefix because we don't know bech32 prefix to cast

		// also works with alias
		sc.requireResolveDymNameAddress(testAccount.hexStr() + "@dym").
			Equals(testAccount.bech32())
		sc.requireResolveDymNameAddress(testAccount.hexStr() + "@one").
			Equals(testAccount.bech32C("one"))
	})

	t.Run("resolve bech32 to chain-based bech32", func(t *testing.T) {
		// kava1...@dymension
		sc.requireResolveDymNameAddress(testAccount.bech32C("kava") + "@dymension").
			Equals(testAccount.bech32())

		// dym1...@one_1-1
		sc.requireResolveDymNameAddress(testAccount.bech32() + "@" + rollApp1.RollappId).
			Equals(testAccount.bech32C("one"))

		// whatever1...@two_2-2
		sc.requireResolveDymNameAddress(testAccount.bech32C("whatever") + "@" + rollApp2.RollappId).
			Equals(testAccount.bech32C("two"))

		// also works with alias
		sc.requireResolveDymNameAddress(testAccount.bech32C("kava") + "@dym").
			Equals(testAccount.bech32())
		sc.requireResolveDymNameAddress(testAccount.bech32() + "@two").
			Equals(testAccount.bech32C("two"))
	})

	sc.requireResolveDymNameAddress("otherNotWelformedAddress@dymension").
		WillError()

	t.Run("when alias is defined in params, it has priority over RollApp", func(t *testing.T) {
		_ = sc.
			newRollApp("unintended_3-3").
			withAlias("injective").
			withBech32Prefix("unintended").
			save()

		sc.requireResolveDymNameAddress(testAccount.hexStr() + "@injective").
			Equals(testAccount.bech32C("unintended"))

		sc.registerAlias("injective").forChainId("injective-1") // <= this put into module params

		sc.requireResolveDymNameAddress(testAccount.hexStr() + "@injective").
			NoResult()
		// because Injective is not the RollApp anymore.
		// It neither the host-chain nor RollApp, so it should not resolve
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_ReverseResolve(t *testing.T) {
	/**
	This show advanced details about Reverse-Resolve:
	working with reverse-resolve and some vector affects the resolution
	*/

	sc := setupShowcase(t)

	require.Equal(t, "dymension", sc.ctx.ChainID()) // our chain-id is "dymension"

	owner := sc.newTestAccount()

	// register a Dym-Name
	_ = sc.
		newDymName(
			"my-name",
			owner.bech32(),
		).
		save()

	// register a RollApp
	_ = sc.newRollApp("one_1-1").withBech32Prefix("one").save()

	t.Run("normal reverse-resolve", func(t *testing.T) {
		sc.
			requireReverseResolve(owner.bech32()).
			forChainId("dymension").
			equals("my-name@dymension")
		sc.
			requireReverseResolve(owner.bech32C("any")).
			forChainId("dymension").
			equals("my-name@dymension")

		// reverse-resolve using 0x address
		sc.
			requireReverseResolve(owner.hexStr()).
			forChainId("dymension").
			equals("my-name@dymension")

		// reverse-resolve using bech32 for RollApp Dym-Name-Address
		sc.
			requireReverseResolve(owner.bech32()).
			forChainId("one_1-1").
			equals("my-name@one_1-1")

		// reverse-resolve using any bech32 for RollApp Dym-Name-Address
		sc.
			requireReverseResolve(owner.bech32C("any")).
			forChainId("one_1-1").
			equals("my-name@one_1-1")
	})

	t.Run("normal reverse-resolve for NEITHER host-chain nor RollApp", func(t *testing.T) {
		sc.
			requireReverseResolve(owner.bech32()).
			forChainId("injective-1").
			NoResult() // won't work
		sc.
			requireReverseResolve(owner.hexStr()).
			forChainId("injective-1").
			NoResult() // won't work
		sc.
			requireReverseResolve(owner.bech32()).
			forChainId("cosmoshub-4").
			NoResult() // won't work
		sc.
			requireReverseResolve(owner.hexStr()).
			forChainId("cosmoshub-4").
			NoResult() // won't work
		// the main reason for those do not work is we don't know if the chain is coin-type-60 or not,
		// so we can not blindly reverse-resolve the address.

		// Cosmos Hub is a good example, it is coin-type-118,
		// so we CAN NOT convert from both bech32 & 0x format of Dymension to Cosmos Hub address.
		// Because if user send funds to those address, it will be lost.

		// And that's why we Do Not support reverse-resolve for NEITHER host-chain nor RollApp.

		// But there is a way to make it works, by define the chain, into the module params.
		sc.registerChainId("injective-1").asCoinType60Chain() // <= this put into module params

		// now let's see how things are changed
		sc.
			requireReverseResolve(owner.bech32()).
			forChainId("injective-1").
			equals("my-name@injective-1") // now it works
		sc.
			requireReverseResolve(owner.hexStr()).
			forChainId("injective-1").
			equals("my-name@injective-1") // now it works either
		sc.
			requireReverseResolve(owner.bech32()).
			forChainId("cosmoshub-4").
			NoResult() // won't work
		sc.
			requireReverseResolve(owner.hexStr()).
			forChainId("cosmoshub-4").
			NoResult() // won't work

		// but Forward-resolve will not return result because we don't know the bech32 prefix
		sc.requireResolveDymNameAddress("my-name@injective-1").NoResult()
	})

	t.Run("when alias is registered, reverse resolve will use the alias instead of chain-id", func(t *testing.T) {
		sc.registerAlias("inj").forChainId("injective-1")
		sc.newRollApp("two_2-2").withAlias("two").save()

		sc.
			requireReverseResolve(owner.bech32()).
			forChainId("injective-1").
			equals("my-name@inj")
		sc.
			requireReverseResolve(owner.hexStr()).
			forChainId("injective-1").
			equals("my-name@inj")
		sc.
			requireReverseResolve(owner.hexStr()).
			forChainId("two_2-2").
			equals("my-name@two")
	})
}

/* -------------------------------------------------------------------------- */
/*                     setup area, no need to read                            */
/* -------------------------------------------------------------------------- */

func setupShowcase(t *testing.T) *showcaseSetup {
	const chainId = "dymension"
	now := time.Now().UTC()

	dk, _, rk, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now).WithChainID(chainId)

	scs := &showcaseSetup{
		now: now,

		ctx: ctx,
		dk:  dk,
		rk:  rk,

		t: t,
	}

	scs.dymNameOwner = scs.newTestAccount()

	return scs
}

type showcaseSetup struct {
	now time.Time

	ctx sdk.Context
	dk  dymnskeeper.Keeper
	rk  rollappkeeper.Keeper

	recentTestAccountNo uint64
	dymNameOwner        ta

	t          *testing.T
	laterTests []func()
}

func (m *showcaseSetup) newTestAccount() ta {
	m.recentTestAccountNo++
	return testAddr(m.recentTestAccountNo)
}

func (m *showcaseSetup) newDymName(name string, owner string) *configureDymName {
	return &configureDymName{
		scs:        m,
		name:       name,
		owner:      owner,
		controller: owner,
		expiry:     m.now.Add(365 * 24 * time.Hour),
		configs:    nil,
	}
}

func (m *showcaseSetup) newRollApp(rollAppId string) *configureRollApp {
	return &configureRollApp{
		scs:       m,
		rollAppId: rollAppId,
	}
}

func (m *showcaseSetup) requireDymName(name string) reqDymName {
	return reqDymName{
		scs:  m,
		name: name,
	}
}

func (m *showcaseSetup) requireResolveDymNameAddress(dymNameAddress string) reqResolveDymNameAddress {
	return reqResolveDymNameAddress{
		scs:            m,
		dymNameAddress: dymNameAddress,
	}
}

func (m *showcaseSetup) requireReverseResolve(addresses ...string) *reqReverseResolveDymNameAddress {
	return &reqReverseResolveDymNameAddress{
		scs:            m,
		workingChainId: m.ctx.ChainID(),
		addresses:      addresses,
	}
}

func (m *showcaseSetup) addLaterTest(laterTest func()) {
	m.laterTests = append(m.laterTests, laterTest)
}

func (m *showcaseSetup) runPendingTests() {
	defer func() {
		m.laterTests = nil // clear
	}()
	for _, laterTest := range m.laterTests {
		laterTest()
	}
}

func (m *showcaseSetup) registerAlias(alias string) regAlias {
	return regAlias{
		scs:   m,
		alias: alias,
	}
}

func (m *showcaseSetup) registerChainId(chainId string) regChainId {
	return regChainId{
		scs:     m,
		chainId: chainId,
	}
}

type reqDymName struct {
	scs  *showcaseSetup
	name string
}

func (m reqDymName) equals(otherDymName dymnstypes.DymName) {
	dymName := m.scs.dk.GetDymName(m.scs.ctx, m.name)
	require.NotNil(m.scs.t, dymName)
	require.Equal(m.scs.t, otherDymName, *dymName)
}

func (m reqDymName) get() *dymnstypes.DymName {
	return m.scs.dk.GetDymName(m.scs.ctx, m.name)
}

func (m reqDymName) MustHasConfig(filter func(cfg dymnstypes.DymNameConfig) bool) {
	dymName := m.get()
	var anyMatch bool
	for _, cfg := range dymName.Configs {
		if filter(cfg) {
			anyMatch = true
			break
		}
	}
	require.True(m.scs.t, anyMatch)
}

func (m reqDymName) NotHaveConfig(filter func(cfg dymnstypes.DymNameConfig) bool) {
	dymName := m.get()
	for _, cfg := range dymName.Configs {
		require.False(m.scs.t, filter(cfg))
	}
}

func (m reqDymName) update(dymName dymnstypes.DymName) {
	require.Equal(m.scs.t, m.name, dymName.Name, "passed wrong Dym-Name")

	for i, config := range dymName.Configs {
		if config.ChainId == m.scs.ctx.ChainID() {
			config.ChainId = ""
			dymName.Configs[i] = config
		}
	}

	setDymNameWithFunctionsAfter(m.scs.ctx, dymName, m.scs.t, m.scs.dk)
}

type configureDymName struct {
	scs        *showcaseSetup
	name       string
	owner      string
	controller string
	expiry     time.Time
	configs    []dymnstypes.DymNameConfig
}

func (m *configureDymName) withExpiry(expiry time.Time) *configureDymName {
	m.expiry = expiry
	return m
}

func (m configureDymName) build() dymnstypes.DymName {
	return dymnstypes.DymName{
		Name:       m.name,
		Owner:      m.owner,
		Controller: m.controller,
		ExpireAt:   m.expiry.Unix(),
		Configs:    m.configs,
		Contact:    "",
	}
}

func (m configureDymName) save() *dymnstypes.DymName {
	dymName := m.build()
	setDymNameWithFunctionsAfter(m.scs.ctx, dymName, m.scs.t, m.scs.dk)

	record := m.scs.dk.GetDymName(m.scs.ctx, dymName.Name)
	require.NotNil(m.scs.t, record)
	return record
}

type configureRollApp struct {
	scs          *showcaseSetup
	rollAppId    string
	alias        string
	bech32Prefix string
}

func (m *configureRollApp) withAlias(alias string) *configureRollApp {
	m.alias = alias
	return m
}

func (m *configureRollApp) withBech32Prefix(bech32Prefix string) *configureRollApp {
	m.bech32Prefix = bech32Prefix
	return m
}

func (m *configureRollApp) withoutBech32Prefix() *configureRollApp {
	m.bech32Prefix = ""
	return m
}

func (m configureRollApp) save() *rollapptypes.Rollapp {
	registerRollApp(m.scs.t, m.scs.ctx, m.scs.rk, m.scs.dk, m.rollAppId, m.bech32Prefix, m.alias)

	rollApp, found := m.scs.rk.GetRollapp(m.scs.ctx, m.rollAppId)
	require.True(m.scs.t, found)
	return &rollApp
}

type reqResolveDymNameAddress struct {
	scs            *showcaseSetup
	dymNameAddress string
}

func (m reqResolveDymNameAddress) Equals(want string) {
	outputAddress, err := m.scs.dk.ResolveByDymNameAddress(m.scs.ctx, m.dymNameAddress)
	require.NoError(m.scs.t, err)
	require.Equal(m.scs.t, want, outputAddress)
}

func (m reqResolveDymNameAddress) NotEquals(want string) {
	outputAddress, err := m.scs.dk.ResolveByDymNameAddress(m.scs.ctx, m.dymNameAddress)
	require.NoError(m.scs.t, err)
	require.NotEqual(m.scs.t, want, outputAddress)
}

func (m reqResolveDymNameAddress) NoResult() {
	outputAddress, err := m.scs.dk.ResolveByDymNameAddress(m.scs.ctx, m.dymNameAddress)
	if err != nil {
		require.Contains(m.scs.t, err.Error(), "not found")
	} else {
		require.Empty(m.scs.t, outputAddress)
	}
}

func (m reqResolveDymNameAddress) WillError() {
	_, err := m.scs.dk.ResolveByDymNameAddress(m.scs.ctx, m.dymNameAddress)
	require.Error(m.scs.t, err)
}

type reqReverseResolveDymNameAddress struct {
	scs            *showcaseSetup
	workingChainId string
	addresses      []string
}

func (m *reqReverseResolveDymNameAddress) forChainId(workingChainId string) *reqReverseResolveDymNameAddress {
	m.workingChainId = workingChainId
	return m
}

func (m reqReverseResolveDymNameAddress) equals(wantMany ...string) {
	for _, address := range m.addresses {
		m.scs.t.Run("reverse-resolve for "+address, func(t *testing.T) {
			list, err := m.scs.dk.ReverseResolveDymNameAddress(m.scs.ctx, address, m.workingChainId)
			require.NoError(m.scs.t, err)

			var dymNameAddresses []string
			for _, dna := range list {
				dymNameAddresses = append(dymNameAddresses, dna.String())
			}

			sort.Strings(dymNameAddresses)
			sort.Strings(wantMany)
			require.Equal(m.scs.t, wantMany, dymNameAddresses)
		})
	}
}

func (m reqReverseResolveDymNameAddress) NoResult() {
	for _, address := range m.addresses {
		m.scs.t.Run("reverse-resolve for "+address, func(t *testing.T) {
			list, err := m.scs.dk.ReverseResolveDymNameAddress(m.scs.ctx, address, m.workingChainId)
			require.NoError(m.scs.t, err)
			require.Empty(m.scs.t, list)
		})
	}
}

type udtDymName struct {
	dymName *dymnstypes.DymName
}

func updateDymName(dymName *dymnstypes.DymName) *udtDymName {
	return &udtDymName{dymName: dymName}
}

func (m *udtDymName) resolveTo(value string) *udtDymNameConfigResolveTo {
	return &udtDymNameConfigResolveTo{
		udtDymName: m,
		value:      value,
	}
}

type udtDymNameConfigResolveTo struct {
	udtDymName *udtDymName
	chainId    string
	subName    string
	value      string
}

func (m *udtDymNameConfigResolveTo) onChain(chainId string) *udtDymNameConfigResolveTo {
	m.chainId = chainId
	return m
}

func (m *udtDymNameConfigResolveTo) withSubName(subName string) *udtDymNameConfigResolveTo {
	m.subName = subName
	return m
}

func (m *udtDymNameConfigResolveTo) add() {
	value := m.value

	if dymnsutils.IsValidHexAddress(value) {
		value = strings.ToLower(value)
	}

	m.udtDymName.dymName.Configs = append(m.udtDymName.dymName.Configs, dymnstypes.DymNameConfig{
		Type:    dymnstypes.DymNameConfigType_NAME,
		ChainId: m.chainId,
		Path:    m.subName,
		Value:   value,
	})
}

type regAlias struct {
	scs   *showcaseSetup
	alias string
}

func (m regAlias) forChainId(chainId string) {
	require.NotEmpty(m.scs.t, m.alias)
	require.NotEmpty(m.scs.t, chainId)

	moduleParams := m.scs.dk.GetParams(m.scs.ctx)
	moduleParams.Chains.AliasesOfChainIds = append(moduleParams.Chains.AliasesOfChainIds, dymnstypes.AliasesOfChainId{
		ChainId: chainId,
		Aliases: []string{m.alias},
	})

	err := m.scs.dk.SetParams(m.scs.ctx, moduleParams)
	require.NoError(m.scs.t, err)
}

type regChainId struct {
	scs     *showcaseSetup
	chainId string
}

func (m regChainId) asCoinType60Chain() {
	require.NotEmpty(m.scs.t, m.chainId)

	moduleParams := m.scs.dk.GetParams(m.scs.ctx)
	moduleParams.Chains.CoinType60ChainIds = append(moduleParams.Chains.CoinType60ChainIds, m.chainId)

	err := m.scs.dk.SetParams(m.scs.ctx, moduleParams)
	require.NoError(m.scs.t, err)
}
