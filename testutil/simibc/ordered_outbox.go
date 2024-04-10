package simibc

import channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

// The number of blocks to wait before a packet or ack is available for delivery
// after it has been committed on the chain.
// For example, if DelayPeriodBlocks is 0 and a packet p was sent at height h
// (i.e. the chain has produced a header height h)
// the packet can immediately be received.
// If DelayPeriodBlocks is 1, the packet can be received
// once the chain has produced a header for height h + 1.
const DelayPeriodBlocks = 1

// Ack represents a (sent) ack committed to block state
type Ack struct {
	Ack []byte
	// The packet to which this ack is a response
	Packet channeltypes.Packet
	// The number of App.Commits that have occurred since this ack was sent
	// For example, if the ack was sent at height h, and the blockchain
	// has headers ..., h, h+1, h+2 then Commits = 3
	Commits int
}

// Packet represents a (sent) packet committed to block state
type Packet struct {
	Packet channeltypes.Packet
	// The number of App.Commits that have occurred since this packet was sent
	// For example, if the ack was sent at height h, and the blockchain
	// has headers ..., h, h+1, h+2 then Commits = 3
	Commits int
}

// OrderedOutbox is a collection of ORDERED packets and acks that have been sent
// by different chains, but have not yet been delivered to their target.
// The methods take care of bookkeeping, making it easier to simulate
// a real relayed IBC connection.
//
// Each sent packet or ack can be added here. When a sufficient number of
// block commits have followed each sent packet or ack, they can be consumed:
// delivered to their target. Since the sequences are ordered, this is useful
// for testing ORDERED ibc channels.
//
// NOTE: OrderedOutbox MAY be used independently of the rest of simibc.
type OrderedOutbox struct {
	// An ordered sequence of packets from each sender
	OutboxPackets map[string][]Packet
	// An ordered sequence of acks from each sender
	OutboxAcks map[string][]Ack
}

// MakeOrderedOutbox creates a new empty OrderedOutbox.
func MakeOrderedOutbox() OrderedOutbox {
	return OrderedOutbox{
		OutboxPackets: map[string][]Packet{},
		OutboxAcks:    map[string][]Ack{},
	}
}

// AddPacket adds an outbound packet from the sender.
func (n OrderedOutbox) AddPacket(sender string, packet channeltypes.Packet) {
	n.OutboxPackets[sender] = append(n.OutboxPackets[sender], Packet{packet, 0})
}

// AddAck adds an outbound ack from the sender. The ack is a response to the packet.
func (n OrderedOutbox) AddAck(sender string, ack []byte, packet channeltypes.Packet) {
	n.OutboxAcks[sender] = append(n.OutboxAcks[sender], Ack{ack, packet, 0})
}

// ConsumePackets returns the first num packets with 2 or more commits. Returned
// packets are removed from the outbox and will not be returned again (consumed).
func (n OrderedOutbox) ConsumePackets(sender string, num int) []Packet {
	ret := []Packet{}
	sz := len(n.OutboxPackets[sender])
	if sz < num {
		num = sz
	}
	for _, p := range n.OutboxPackets[sender][:num] {
		if DelayPeriodBlocks < p.Commits {
			ret = append(ret, p)
		} else {
			break
		}
	}
	n.OutboxPackets[sender] = n.OutboxPackets[sender][len(ret):]
	return ret
}

// ConsumerAcks returns the first num packets with 2 or more commits. Returned
// acks are removed from the outbox and will not be returned again (consumed).
func (n OrderedOutbox) ConsumeAcks(sender string, num int) []Ack {
	ret := []Ack{}
	sz := len(n.OutboxAcks[sender])
	if sz < num {
		num = sz
	}
	for _, a := range n.OutboxAcks[sender][:num] {
		if 1 < a.Commits {
			ret = append(ret, a)
		} else {
			break
		}
	}
	n.OutboxAcks[sender] = n.OutboxAcks[sender][len(ret):]
	return ret
}

// Commit marks a block commit, increasing the commit count for all
// packets and acks in the sender's outbox.
// When a packet or ack has 2 or more commits, it is available for
// delivery to the counterparty chain.
// Note that 2 commits are necessary instead of 1:
//   - 1st commit is necessary for the packet to included in the block
//   - 2nd commit is necessary because in practice the ibc light client
//     needs to have block h + 1 to be able to verify the packet in block h.
func (n OrderedOutbox) Commit(sender string) {
	for i := range n.OutboxPackets[sender] {
		n.OutboxPackets[sender][i].Commits += 1
	}
	for i := range n.OutboxAcks[sender] {
		n.OutboxAcks[sender][i].Commits += 1
	}
}
