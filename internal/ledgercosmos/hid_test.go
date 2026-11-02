package ledgercosmos

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrameRoundTrip(t *testing.T) {
	for _, n := range []int{0, 1, 57, 58, 59, 200, 255 + 5} {
		apdu := bytes.Repeat([]byte{0xab}, n)
		reports := frame(apdu)
		for _, r := range reports {
			require.Len(t, r, packetSize)
		}
		reads := make(chan []byte, len(reports))
		for _, r := range reports {
			reads <- r
		}
		got, err := unframe(reads)
		require.NoError(t, err)
		require.Equal(t, apdu, got, "length %d", n)
	}
}

func TestUnframeRejectsWrongSequence(t *testing.T) {
	reports := frame(bytes.Repeat([]byte{1}, 100))
	reads := make(chan []byte, 2)
	reads <- reports[1]
	_, err := unframe(reads)
	require.ErrorContains(t, err, "unexpected report")
}
