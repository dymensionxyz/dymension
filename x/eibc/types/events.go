package types

import (
	"encoding/base64"
)

func GetCreatedEvent(m *DemandOrder, proofHeight uint64, amount string) *EventDemandOrderCreated {
	packetKey := base64.StdEncoding.EncodeToString([]byte(m.TrackingPacketKey))
	return &EventDemandOrderCreated{
		OrderId:      m.Id,
		Price:        m.Price.String(),
		Fee:          m.Fee.String(),
		PacketStatus: m.TrackingPacketStatus.String(),
		PacketKey:    packetKey,
		RollappId:    m.RollappId,
		Recipient:    m.Recipient,
		PacketType:   m.Type.String(),
		ProofHeight:  proofHeight,
		Amount:       amount,
	}
}

func GetFulfilledEvent(m *DemandOrder) *EventDemandOrderFulfilled {
	return &EventDemandOrderFulfilled{
		OrderId:      m.Id,
		Price:        m.Price.String(),
		Fee:          m.Fee.String(),
		IsFulfilled:  true,
		PacketStatus: m.TrackingPacketStatus.String(),
		Fulfiller:    m.FulfillerAddress,
		PacketType:   m.Type.String(),
	}
}

func GetFulfilledAuthorizedEvent(m *DemandOrder,
	creationHeight uint64,
	lpAddress, operatorAddress, operatorFee string,
) *EventDemandOrderFulfilledAuthorized {
	return &EventDemandOrderFulfilledAuthorized{
		OrderId:         m.Id,
		Price:           m.Price.String(),
		Fee:             m.Fee.String(),
		IsFulfilled:     true,
		PacketStatus:    m.TrackingPacketStatus.String(),
		Fulfiller:       m.FulfillerAddress,
		PacketType:      m.Type.String(),
		CreationHeight:  creationHeight,
		LpAddress:       lpAddress,
		OperatorAddress: operatorAddress,
		OperatorFee:     operatorFee,
	}
}

func GetUpdatedEvent(m *DemandOrder, proofHeight uint64, amount string) *EventDemandOrderFeeUpdated {
	return &EventDemandOrderFeeUpdated{
		OrderId:      m.Id,
		NewFee:       m.Fee.String(),
		Price:        m.Price.String(),
		PacketStatus: m.TrackingPacketStatus.String(),
		RollappId:    m.RollappId,
		ProofHeight:  proofHeight,
		Amount:       amount,
	}
}

func GetPacketStatusUpdatedEvent(m *DemandOrder) *EventDemandOrderPacketStatusUpdated {
	return &EventDemandOrderPacketStatusUpdated{
		OrderId:         m.Id,
		NewPacketStatus: m.TrackingPacketStatus,
		IsFulfilled:     m.IsFulfilled(),
	}
}
