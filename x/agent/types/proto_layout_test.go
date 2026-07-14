package types_test

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

func TestAgentProtoFieldNumbers(t *testing.T) {
	bz, err := proto.Marshal(&types.Agent{
		Id:        "i",
		Owner:     "o",
		Policy:    tee.Policy{GcpRootCertPem: "p"},
		Active:    true,
		ActionSeq: 1,
	})
	require.NoError(t, err)
	require.Equal(t, []byte{0x0a, 0x01, 'i', 0x12, 0x03, 0x0a, 0x01, 'p', 0x18, 0x01, 0x20, 0x01, 0x2a, 0x01, 'o'}, bz)
}

func TestParamsProtoFieldNumbers(t *testing.T) {
	bz, err := proto.Marshal(&types.Params{
		AgentRegistrationFee: types.DefaultParams().AgentRegistrationFee,
		MaxActionBytes:       1,
	})
	require.NoError(t, err)
	require.Equal(t, []byte{0x08, 0x01}, bz[:2])
	require.Equal(t, byte(0x12), bz[2])
}
