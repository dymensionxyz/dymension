package keeper_test

import (
	"strings"
	"time"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) TestKeeper_GetAddReverseMappingOwnerToOwnedDymName() {
	s.Run("should not allow invalid owner address", func() {
		s.Require().Error(s.dymNsKeeper.AddReverseMappingOwnerToOwnedDymName(s.ctx, "0x", "a"))

		_, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, "0x")
		s.Require().Error(err)
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
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName11))

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName21))

	dymName22 := dymnstypes.DymName{
		Name:       "n22",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName22))

	s.Run("can add", func() {
		var err error

		err = s.dymNsKeeper.AddReverseMappingOwnerToOwnedDymName(s.ctx, owner1a, dymName11.Name)
		s.Require().NoError(err)

		err = s.dymNsKeeper.AddReverseMappingOwnerToOwnedDymName(s.ctx, owner2a, dymName21.Name)
		s.Require().NoError(err)

		err = s.dymNsKeeper.AddReverseMappingOwnerToOwnedDymName(s.ctx, owner2a, dymName22.Name)
		s.Require().NoError(err)
	})

	s.Run("can add non-existing dym-name", func() {
		err := s.dymNsKeeper.AddReverseMappingOwnerToOwnedDymName(s.ctx, owner2a, "not-exists")
		s.Require().NoError(err)
	})

	s.Run("no error when adding duplicated name", func() {
		for i := 0; i < 3; i++ {
			err := s.dymNsKeeper.AddReverseMappingOwnerToOwnedDymName(s.ctx, owner2a, dymName21.Name)
			s.Require().NoError(err)
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
				s.Require().NoError(
					s.dymNsKeeper.AddReverseMappingOwnerToOwnedDymName(s.ctx, owner2a, dymName11.Name),
					"no error if dym-name owned by another owner",
				)
				s.Require().NoError(
					s.dymNsKeeper.AddReverseMappingOwnerToOwnedDymName(s.ctx, owner2a, "non-existence"),
					"no error if dym-name owned by another owner",
				)
			},
			want: []string{dymName21.Name, dymName22.Name},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.preRun != nil {
				tt.preRun()
			}

			ownedDymNames, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, tt.owner)
			s.Require().NoError(err)

			s.requireDymNameList(ownedDymNames, tt.want)
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_RemoveReverseMappingOwnerToOwnedDymName() {
	s.Require().Error(
		s.dymNsKeeper.RemoveReverseMappingOwnerToOwnedDymName(s.ctx, "0x", "a"),
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
	s.setDymNameWithFunctionsAfter(dymName11)

	dymName12 := dymnstypes.DymName{
		Name:       "n12",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.setDymNameWithFunctionsAfter(dymName12)

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.setDymNameWithFunctionsAfter(dymName21)

	s.Require().NoError(
		s.dymNsKeeper.RemoveReverseMappingOwnerToOwnedDymName(s.ctx, notOwnerA, dymName11.Name),
		"no error if owner non-exists",
	)

	s.Require().NoError(
		s.dymNsKeeper.RemoveReverseMappingOwnerToOwnedDymName(s.ctx, owner1a, dymName21.Name),
		"no error if not owned dym-name",
	)
	ownedBy, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, owner1a)
	s.Require().NoError(err)

	// existing data must be kept
	s.requireDymNameList(ownedBy, []string{dymName11.Name, dymName12.Name})

	s.Require().NoError(
		s.dymNsKeeper.RemoveReverseMappingOwnerToOwnedDymName(s.ctx, owner1a, "not-exists"),
		"no error if not-exists dym-name",
	)
	ownedBy, err = s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, owner1a)
	s.Require().NoError(err)
	// existing data must be kept
	s.requireDymNameList(ownedBy, []string{dymName11.Name, dymName12.Name})

	s.Require().NoError(
		s.dymNsKeeper.RemoveReverseMappingOwnerToOwnedDymName(s.ctx, owner1a, dymName11.Name),
	)
	ownedBy, err = s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, owner1a)
	s.Require().NoError(err)
	s.requireDymNameList(ownedBy, []string{dymName12.Name})

	s.Require().NoError(
		s.dymNsKeeper.RemoveReverseMappingOwnerToOwnedDymName(s.ctx, owner1a, dymName12.Name),
	)
	ownedBy, err = s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, owner1a)
	s.Require().NoError(err)
	s.Require().Len(ownedBy, 0)
}

func (s *KeeperTestSuite) TestKeeper_GetAddReverseMappingConfiguredAddressToDymName() {
	s.Run("fail - should reject blank address", func() {
		s.Require().Error(s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, " ", "a"))

		_, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, " ")
		s.Require().Error(err)
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
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName11))
	err := s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, anotherA, dymName11.Name)
	s.Require().NoError(err)

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName21))
	err = s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, owner2a, dymName21.Name)
	s.Require().NoError(err)

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
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName22))
	err = s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, anotherA, dymName22.Name)
	s.Require().NoError(err)

	s.Require().NoError(
		s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, anotherA, "not-exists"),
		"no check non-existing dym-name",
	)

	s.Run("no error if duplicated name", func() {
		for i := 0; i < 3; i++ {
			s.Require().NoError(
				s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, owner2a, dymName21.Name),
			)
		}
	})

	linked1, err1 := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, anotherA)
	s.Require().NoError(err1)
	s.Require().Len(linked1, 2)
	s.requireDymNameList(linked1, []string{dymName11.Name, dymName22.Name})

	linked2, err2 := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, owner2a)
	s.Require().NoError(err2)
	s.Require().NotEqual(2, len(linked2), "should not include non-existing dym-name")
	s.Require().Len(linked2, 1)
	s.Require().Equal(dymName21.Name, linked2[0].Name)

	linkedByNotExists, err3 := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, someoneA)
	s.Require().NoError(err3)
	s.Require().Len(linkedByNotExists, 0)

	s.Run("allow Interchain Account (32 bytes)", func() {
		s.Require().NoError(
			s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, icaA, dymName11.Name),
		)

		linked3, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, icaA)
		s.Require().NoError(err)
		s.Require().Len(linked3, 1)
		s.Require().Equal(dymName11.Name, linked3[0].Name)
	})

	s.Run("insert and get must be case-sensitive", func() {
		addr1 := strings.ToLower(owner1a)
		addr2 := strings.ToUpper(owner1a)
		addr3 := strings.ToLower(owner1a[:10]) + strings.ToUpper(owner1a[10:20]) + strings.ToLower(owner1a[20:])

		dymName := dymnstypes.DymName{
			Name:       "my-name",
			Owner:      owner1a,
			Controller: owner1a,
			ExpireAt:   s.ctx.BlockTime().Add(time.Hour).Unix(),
		}
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName))

		err = s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, addr1, dymName.Name)
		s.Require().NoError(err)
		err = s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, addr2, dymName.Name)
		s.Require().NoError(err)
		err = s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, addr3, dymName.Name)
		s.Require().NoError(err)

		linked1, err1 := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, addr1)
		s.Require().NoError(err1)
		s.requireDymNameList(linked1, []string{dymName.Name})

		linked2, err2 := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, addr2)
		s.Require().NoError(err2)
		s.requireDymNameList(linked2, []string{dymName.Name})

		linked3, err3 := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, addr3)
		s.Require().NoError(err3)
		s.requireDymNameList(linked3, []string{dymName.Name})
	})
}

func (s *KeeperTestSuite) TestKeeper_RemoveReverseMappingConfiguredAddressToDymName() {
	s.Require().Error(
		s.dymNsKeeper.RemoveReverseMappingConfiguredAddressToDymName(s.ctx, " ", "a"),
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
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName1))
	err := s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, anotherA, dymName1.Name)
	s.Require().NoError(err)

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
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName2))
	err = s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, anotherA, dymName2.Name)
	s.Require().NoError(err)

	s.Require().NoError(
		s.dymNsKeeper.RemoveReverseMappingConfiguredAddressToDymName(s.ctx, someoneA, dymName2.Name),
		"no error if record not exists",
	)

	linked, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, anotherA)
	s.Require().NoError(err)
	s.Require().Len(linked, 2, "existing data must be kept")

	s.Run("no error if element is not in the list", func() {
		s.Require().NoError(
			s.dymNsKeeper.RemoveReverseMappingConfiguredAddressToDymName(s.ctx, anotherA, "not-exists"),
		)
		linked, err = s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, anotherA)
		s.Require().NoError(err)
		s.Require().Len(linked, 2, "existing data must be kept")
	})

	s.Run("remove correctly", func() {
		s.Require().NoError(
			s.dymNsKeeper.RemoveReverseMappingConfiguredAddressToDymName(s.ctx, anotherA, dymName1.Name),
		)

		linked, err = s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, anotherA)
		s.Require().NoError(err)
		s.Require().Len(linked, 1)
		s.Require().Equal(dymName2.Name, linked[0].Name)

		s.Require().NoError(
			s.dymNsKeeper.RemoveReverseMappingConfiguredAddressToDymName(s.ctx, anotherA, dymName2.Name),
		)

		linked, err = s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, anotherA)
		s.Require().NoError(err)
		s.Require().Empty(linked)
	})

	s.Run("remove correctly with Interchain Account (32 bytes)", func() {
		s.Require().NoError(
			s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, icaA, dymName1.Name),
		)

		linked3, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, icaA)
		s.Require().NoError(err)
		s.Require().Len(linked3, 1)

		s.Require().NoError(
			s.dymNsKeeper.RemoveReverseMappingConfiguredAddressToDymName(s.ctx, icaA, dymName1.Name),
		)

		linked, err = s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, icaA)
		s.Require().NoError(err)
		s.Require().Empty(linked)
	})

	s.Run("remove must be case-sensitive", func() {
		addr1 := strings.ToLower(ownerA)
		addr2 := strings.ToUpper(ownerA)
		addr3 := strings.ToLower(ownerA[:10]) + strings.ToUpper(ownerA[10:20]) + strings.ToLower(ownerA[20:])

		dymName := dymnstypes.DymName{
			Name:       "my-name",
			Owner:      ownerA,
			Controller: ownerA,
			ExpireAt:   s.ctx.BlockTime().Add(time.Hour).Unix(),
		}
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName))

		err = s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, addr1, dymName.Name)
		s.Require().NoError(err)
		err = s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, addr2, dymName.Name)
		s.Require().NoError(err)
		err = s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, addr3, dymName.Name)
		s.Require().NoError(err)

		linked1, err1 := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, addr1)
		s.Require().NoError(err1)
		s.requireDymNameList(linked1, []string{dymName.Name})

		linked2, err2 := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, addr2)
		s.Require().NoError(err2)
		s.requireDymNameList(linked2, []string{dymName.Name})

		linked3, err3 := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, addr3)
		s.Require().NoError(err3)
		s.requireDymNameList(linked3, []string{dymName.Name})

		err = s.dymNsKeeper.RemoveReverseMappingConfiguredAddressToDymName(s.ctx, addr3, dymName.Name)
		s.Require().NoError(err)

		linked3, err3 = s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, addr3)
		s.Require().NoError(err3)
		s.Require().Empty(linked3)
	})
}

func (s *KeeperTestSuite) TestKeeper_GetAddReverseMappingFallbackAddressToDymName() {
	for size := 0; size <= 128; size++ {
		if size == 20 || size == 32 {
			continue // two valid size
		}

		addr := make([]byte, size)

		s.Require().Errorf(
			s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, addr, "a"),
			"should not allow %d bytes address", size,
		)

		_, err := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, addr)
		s.Require().Errorf(
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
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName11))
	err := s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, anotherAcc.bytes(), dymName11.Name)
	s.Require().NoError(err)

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2Acc.bech32(),
		Controller: owner2Acc.bech32(),
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName21))
	err = s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(
		s.ctx,
		owner2Acc.bytes(),
		dymName21.Name,
	)
	s.Require().NoError(err)

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
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName22))
	err = s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, anotherAcc.bytes(), dymName22.Name)
	s.Require().NoError(err)

	s.Require().NoError(
		s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, anotherAcc.bytes(), "not-exists"),
		"no check non-existing dym-name",
	)

	s.Run("no error if duplicated name", func() {
		for i := 0; i < 3; i++ {
			s.Require().NoError(
				s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, owner2Acc.bytes(), dymName21.Name),
			)
		}
	})

	linked1, err1 := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, anotherAcc.bytes())
	s.Require().NoError(err1)
	s.Require().Len(linked1, 2)
	s.requireDymNameList(linked1, []string{dymName11.Name, dymName22.Name})

	linked2, err2 := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, owner2Acc.bytes())
	s.Require().NoError(err2)
	s.Require().NotEqual(2, len(linked2), "should not include non-existing dym-name")
	s.Require().Len(linked2, 1)
	s.Require().Equal(dymName21.Name, linked2[0].Name)

	linkedByNotExists, err3 := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(
		s.ctx,
		make([]byte, 20),
	)
	s.Require().NoError(err3)
	s.Require().Len(linkedByNotExists, 0)

	s.Run("allow Interchain Account (32 bytes)", func() {
		s.Require().NoError(
			s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, icaAcc.bytes(), dymName11.Name),
		)

		linked3, err := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, icaAcc.bytes())
		s.Require().NoError(err)
		s.Require().Len(linked3, 1)
		s.Require().Equal(dymName11.Name, linked3[0].Name)
	})
}

func (s *KeeperTestSuite) TestKeeper_RemoveReverseMappingFallbackAddressToDymName() {
	for size := 0; size <= 128; size++ {
		if size == 20 || size == 32 {
			continue // two valid size
		}

		bz := make([]byte, size)

		s.Require().Errorf(
			s.dymNsKeeper.RemoveReverseMappingFallbackAddressToDymName(s.ctx, bz, "a"),
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
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName1))
	err := s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, anotherAcc.bytes(), dymName1.Name)
	s.Require().NoError(err)

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
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName2))
	err = s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, anotherAcc.bytes(), dymName2.Name)
	s.Require().NoError(err)

	s.Require().NoError(
		s.dymNsKeeper.RemoveReverseMappingFallbackAddressToDymName(s.ctx,
			someoneAcc.bytes(),
			dymName2.Name,
		),
		"no error if record not exists",
	)

	linked, err := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, anotherAcc.bytes())
	s.Require().NoError(err)
	s.Require().Len(linked, 2, "existing data must be kept")

	s.Run("no error if element is not in the list", func() {
		s.Require().NoError(
			s.dymNsKeeper.RemoveReverseMappingFallbackAddressToDymName(s.ctx, anotherAcc.bytes(), "not-in-list"),
		)
		linked, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, anotherAcc.bytes())
		s.Require().NoError(err)
		s.Require().Len(linked, 2, "existing data must be kept")
	})

	s.Run("remove correctly", func() {
		s.Require().NoError(
			s.dymNsKeeper.RemoveReverseMappingFallbackAddressToDymName(s.ctx, anotherAcc.bytes(), dymName1.Name),
		)

		linked, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, anotherAcc.bytes())
		s.Require().NoError(err)
		s.Require().Len(linked, 1)
		s.Require().Equal(dymName2.Name, linked[0].Name)

		s.Require().NoError(
			s.dymNsKeeper.RemoveReverseMappingFallbackAddressToDymName(s.ctx, anotherAcc.bytes(), dymName2.Name),
		)

		linked, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, anotherAcc.bytes())
		s.Require().NoError(err)
		s.Require().Empty(linked)
	})

	s.Run("allow Interchain Account (32 bytes)", func() {
		s.Require().NoError(
			s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, icaAcc.bytes(), dymName1.Name),
		)

		linked3, err := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, icaAcc.bytes())
		s.Require().NoError(err)
		s.Require().Len(linked3, 1)

		s.Require().NoError(
			s.dymNsKeeper.RemoveReverseMappingFallbackAddressToDymName(s.ctx, icaAcc.bytes(), dymName1.Name),
		)
		linked3, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, icaAcc.bytes())
		s.Require().NoError(err)
		s.Require().Empty(linked3)
	})
}
