package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgNonAvailableBatch_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgNonAvailableBatch
		err  error
	}{
		{
			name: "valid non inclusion proof",
			msg: MsgNonAvailableBatch{
				Creator:           sample.AccAddress(),
				RollappId:         "rollapptest_123-1",
				SlIndex:           0,
				DAPath:            "celestia.1.2.4.aa5b76fe9c42a5aff1fcfe1cc5088b3941cb1cc854c22ce6c0c0fb98a5461f8e.e06c57a64b049d6463ef.T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=",
				NonInclusionProof: "{\"rproofs\": \"CCAaIEPfyrlgwKe73bHoJPu/s3g/cqpCD+GSBhprXt6dQKvfIiCxvoDp0he9J928CKGi9YRyq6e/TO3VhMmtf31kZeA9XSIgVQgl6ThdDbM23Jb12Qpj2432o8n326zTSifjYdENdwMiIK2G2DYjwb3UZvotBTsSD++LddcxR75oczW1aJfz4fiyIiDWJEaMImR0qENxIHxcF3dTz4zl4JNLS4qU/bst0GdwsSIgvHPiMyjd1cVH4kYeTZrUUQQfxAeiDy96jyGO2Zhnlmg=\", \"dataroot\": \"T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=\"}",
			},
		},
		{
			name: "non valid proof",
			msg: MsgNonAvailableBatch{
				Creator:           sample.AccAddress(),
				RollappId:         "rollapptest_123-1",
				SlIndex:           0,
				DAPath:            "",
				NonInclusionProof: "",
			},
			err: sdkerrors.ErrInvalidRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
