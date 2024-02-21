package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestWrongCommitmentBatch_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgWrongCommitmentBatch
		err  error
	}{
		{
			name: "valid inclusion proof",
			msg: MsgWrongCommitmentBatch{
				Creator:        sample.AccAddress(),
				RollappId:      "rollapptest_123-1",
				SlIndex:        0,
				DAPath:         "",
				InclusionProof: "{\"blob\":\"\",\"nmtproofs\":[\"eyJzdGFydCI6MSwiZW5kIjo4LCJub2RlcyI6WyJBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQVFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQkZ4eUh2eEJrUWptVGRJTFg5c0UzVUk5MmVKOFAxcFo2bFk4dERLT0R5eVIiLCIvLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLzE2UlFoMFVEREkvS1lCS0lEQ2krYnhOTzZyb0ZqYUY0Q1hwUDkwZmdNQ0giXSwiaXNfbWF4X25hbWVzcGFjZV9pZ25vcmVkIjp0cnVlfQ==\",\"eyJlbmQiOjgsIm5vZGVzIjpbIi8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vL3pnMVFadisxaGhYdkFsaUJDcGlrVU9ZMDA2dStpQ1Axd1J4eG13R3dCeiJdLCJpc19tYXhfbmFtZXNwYWNlX2lnbm9yZWQiOnRydWV9\",\"eyJlbmQiOjcsIm5vZGVzIjpbIi8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vNy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vcGxFcWdSL2M0SUFWa05kWVJXT1lPQUVTRDR3aG5lS1I1NER6NURmZTRwMiIsIi8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vLy8vN01VeUNVY0VoY3VKc3JmVWR1a0phWlhlYTR5TEx2cGVvRThpVlJWZnpoZyJdLCJpc19tYXhfbmFtZXNwYWNlX2lnbm9yZWQiOnRydWV9\"],\"nmtroots\":[\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAA4GxXpksEnWRj7/LHlmIIxMYh7zsSKa9efrarvjvfiLNcfQgeyTIO0gub\",\"AAAAAAAAAAAAAAAAAAAAAAAAAOBsV6ZLBJ1kY+8AAAAAAAAAAAAAAAAAAAAAAAAA4GxXpksEnWRj7yRfa8M1PPrhJVFJ9REHeGdbXzRe5B09V29KeJDiCgff\",\"AAAAAAAAAAAAAAAAAAAAAAAAAOBsV6ZLBJ1kY+///////////////////////////////////////kB2GVTGeRRurn0EVPMkqmE47DZak0WiAIUtWYuNYVP5\"],\"rproofs\":[\"CCAaIJEj8f2IPQclrU2Idx3MUPvS6w+uYLsDI8Ra/V4FazqmIiDy+Px4UL5PvrKaDurYmnviREczHIR5IwJHqxI3LKKSDCIguFOX9794Pc38TRp4WRmaBekhBDGYEMP15NoGWnncib0iIK2G2DYjwb3UZvotBTsSD++LddcxR75oczW1aJfz4fiyIiD3uhkhupctQ422YZnvME78sHZn4BBoNff2m0bMQT81OyIg39Yb59vgAR5QHpQQowJsWcwcWc6UxcttOijHef0gcwo=\",\"CCAQARog8vj8eFC+T76ymg7q2Jp74kRHMxyEeSMCR6sSNyyikgwiIJEj8f2IPQclrU2Idx3MUPvS6w+uYLsDI8Ra/V4FazqmIiC4U5f3v3g9zfxNGnhZGZoF6SEEMZgQw/Xk2gZaedyJvSIgrYbYNiPBvdRm+i0FOxIP74t11zFHvmhzNbVol/Ph+LIiIPe6GSG6ly1DjbZhme8wTvywdmfgEGg19/abRsxBPzU7IiDf1hvn2+ABHlAelBCjAmxZzBxZzpTFy206KMd5/SBzCg==\",\"CCAQAhogxA6mrY8XWkKLXH/PTaZz1CeGwlAjsGjWQ+5Q7+DOHU0iID7qb6DuN7B4M8BBaYNDa8t1GtdRbfubuF6CIenr9So2IiCCH6RsEXOyXgISuSD0kEAFZbD22bSIV1cKtta41AMwoyIgrYbYNiPBvdRm+i0FOxIP74t11zFHvmhzNbVol/Ph+LIiIPe6GSG6ly1DjbZhme8wTvywdmfgEGg19/abRsxBPzU7IiDf1hvn2+ABHlAelBCjAmxZzBxZzpTFy206KMd5/SBzCg==\"],\"dataroot\":\"V6G1mPCsYwQevmqXpQltHKqBNP8ZL9N0hHpzQx4ipL0=\"}",
			},
		},
		{
			name: "non valid proof",
			msg: MsgWrongCommitmentBatch{
				Creator:        sample.AccAddress(),
				RollappId:      "rollapptest_123-1",
				SlIndex:        0,
				DAPath:         "",
				InclusionProof: "",
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
