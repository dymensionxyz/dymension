package ante_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cometprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
	comettypes "github.com/cometbft/cometbft/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibcsolomachine "github.com/cosmos/ibc-go/v7/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/ante"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func ConvertValidator(src comettypes.Validator) *cometprototypes.Validator {
	// TODO: surely this must already exist somewhere

	pk, err := cryptocodec.FromTmPubKeyInterface(src.PubKey)
	if err != nil {
		panic(err)
	}
	pkP, err := cryptocodec.ToTmProtoPublicKey(pk)
	if err != nil {
		panic(err)
	}
	dst := &cometprototypes.Validator{
		Address:          src.Address,
		VotingPower:      src.VotingPower,
		ProposerPriority: src.ProposerPriority,
		PubKey:           pkP,
	}
	return dst
}

func ConvertValidatorSet(src *comettypes.ValidatorSet) *cometprototypes.ValidatorSet {
	// TODO: surely this must already exist somewhere

	if src == nil {
		return nil
	}

	dst := &cometprototypes.ValidatorSet{
		Validators: make([]*cometprototypes.Validator, len(src.Validators)),
	}

	for i, validator := range src.Validators {
		dst.Validators[i] = ConvertValidator(*validator)
	}
	dst.TotalVotingPower = src.TotalVotingPower()
	dst.Proposer = ConvertValidator(*src.Proposer)

	return dst
}

func TestHandleMsgUpdateClientGood(t *testing.T) {
	k, ctx := keepertest.LightClientKeeper(t)
	testClientStates := map[string]exported.ClientState{
		"non-tm-client-id": &ibcsolomachine.ClientState{},
	}
	testClientStates[keepertest.CanonClientID] = &ibctm.ClientState{
		ChainId: keepertest.DefaultRollapp,
	}

	blocktimestamp := time.Unix(1724392989, 0)
	var trustedVals *cmtproto.ValidatorSet
	signedHeader := &cmtproto.SignedHeader{
		Header: &cmtproto.Header{
			AppHash:            []byte("appHash"),
			ProposerAddress:    keepertest.Alice.MustProposerAddr(),
			Time:               blocktimestamp,
			ValidatorsHash:     keepertest.Alice.MustValsetHash(),
			NextValidatorsHash: keepertest.Alice.MustValsetHash(),
			Height:             1,
		},
		Commit: &cmtproto.Commit{},
	}
	header := ibctm.Header{
		SignedHeader:      signedHeader,
		ValidatorSet:      ConvertValidatorSet(keepertest.Alice.MustValset()),
		TrustedHeight:     ibcclienttypes.MustParseHeight("1-1"),
		TrustedValidators: trustedVals,
	}

	rollapps := map[string]rollapptypes.Rollapp{
		keepertest.DefaultRollapp: {
			RollappId: keepertest.DefaultRollapp,
		},
	}
	stateInfos := map[string]map[uint64]rollapptypes.StateInfo{
		keepertest.DefaultRollapp: {
			1: {
				Sequencer: keepertest.Alice.Address,
				StateInfoIndex: rollapptypes.StateInfoIndex{
					Index: 1,
				},
				StartHeight: 1,
				NumBlocks:   2,
				BDs: rollapptypes.BlockDescriptors{
					BD: []rollapptypes.BlockDescriptor{
						{
							Height:    1,
							StateRoot: []byte("appHash"),
							Timestamp: header.SignedHeader.Header.Time,
						},
						{
							Height:    2,
							StateRoot: []byte("appHash2"),
							Timestamp: header.SignedHeader.Header.Time.Add(1),
						},
					},
				},
			},
		},
	}

	ibcclientKeeper := NewMockIBCClientKeeper(testClientStates)
	ibcchannelKeeper := NewMockIBCChannelKeeper(nil)
	rollappKeeper := NewMockRollappKeeper(rollapps, stateInfos)
	ibcMsgDecorator := ante.NewIBCMessagesDecorator(*k, ibcclientKeeper, ibcchannelKeeper, rollappKeeper)
	clientMsg, err := ibcclienttypes.PackClientMessage(&header)
	require.NoError(t, err)
	msg := &ibcclienttypes.MsgUpdateClient{
		ClientId:      keepertest.CanonClientID,
		ClientMessage: clientMsg,
		Signer:        "relayerAddr",
	}
	err = ibcMsgDecorator.HandleMsgUpdateClient(ctx, msg)
	require.NoError(t, err)
}

// TODO: bring back the rest of the old tests https://github.com/dymensionxyz/dymension/issues/1364
