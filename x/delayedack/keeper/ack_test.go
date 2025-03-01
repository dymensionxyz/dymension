package keeper_test

import (
	"encoding/json"
	"errors"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	chantypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func testNonDetAck(cdc codec.Codec, bz []byte) (*chantypes.Acknowledgement, error) {
	var ack chantypes.Acknowledgement
	if err := cdc.UnmarshalJSON(bz, &ack); err != nil {
		return nil, errorsmod.Wrapf(types.ErrUnknownRequest, "unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	return &ack, nil
}

// demonstrate unexpected behaviour of ack json
func (s *DelayedAckTestSuite) TestNonDetAck() {
	s.T().Skip()
	k, ctx := s.App.DelayedAckKeeper, s.Ctx
	j := `
	{
		"result": "dGVzdA==",
		"error": "Some error"
	}
		`
	bz := []byte(j)
	s.True(json.Valid(bz))
	_ = ctx
	for range 100 {
		ack, err := testNonDetAck(k.Cdc(), bz)
		s.T().Log(ack.Success(), err) // sometimes true, sometimes false
	}
}

// make sure ack is properly parsed
func (s *DelayedAckTestSuite) TestVerifiesAck() {
	k, ctx := s.App.DelayedAckKeeper, s.Ctx

	example := chantypes.NewErrorAcknowledgement(errors.New("new example"))
	j := k.Cdc().MustMarshalJSON(&example)
	s.T().Log(string(j))

	{

		j := `
	{
		"result": "dGVzdA==",
		"error": "ABCI code: 1: error handling packet: see events for details"
	}
		`
		bz := []byte(j)
		s.True(json.Valid(bz))

		ack, err := k.ReadAck(bz)
		s.Require().Error(err)
		s.True(strings.Contains(err.Error(), "to expected bytes")) // :(
		_ = ctx
		_ = ack
	}
	{

		j := `{"error":"ABCI code: 1: error handling packet: see events for details"}`
		bz := []byte(j)
		s.True(json.Valid(bz))

		ack, err := k.ReadAck(bz)
		s.Require().NoError(err)
		s.Require().False(ack.Success())
		_ = ctx
		_ = ack
	}
	{

		j := `{"result":"dGVzdA=="}`
		bz := []byte(j)
		s.True(json.Valid(bz))

		ack, err := k.ReadAck(bz)
		s.Require().NoError(err)
		s.Require().True(ack.Success())
		_ = ctx
		_ = ack
	}
}
