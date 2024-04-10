package simibc

import (
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	simapp "github.com/cosmos/ibc-go/v6/testing/simapp"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmtypes "github.com/tendermint/tendermint/types"
)

// UpdateReceiverClient DELIVERs a header to the receiving endpoint
// and update the respective client of the receiving chain.
//
// The header is a header of the sender chain. The receiver chain
// must have a client of the sender chain that it can update.
//
// NOTE: this function MAY be used independently of the rest of simibc.
func UpdateReceiverClient(sender, receiver *ibctesting.Endpoint, header *ibctmtypes.Header, expectExpiration bool) (err error) {
	err = augmentHeader(sender.Chain, receiver.Chain, receiver.ClientID, header)
	if err != nil {
		return err
	}

	msg, err := clienttypes.NewMsgUpdateClient(
		receiver.ClientID, header,
		receiver.Chain.SenderAccount.GetAddress().String(),
	)
	if err != nil {
		return err
	}

	_, _, err = simapp.SignAndDeliver(
		receiver.Chain.T,
		receiver.Chain.TxConfig,
		receiver.Chain.App.GetBaseApp(),
		receiver.Chain.GetContext().BlockHeader(),
		[]sdk.Msg{msg},
		receiver.Chain.ChainID,
		[]uint64{receiver.Chain.SenderAccount.GetAccountNumber()},
		[]uint64{receiver.Chain.SenderAccount.GetSequence()},
		true, !expectExpiration, receiver.Chain.SenderPrivKey,
	)

	setSequenceErr := receiver.Chain.SenderAccount.SetSequence(receiver.Chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return err
	}

	if setSequenceErr != nil {
		return setSequenceErr
	}

	return nil
}

// TryRecvPacket will try once to DELIVER a packet from sender to receiver. If successful,
// it will return the acknowledgement bytes.
//
// The packet must be sent from the sender chain to the receiver chain, and the
// receiver chain must have a client for the sender chain which has been updated
// to a recent height of the sender chain so that it can verify the packet.
func TryRecvPacket(sender, receiver *ibctesting.Endpoint, packet channeltypes.Packet, expectError bool) (ack []byte, err error) {
	packetKey := host.PacketCommitmentKey(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	proof, proofHeight := sender.Chain.QueryProof(packetKey)

	RPmsg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, receiver.Chain.SenderAccount.GetAddress().String())

	_, resWithAck, err := simapp.SignAndDeliver(
		receiver.Chain.T,
		receiver.Chain.TxConfig,
		receiver.Chain.App.GetBaseApp(),
		receiver.Chain.GetContext().BlockHeader(),
		[]sdk.Msg{RPmsg},
		receiver.Chain.ChainID,
		[]uint64{receiver.Chain.SenderAccount.GetAccountNumber()},
		[]uint64{receiver.Chain.SenderAccount.GetSequence()},
		true, !expectError, receiver.Chain.SenderPrivKey,
	)
	// need to set the sequence even if there was an error in delivery
	setSequenceErr := receiver.Chain.SenderAccount.SetSequence(receiver.Chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	if setSequenceErr != nil {
		return nil, setSequenceErr
	}

	ack, err = ibctesting.ParseAckFromEvents(resWithAck.GetEvents())
	if err != nil {
		return nil, err
	}

	return ack, nil
}

// TryRecvAck will try once to DELIVER an ack from sender to receiver.
//
// The ack must have been sent from the sender to the receiver, in response
// to packet which was previously delivered from the receiver to the sender.
// The receiver chain must have a client for the sender chain which has been
// updated to a recent height of the sender chain so that it can verify the packet.
func TryRecvAck(sender, receiver *ibctesting.Endpoint, packet channeltypes.Packet, ack []byte) (err error) {
	p := packet
	packetKey := host.PacketAcknowledgementKey(p.GetDestPort(), p.GetDestChannel(), p.GetSequence())
	proof, proofHeight := sender.Chain.QueryProof(packetKey)

	ackMsg := channeltypes.NewMsgAcknowledgement(p, ack, proof, proofHeight, receiver.Chain.SenderAccount.GetAddress().String())

	_, _, err = simapp.SignAndDeliver(
		receiver.Chain.T,
		receiver.Chain.TxConfig,
		receiver.Chain.App.GetBaseApp(),
		receiver.Chain.GetContext().BlockHeader(),
		[]sdk.Msg{ackMsg},
		receiver.Chain.ChainID,
		[]uint64{receiver.Chain.SenderAccount.GetAccountNumber()},
		[]uint64{receiver.Chain.SenderAccount.GetSequence()},
		true, true, receiver.Chain.SenderPrivKey,
	)

	setSequenceErr := receiver.Chain.SenderAccount.SetSequence(receiver.Chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return err
	}

	if setSequenceErr != nil {
		return setSequenceErr
	}

	return nil
}

// augmentHeader is a helper that augments the header with the height and validators that are most recently trusted
// by the receiver chain. If there is an error, the header will not be modified.
func augmentHeader(sender, receiver *ibctesting.TestChain, clientID string, header *ibctmtypes.Header) error {
	trustedHeight := receiver.GetClientState(clientID).GetLatestHeight().(clienttypes.Height)

	var (
		tmTrustedVals *tmtypes.ValidatorSet
		ok            bool
	)
	// Once we get TrustedHeight from client, we must query the validators from the counterparty chain
	// If the LatestHeight == LastHeader.Height, then TrustedValidators are current validators
	// If LatestHeight < LastHeader.Height, we can query the historical validator set from HistoricalInfo
	if trustedHeight == sender.LastHeader.GetHeight() {
		tmTrustedVals = sender.Vals
	} else {
		// NOTE: We need to get validators from counterparty at height: trustedHeight+1
		// since the last trusted validators for a header at height h
		// is the NextValidators at h+1 committed to in header h by
		// NextValidatorsHash
		tmTrustedVals, ok = sender.GetValsAtHeight(int64(trustedHeight.RevisionHeight + 1))
		if !ok {
			return errorsmod.Wrapf(ibctmtypes.ErrInvalidHeaderHeight, "could not retrieve trusted validators at trustedHeight: %d", trustedHeight)
		}
	}
	trustedVals, err := tmTrustedVals.ToProto()
	if err != nil {
		return err
	}
	// inject trusted fields into last header
	// for now assume revision number is 0
	header.TrustedHeight = trustedHeight
	header.TrustedValidators = trustedVals

	return nil
}
