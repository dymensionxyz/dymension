package simibc

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
)

// RelayedPath is a wrapper around ibctesting.Path gives fine-grained
// control over delivery packets and acks, and client updates. Specifically,
// the path represents a bidirectional ORDERED channel between two chains.
// It is possible to control the precise order that packets and acks are
// delivered, and the precise independent and relative order and timing of
// new blocks on each chain.
type RelayedPath struct {
	t    *testing.T
	Path *ibctesting.Path
	// clientHeaders is a map from chainID to an ordered list of headers that
	// have been committed to that chain. The headers are used to update the
	// client of the counterparty chain.
	clientHeaders map[string][]*ibctmtypes.Header

	// TODO: Make this private and expose methods to add packets and acks.
	//       Currently, packets and acks are added directly to the outboxes,
	//       but we should hide this implementation detail.
	Outboxes OrderedOutbox
}

// MakeRelayedPath returns an initialized RelayedPath without any
// packets, acks or headers. Requires a fully initialised path where
// the connection and any channel handshakes have been COMPLETED.
func MakeRelayedPath(t *testing.T, path *ibctesting.Path) *RelayedPath {
	t.Helper()
	return &RelayedPath{
		t:             t,
		clientHeaders: map[string][]*ibctmtypes.Header{},
		Path:          path,
		Outboxes:      MakeOrderedOutbox(),
	}
}

// PacketSentByChain returns true if the packet belongs to this relayed path.
func (f *RelayedPath) PacketSentByChain(packet channeltypes.Packet, chainID string) bool {
	if chainID == f.Path.EndpointA.Chain.ChainID {
		return f.PacketSentByA(packet)
	} else if chainID == f.Path.EndpointB.Chain.ChainID {
		return f.PacketSentByB(packet)
	}
	return false
}

// PacketSentByA returns true if the given packet was sent by chain A on this path.
func (f *RelayedPath) PacketSentByA(packet channeltypes.Packet) bool {
	return packet.SourcePort == f.Path.EndpointA.ChannelConfig.PortID &&
		packet.SourceChannel == f.Path.EndpointA.ChannelID &&
		packet.DestinationPort == f.Path.EndpointB.ChannelConfig.PortID &&
		packet.DestinationChannel == f.Path.EndpointB.ChannelID
}

// PacketSentByB returns true if the given packet was sent by chain B on this path.
func (f *RelayedPath) PacketSentByB(packet channeltypes.Packet) bool {
	return packet.SourcePort == f.Path.EndpointB.ChannelConfig.PortID &&
		packet.SourceChannel == f.Path.EndpointB.ChannelID &&
		packet.DestinationPort == f.Path.EndpointA.ChannelConfig.PortID &&
		packet.DestinationChannel == f.Path.EndpointA.ChannelID
}

// AddPacket adds a packet to the outbox of the chain with chainID.
// It will fail if the chain is not involved in the relayed path,
// or if the packet does not belong to this path,
// i.e. if the pace
func (f *RelayedPath) AddPacket(chainID string, packet channeltypes.Packet) {
	if !f.InvolvesChain(chainID) {
		f.t.Fatal("in relayed path could not add packet to chain: ", chainID, " because it is not involved in the relayed path")
	}

	f.Outboxes.AddPacket(chainID, packet)
}

// AddClientHeader adds a client header to the chain with chainID.
// The header is used to update the client of the counterparty chain.
// It will fail if the chain is not involved in the relayed path.
func (f *RelayedPath) AddClientHeader(chainID string, header *ibctmtypes.Header) {
	if !f.InvolvesChain(chainID) {
		f.t.Fatal("in relayed path could not add client header to chain: ", chainID, " because it is not involved in the relayed path")
	}
	f.clientHeaders[chainID] = append(f.clientHeaders[chainID], header)
}

// Chain returns the chain with chainID
func (f *RelayedPath) Chain(chainID string) *ibctesting.TestChain {
	if f.Path.EndpointA.Chain.ChainID == chainID {
		return f.Path.EndpointA.Chain
	}
	if f.Path.EndpointB.Chain.ChainID == chainID {
		return f.Path.EndpointB.Chain
	}
	f.t.Fatal("no chain found in relayed path with chainID: ", chainID)
	return nil
}

// InvolvesChain returns true if the chain is involved in the relayed path,
// i.e. if it is either the source or destination chain.
func (f *RelayedPath) InvolvesChain(chainID string) bool {
	return f.Path.EndpointA.Chain.ChainID == chainID || f.Path.EndpointB.Chain.ChainID == chainID
}

// UpdateClient updates the chain with the latest sequence
// of available headers committed by the counterparty chain since
// the last call to UpdateClient (or all for the first call).
func (f *RelayedPath) UpdateClient(chainID string, expectExpiration bool) error {
	for _, header := range f.clientHeaders[f.Counterparty(chainID)] {
		err := UpdateReceiverClient(f.endpoint(f.Counterparty(chainID)), f.endpoint(chainID), header, expectExpiration)
		if err != nil {
			return err
		}
	}
	f.clientHeaders[f.Counterparty(chainID)] = []*ibctmtypes.Header{}
	return nil
}

// DeliverPackets delivers UP TO <num> packets to the chain which have been
// sent to it by the counterparty chain and are ready to be delivered.
//
// A packet is ready to be delivered if the sender chain has progressed
// a sufficient number of blocks since the packet was sent. This is because
// all sent packets must be committed to block state before they can be queried.
// Additionally, in practice, light clients require a header (h+1) to deliver a
// packet sent in header h.
//
// In order to deliver packets, the chain must have an up-to-date client
// of the counterparty chain. Ie. UpdateClient should be called before this.
//
// If expectError is true, we expect *each* packet to be delivered to cause an error.
func (f *RelayedPath) DeliverPackets(chainID string, num int, expectError bool) {
	for _, p := range f.Outboxes.ConsumePackets(f.Counterparty(chainID), num) {
		ack, err := TryRecvPacket(f.endpoint(f.Counterparty(chainID)), f.endpoint(chainID), p.Packet, expectError)
		if err != nil && !expectError {
			f.t.Fatal("Got an error from TryRecvPacket: ", err)
		} else {
			f.Outboxes.AddAck(chainID, ack, p.Packet)
		}
	}
}

// DeliverPackets delivers UP TO <num> acks to the chain which have been
// sent to it by the counterparty chain and are ready to be delivered.
//
// An ack is ready to be delivered if the sender chain has progressed
// a sufficient number of blocks since the ack was sent. This is because
// all sent acks must be committed to block state before they can be queried.
// Additionally, in practice, light clients require a header (h+1) to deliver
// an ack sent in header h.
//
// In order to deliver acks, the chain must have an up-to-date client
// of the counterparty chain. Ie. UpdateClient should be called before this.
func (f *RelayedPath) DeliverAcks(chainID string, num int) {
	for _, ack := range f.Outboxes.ConsumeAcks(f.Counterparty(chainID), num) {
		err := TryRecvAck(f.endpoint(f.Counterparty(chainID)), f.endpoint(chainID), ack.Packet, ack.Ack)
		if err != nil {
			f.t.Fatal("deliverAcks")
		}
	}
}

// Counterparty returns the chainID of the other chain,
// from the perspective of the given chain.
func (f *RelayedPath) Counterparty(chainID string) string {
	if f.Path.EndpointA.Chain.ChainID == chainID {
		return f.Path.EndpointB.Chain.ChainID
	}
	if f.Path.EndpointB.Chain.ChainID == chainID {
		return f.Path.EndpointA.Chain.ChainID
	}
	f.t.Fatal("no chain found in relayed path with chainID: ", chainID)
	return ""
}

// endpoint is a helper returning the endpoint for the chain
func (f *RelayedPath) endpoint(chainID string) *ibctesting.Endpoint {
	if chainID == f.Path.EndpointA.Chain.ChainID {
		return f.Path.EndpointA
	}
	if chainID == f.Path.EndpointB.Chain.ChainID {
		return f.Path.EndpointB
	}
	f.t.Fatal("no chain found in relayed path with chainID: ", chainID)
	return nil
}
