package keeper_test

import (
	"sort"
	"strings"
	"testing"
	"time"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetAddReverseMappingOwnerToOwnedDymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	t.Run("should not allow invalid owner address", func(t *testing.T) {
		require.Error(t, dk.AddReverseMappingOwnerToOwnedDymName(ctx, "0x", "a"))

		_, err := dk.GetDymNamesOwnedBy(ctx, "0x")
		require.Error(t, err)
	})

	owner1a := testAddr(1).bech32()
	owner2a := testAddr(2).bech32()
	notOwnerA := testAddr(3).bech32()

	dymName11 := dymnstypes.DymName{
		Name:       "n11",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName11))

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName21))

	dymName22 := dymnstypes.DymName{
		Name:       "n22",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName22))

	t.Run("can add", func(t *testing.T) {
		var err error

		err = dk.AddReverseMappingOwnerToOwnedDymName(ctx, owner1a, dymName11.Name)
		require.NoError(t, err)

		err = dk.AddReverseMappingOwnerToOwnedDymName(ctx, owner2a, dymName21.Name)
		require.NoError(t, err)

		err = dk.AddReverseMappingOwnerToOwnedDymName(ctx, owner2a, dymName22.Name)
		require.NoError(t, err)
	})

	t.Run("can add non-existing dym-name", func(t *testing.T) {
		err := dk.AddReverseMappingOwnerToOwnedDymName(ctx, owner2a, "not-exists")
		require.NoError(t, err)
	})

	t.Run("no error when adding duplicated name", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			err := dk.AddReverseMappingOwnerToOwnedDymName(ctx, owner2a, dymName21.Name)
			require.NoError(t, err)
		}
	})

	tests := []struct {
		name   string
		owner  string
		preRun func()
		want   []string
	}{
		{
			name:  "get - returns correctly",
			owner: owner1a,
			want:  []string{dymName11.Name},
		},
		{
			name:  "get - returns correctly",
			owner: owner2a,
			want:  []string{dymName21.Name, dymName22.Name},
		},
		{
			name:  "get - returns empty if account not owned any Dym-Name",
			owner: notOwnerA,
			want:  nil,
		},
		{
			name:  "get - result not include not-owned Dym-Name",
			owner: owner2a,
			preRun: func() {
				require.NoError(
					t,
					dk.AddReverseMappingOwnerToOwnedDymName(ctx, owner2a, dymName11.Name),
					"no error if dym-name owned by another owner",
				)
				require.NoError(
					t,
					dk.AddReverseMappingOwnerToOwnedDymName(ctx, owner2a, "non-existence"),
					"no error if dym-name owned by another owner",
				)
			},
			want: []string{dymName21.Name, dymName22.Name},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preRun != nil {
				tt.preRun()
			}

			ownedDymNames, err := dk.GetDymNamesOwnedBy(ctx, tt.owner)
			require.NoError(t, err)

			requireDymNameList(ownedDymNames, tt.want, t)
		})
	}
}

func TestKeeper_RemoveReverseMappingOwnerToOwnedDymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Error(
		t,
		dk.RemoveReverseMappingOwnerToOwnedDymName(ctx, "0x", "a"),
		"should not allow invalid owner address",
	)

	owner1a := testAddr(1).bech32()
	owner2a := testAddr(2).bech32()
	notOwnerA := testAddr(3).bech32()

	dymName11 := dymnstypes.DymName{
		Name:       "n11",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName11, t, dk)

	dymName12 := dymnstypes.DymName{
		Name:       "n12",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName12, t, dk)

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName21, t, dk)

	require.NoError(
		t,
		dk.RemoveReverseMappingOwnerToOwnedDymName(ctx, notOwnerA, dymName11.Name),
		"no error if owner non-exists",
	)

	require.NoError(
		t,
		dk.RemoveReverseMappingOwnerToOwnedDymName(ctx, owner1a, dymName21.Name),
		"no error if not owned dym-name",
	)
	ownedBy, err := dk.GetDymNamesOwnedBy(ctx, owner1a)
	require.NoError(t, err)
	requireDymNameList(ownedBy, []string{dymName11.Name, dymName12.Name}, t, "existing data must be kept")

	require.NoError(
		t,
		dk.RemoveReverseMappingOwnerToOwnedDymName(ctx, owner1a, "not-exists"),
		"no error if not-exists dym-name",
	)
	ownedBy, err = dk.GetDymNamesOwnedBy(ctx, owner1a)
	require.NoError(t, err)
	requireDymNameList(ownedBy, []string{dymName11.Name, dymName12.Name}, t, "existing data must be kept")

	require.NoError(
		t,
		dk.RemoveReverseMappingOwnerToOwnedDymName(ctx, owner1a, dymName11.Name),
	)
	ownedBy, err = dk.GetDymNamesOwnedBy(ctx, owner1a)
	require.NoError(t, err)
	requireDymNameList(ownedBy, []string{dymName12.Name}, t)

	require.NoError(
		t,
		dk.RemoveReverseMappingOwnerToOwnedDymName(ctx, owner1a, dymName12.Name),
	)
	ownedBy, err = dk.GetDymNamesOwnedBy(ctx, owner1a)
	require.NoError(t, err)
	require.Len(t, ownedBy, 0)
}

func TestKeeper_GetAddReverseMappingConfiguredAddressToDymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	t.Run("fail - should reject blank address", func(t *testing.T) {
		require.Error(t, dk.AddReverseMappingConfiguredAddressToDymName(ctx, " ", "a"))

		_, err := dk.GetDymNamesContainsConfiguredAddress(ctx, " ")
		require.Error(t, err)
	})

	owner1a := testAddr(1).bech32()
	owner2a := testAddr(2).bech32()
	anotherA := testAddr(3).bech32()
	icaA := testICAddr(4).bech32()
	someoneA := testAddr(5).bech32()

	dymName11 := dymnstypes.DymName{
		Name:       "n11",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: anotherA,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName11))
	err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, anotherA, dymName11.Name)
	require.NoError(t, err)

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName21))
	err = dk.AddReverseMappingConfiguredAddressToDymName(ctx, owner2a, dymName21.Name)
	require.NoError(t, err)

	dymName22 := dymnstypes.DymName{
		Name:       "n22",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: anotherA,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName22))
	err = dk.AddReverseMappingConfiguredAddressToDymName(ctx, anotherA, dymName22.Name)
	require.NoError(t, err)

	require.NoError(
		t,
		dk.AddReverseMappingConfiguredAddressToDymName(ctx, anotherA, "not-exists"),
		"no check non-existing dym-name",
	)

	t.Run("no error if duplicated name", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t,
				dk.AddReverseMappingConfiguredAddressToDymName(ctx, owner2a, dymName21.Name),
			)
		}
	})

	linked1, err1 := dk.GetDymNamesContainsConfiguredAddress(ctx, anotherA)
	require.NoError(t, err1)
	require.Len(t, linked1, 2)
	requireEqualsStrings(t,
		[]string{dymName11.Name, dymName22.Name},
		[]string{linked1[0].Name, linked1[1].Name},
	)

	linked2, err2 := dk.GetDymNamesContainsConfiguredAddress(ctx, owner2a)
	require.NoError(t, err2)
	require.NotEqual(t, 2, len(linked2), "should not include non-existing dym-name")
	require.Len(t, linked2, 1)
	requireEqualsStrings(t,
		[]string{dymName21.Name},
		[]string{linked2[0].Name},
	)

	linkedByNotExists, err3 := dk.GetDymNamesContainsConfiguredAddress(ctx, someoneA)
	require.NoError(t, err3)
	require.Len(t, linkedByNotExists, 0)

	t.Run("allow Interchain Account (32 bytes)", func(t *testing.T) {
		require.NoError(
			t,
			dk.AddReverseMappingConfiguredAddressToDymName(ctx, icaA, dymName11.Name),
		)

		linked3, err := dk.GetDymNamesContainsConfiguredAddress(ctx, icaA)
		require.NoError(t, err)
		require.Len(t, linked3, 1)
		require.Equal(t, dymName11.Name, linked3[0].Name)
	})

	t.Run("insert and get must be case-sensitive", func(t *testing.T) {
		addr1 := strings.ToLower(owner1a)
		addr2 := strings.ToUpper(owner1a)
		addr3 := strings.ToLower(owner1a[:10]) + strings.ToUpper(owner1a[10:20]) + strings.ToLower(owner1a[20:])

		dymName := dymnstypes.DymName{
			Name:       "my-name",
			Owner:      owner1a,
			Controller: owner1a,
			ExpireAt:   ctx.BlockTime().Add(time.Hour).Unix(),
		}
		require.NoError(t, dk.SetDymName(ctx, dymName))

		err = dk.AddReverseMappingConfiguredAddressToDymName(ctx, addr1, dymName.Name)
		require.NoError(t, err)
		err = dk.AddReverseMappingConfiguredAddressToDymName(ctx, addr2, dymName.Name)
		require.NoError(t, err)
		err = dk.AddReverseMappingConfiguredAddressToDymName(ctx, addr3, dymName.Name)
		require.NoError(t, err)

		linked1, err1 := dk.GetDymNamesContainsConfiguredAddress(ctx, addr1)
		require.NoError(t, err1)
		requireDymNameList(linked1, []string{dymName.Name}, t)

		linked2, err2 := dk.GetDymNamesContainsConfiguredAddress(ctx, addr2)
		require.NoError(t, err2)
		requireDymNameList(linked2, []string{dymName.Name}, t)

		linked3, err3 := dk.GetDymNamesContainsConfiguredAddress(ctx, addr3)
		require.NoError(t, err3)
		requireDymNameList(linked3, []string{dymName.Name}, t)
	})
}

func TestKeeper_RemoveReverseMappingConfiguredAddressToDymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Error(
		t,
		dk.RemoveReverseMappingConfiguredAddressToDymName(ctx, " ", "a"),
		"should not allow blank address",
	)

	ownerA := testAddr(1).bech32()
	anotherA := testAddr(2).bech32()
	icaA := testICAddr(3).bech32()
	someoneA := testAddr(4).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: anotherA,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName1))
	err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, anotherA, dymName1.Name)
	require.NoError(t, err)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: anotherA,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))
	err = dk.AddReverseMappingConfiguredAddressToDymName(ctx, anotherA, dymName2.Name)
	require.NoError(t, err)

	require.NoError(
		t,
		dk.RemoveReverseMappingConfiguredAddressToDymName(ctx, someoneA, dymName2.Name),
		"no error if record not exists",
	)

	linked, err := dk.GetDymNamesContainsConfiguredAddress(ctx, anotherA)
	require.NoError(t, err)
	require.Len(t, linked, 2, "existing data must be kept")

	t.Run("no error if element is not in the list", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingConfiguredAddressToDymName(ctx, anotherA, "not-exists"),
		)
		linked, err = dk.GetDymNamesContainsConfiguredAddress(ctx, anotherA)
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("remove correctly", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingConfiguredAddressToDymName(ctx, anotherA, dymName1.Name),
		)

		linked, err = dk.GetDymNamesContainsConfiguredAddress(ctx, anotherA)
		require.NoError(t, err)
		require.Len(t, linked, 1)
		require.Equal(t, dymName2.Name, linked[0].Name)

		require.NoError(
			t,
			dk.RemoveReverseMappingConfiguredAddressToDymName(ctx, anotherA, dymName2.Name),
		)

		linked, err = dk.GetDymNamesContainsConfiguredAddress(ctx, anotherA)
		require.NoError(t, err)
		require.Empty(t, linked)
	})

	t.Run("remove correctly with Interchain Account (32 bytes)", func(t *testing.T) {
		require.NoError(
			t,
			dk.AddReverseMappingConfiguredAddressToDymName(ctx, icaA, dymName1.Name),
		)

		linked3, err := dk.GetDymNamesContainsConfiguredAddress(ctx, icaA)
		require.NoError(t, err)
		require.Len(t, linked3, 1)

		require.NoError(
			t,
			dk.RemoveReverseMappingConfiguredAddressToDymName(ctx, icaA, dymName1.Name),
		)

		linked, err = dk.GetDymNamesContainsConfiguredAddress(ctx, icaA)
		require.NoError(t, err)
		require.Empty(t, linked)
	})

	t.Run("remove must be case-sensitive", func(t *testing.T) {
		addr1 := strings.ToLower(ownerA)
		addr2 := strings.ToUpper(ownerA)
		addr3 := strings.ToLower(ownerA[:10]) + strings.ToUpper(ownerA[10:20]) + strings.ToLower(ownerA[20:])

		dymName := dymnstypes.DymName{
			Name:       "my-name",
			Owner:      ownerA,
			Controller: ownerA,
			ExpireAt:   ctx.BlockTime().Add(time.Hour).Unix(),
		}
		require.NoError(t, dk.SetDymName(ctx, dymName))

		err = dk.AddReverseMappingConfiguredAddressToDymName(ctx, addr1, dymName.Name)
		require.NoError(t, err)
		err = dk.AddReverseMappingConfiguredAddressToDymName(ctx, addr2, dymName.Name)
		require.NoError(t, err)
		err = dk.AddReverseMappingConfiguredAddressToDymName(ctx, addr3, dymName.Name)
		require.NoError(t, err)

		linked1, err1 := dk.GetDymNamesContainsConfiguredAddress(ctx, addr1)
		require.NoError(t, err1)
		requireDymNameList(linked1, []string{dymName.Name}, t)

		linked2, err2 := dk.GetDymNamesContainsConfiguredAddress(ctx, addr2)
		require.NoError(t, err2)
		requireDymNameList(linked2, []string{dymName.Name}, t)

		linked3, err3 := dk.GetDymNamesContainsConfiguredAddress(ctx, addr3)
		require.NoError(t, err3)
		requireDymNameList(linked3, []string{dymName.Name}, t)

		err = dk.RemoveReverseMappingConfiguredAddressToDymName(ctx, addr3, dymName.Name)
		require.NoError(t, err)

		linked3, err3 = dk.GetDymNamesContainsConfiguredAddress(ctx, addr3)
		require.NoError(t, err3)
		require.Empty(t, linked3)
	})
}

func TestKeeper_GetAddReverseMappingFallbackAddressToDymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	for size := 0; size <= 128; size++ {
		if size == 20 || size == 32 {
			continue // two valid size
		}

		addr := make([]byte, size)

		require.Errorf(
			t,
			dk.AddReverseMappingFallbackAddressToDymName(ctx, addr, "a"),
			"should not allow %d bytes address", size,
		)

		_, err := dk.GetDymNamesContainsFallbackAddress(ctx, addr)
		require.Errorf(
			t,
			err,
			"should not allow %d bytes address", size,
		)
	}

	owner1Acc := testAddr(1)
	owner2Acc := testAddr(2)
	anotherAcc := testAddr(3)
	icaAcc := testICAddr(4)

	dymName11 := dymnstypes.DymName{
		Name:       "n11",
		Owner:      owner1Acc.bech32(),
		Controller: owner1Acc.bech32(),
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: anotherAcc.bech32(),
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName11))
	err := dk.AddReverseMappingFallbackAddressToDymName(ctx, anotherAcc.bytes(), dymName11.Name)
	require.NoError(t, err)

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2Acc.bech32(),
		Controller: owner2Acc.bech32(),
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName21))
	err = dk.AddReverseMappingFallbackAddressToDymName(
		ctx,
		owner2Acc.bytes(),
		dymName21.Name,
	)
	require.NoError(t, err)

	dymName22 := dymnstypes.DymName{
		Name:       "n22",
		Owner:      owner2Acc.bech32(),
		Controller: owner2Acc.bech32(),
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: anotherAcc.bech32(),
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName22))
	err = dk.AddReverseMappingFallbackAddressToDymName(ctx, anotherAcc.bytes(), dymName22.Name)
	require.NoError(t, err)

	require.NoError(
		t,
		dk.AddReverseMappingFallbackAddressToDymName(ctx, anotherAcc.bytes(), "not-exists"),
		"no check non-existing dym-name",
	)

	t.Run("no error if duplicated name", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t,
				dk.AddReverseMappingFallbackAddressToDymName(ctx, owner2Acc.bytes(), dymName21.Name),
			)
		}
	})

	linked1, err1 := dk.GetDymNamesContainsFallbackAddress(ctx, anotherAcc.bytes())
	require.NoError(t, err1)
	require.Len(t, linked1, 2)
	requireEqualsStrings(t,
		[]string{dymName11.Name, dymName22.Name},
		[]string{linked1[0].Name, linked1[1].Name},
	)

	linked2, err2 := dk.GetDymNamesContainsFallbackAddress(ctx, owner2Acc.bytes())
	require.NoError(t, err2)
	require.NotEqual(t, 2, len(linked2), "should not include non-existing dym-name")
	require.Len(t, linked2, 1)
	requireEqualsStrings(t,
		[]string{dymName21.Name},
		[]string{linked2[0].Name},
	)

	linkedByNotExists, err3 := dk.GetDymNamesContainsFallbackAddress(
		ctx,
		make([]byte, 20),
	)
	require.NoError(t, err3)
	require.Len(t, linkedByNotExists, 0)

	t.Run("allow Interchain Account (32 bytes)", func(t *testing.T) {
		require.NoError(
			t,
			dk.AddReverseMappingFallbackAddressToDymName(ctx, icaAcc.bytes(), dymName11.Name),
		)

		linked3, err := dk.GetDymNamesContainsFallbackAddress(ctx, icaAcc.bytes())
		require.NoError(t, err)
		require.Len(t, linked3, 1)
		require.Equal(t, dymName11.Name, linked3[0].Name)
	})
}

func TestKeeper_RemoveReverseMappingFallbackAddressToDymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	for size := 0; size <= 128; size++ {
		if size == 20 || size == 32 {
			continue // two valid size
		}

		bz := make([]byte, size)

		require.Errorf(
			t,
			dk.RemoveReverseMappingFallbackAddressToDymName(ctx, bz, "a"),
			"should not allow %d bytes address", size,
		)
	}

	ownerAcc := testAddr(1)
	anotherAcc := testAddr(2)
	someoneAcc := testAddr(3)
	icaAcc := testICAddr(4)

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerAcc.bech32(),
		Controller: ownerAcc.bech32(),
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: anotherAcc.bech32(),
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName1))
	err := dk.AddReverseMappingFallbackAddressToDymName(ctx, anotherAcc.bytes(), dymName1.Name)
	require.NoError(t, err)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerAcc.bech32(),
		Controller: ownerAcc.bech32(),
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: anotherAcc.bech32(),
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))
	err = dk.AddReverseMappingFallbackAddressToDymName(ctx, anotherAcc.bytes(), dymName2.Name)
	require.NoError(t, err)

	require.NoError(
		t,
		dk.RemoveReverseMappingFallbackAddressToDymName(ctx,
			someoneAcc.bytes(),
			dymName2.Name,
		),
		"no error if record not exists",
	)

	linked, err := dk.GetDymNamesContainsFallbackAddress(ctx, anotherAcc.bytes())
	require.NoError(t, err)
	require.Len(t, linked, 2, "existing data must be kept")

	t.Run("no error if element is not in the list", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingFallbackAddressToDymName(ctx, anotherAcc.bytes(), "not-in-list"),
		)
		linked, err = dk.GetDymNamesContainsFallbackAddress(ctx, anotherAcc.bytes())
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("remove correctly", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingFallbackAddressToDymName(ctx, anotherAcc.bytes(), dymName1.Name),
		)

		linked, err = dk.GetDymNamesContainsFallbackAddress(ctx, anotherAcc.bytes())
		require.NoError(t, err)
		require.Len(t, linked, 1)
		require.Equal(t, dymName2.Name, linked[0].Name)

		require.NoError(
			t,
			dk.RemoveReverseMappingFallbackAddressToDymName(ctx, anotherAcc.bytes(), dymName2.Name),
		)

		linked, err = dk.GetDymNamesContainsFallbackAddress(ctx, anotherAcc.bytes())
		require.NoError(t, err)
		require.Empty(t, linked)
	})

	t.Run("allow Interchain Account (32 bytes)", func(t *testing.T) {
		require.NoError(
			t,
			dk.AddReverseMappingFallbackAddressToDymName(ctx, icaAcc.bytes(), dymName1.Name),
		)

		linked3, err := dk.GetDymNamesContainsFallbackAddress(ctx, icaAcc.bytes())
		require.NoError(t, err)
		require.Len(t, linked3, 1)

		require.NoError(
			t,
			dk.RemoveReverseMappingFallbackAddressToDymName(ctx, icaAcc.bytes(), dymName1.Name),
		)
		linked3, err = dk.GetDymNamesContainsFallbackAddress(ctx, icaAcc.bytes())
		require.NoError(t, err)
		require.Empty(t, linked3)
	})
}

func requireEqualsStrings(t *testing.T, expected, actual []string) {
	t.Helper()

	require.Equal(t, len(expected), len(actual))

	sort.Strings(expected)
	sort.Strings(actual)

	require.Equal(t, expected, actual)
}
