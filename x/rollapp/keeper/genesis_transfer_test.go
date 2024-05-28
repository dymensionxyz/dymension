package keeper

import (
	"encoding/json"
	"testing"

	errorsmod "cosmossdk.io/errors"

	sdkerrs "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/stretchr/testify/require"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestParseGenesisTransferDenom(t *testing.T) {
	toStr := func(m genesisTransferDenomMemo) string {
		bz, _ := json.Marshal(m)
		return string(bz)
	}

	t.Run("happy: it came from the rollapp chain and contained denom metadata", func(t *testing.T) {
		var m genesisTransferDenomMemo
		m.DoesNotOriginateFromUser = true
		m.DenomMetadata = banktypes.Metadata{
			Description: "foo",
		}
		memoStr := toStr(m)
		denom, err := ParseGenesisTransferDenom(memoStr)
		require.NoError(t, err)
		require.Equal(t, m.DenomMetadata.Description, denom.Description)
	})
	t.Run("attack: it contained denom metadata but was sent by a user, not the chain", func(t *testing.T) {
		var m genesisTransferDenomMemo
		m.DenomMetadata = banktypes.Metadata{
			Description: "foo",
		}
		memoStr := toStr(m)
		_, err := ParseGenesisTransferDenom(memoStr)
		require.True(t, errorsmod.IsOf(err, sdkerrs.ErrUnauthorized))
	})
}
