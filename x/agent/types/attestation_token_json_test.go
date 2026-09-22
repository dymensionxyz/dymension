package types_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/stretchr/testify/require"
)

func TestAttestationTokensUseRawStringJSON(t *testing.T) {
	// A deliberately unsigned fixture with JWT separators catches accidental base64 decoding.
	const token = "header.payload.signature"
	for _, msg := range []proto.Message{&types.MsgSubmitAttestedAction{}, &types.MsgSubmitAttestedTransfer{}, &types.MsgRespondValidationAttested{}} {
		t.Run(proto.MessageName(msg), func(t *testing.T) {
			input := `{"token":"` + token + `"}`
			require.NoError(t, codec.NewProtoCodec(nil).UnmarshalJSON([]byte(input), msg))
			output, err := codec.ProtoMarshalJSON(msg, nil)
			require.NoError(t, err)
			var fields map[string]any
			require.NoError(t, json.Unmarshal(output, &fields))
			require.Equal(t, token, fields["token"])
			t.Logf("token JSON round-trip: %s", fields["token"])
		})
	}
}
