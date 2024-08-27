package types_test

import (
	"testing"
	"time"

	errorsmod "cosmossdk.io/errors"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

var (
	sequencerPubKey = ed25519.GenPrivKey().PubKey()
	tmPk, _         = cryptocodec.ToTmProtoPublicKey(sequencerPubKey)
	valHash, _      = types.GetValHashForSequencer(tmPk)
	timestamp       = time.Unix(1724392989, 0)

	validIBCState = types.IBCState{
		Root:               []byte("root"),
		Timestamp:          timestamp,
		NextValidatorsHash: valHash,
		ValidatorsHash:     valHash,
	}
	validRollappState = types.RollappState{
		BlockSequencer: tmPk,
		BlockDescriptor: rollapptypes.BlockDescriptor{
			StateRoot: []byte("root"),
			Timestamp: timestamp,
		},
		NextBlockSequencer: tmPk,
		NextBlockDescriptor: rollapptypes.BlockDescriptor{
			StateRoot: []byte("root2"),
			Timestamp: timestamp,
		},
	}
)

func TestCheckCompatibility(t *testing.T) {
	type input struct {
		ibcState types.IBCState
		raState  types.RollappState
	}
	testCases := []struct {
		name  string
		input func() input
		err   error
	}{
		{
			name: "roots are not equal",
			input: func() input {
				invalidRootRaState := validRollappState
				invalidRootRaState.BlockDescriptor.StateRoot = []byte("not same root")
				return input{
					ibcState: validIBCState,
					raState:  invalidRootRaState,
				}
			},
			err: errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "block descriptor state root does not match tendermint header app hash"),
		},
		{
			name: "validator who signed the block header is not the sequencer who submitted the block",
			input: func() input {
				newSequencer := ed25519.GenPrivKey().PubKey()
				newtmPk, err := cryptocodec.ToTmProtoPublicKey(newSequencer)
				require.NoError(t, err)
				invalidValidatorHashRAState := validRollappState
				invalidValidatorHashRAState.BlockSequencer = newtmPk
				return input{
					ibcState: validIBCState,
					raState:  invalidValidatorHashRAState,
				}
			},
			err: errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "validator does not match the sequencer"),
		},
		{
			name: "nextValidatorHash does not match the sequencer who submitted the next block descriptor",
			input: func() input {
				invalidNextValidatorHashIBCState := validIBCState
				invalidNextValidatorHashIBCState.NextValidatorsHash = []byte("wrong next validator hash")
				return input{
					ibcState: invalidNextValidatorHashIBCState,
					raState:  validRollappState,
				}
			},
			err: errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "next validator hash does not match the sequencer for h+1"),
		},
		{
			name: "timestamps is empty",
			input: func() input {
				emptyTimestampRAState := validRollappState
				emptyTimestampRAState.BlockDescriptor.Timestamp = time.Time{}
				emptyTimestampRAState.NextBlockDescriptor.Timestamp = time.Time{}
				return input{
					ibcState: validIBCState,
					raState:  emptyTimestampRAState,
				}
			},
			err: types.ErrTimestampNotFound,
		},
		{
			name: "timestamps are not equal",
			input: func() input {
				invalidTimestampRAState := validRollappState
				invalidTimestampRAState.BlockDescriptor.Timestamp = timestamp.Add(1)
				return input{
					ibcState: validIBCState,
					raState:  invalidTimestampRAState,
				}
			},
			err: errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "block descriptor timestamp does not match tendermint header timestamp"),
		},
		{
			name: "all fields are compatible",
			input: func() input {
				return input{
					ibcState: validIBCState,
					raState:  validRollappState,
				}
			},
			err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.input()
			err := types.CheckCompatibility(input.ibcState, input.raState)
			if err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
