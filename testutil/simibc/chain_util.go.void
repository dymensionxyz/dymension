package simibc

import (
	"time"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// BeginBlock updates the current header and calls the app.BeginBlock method.
// The new block height is the previous block height + 1.
// The new block time is the previous block time + dt.
//
// NOTE: this method may be used independently of the rest of simibc.
func BeginBlock(c *ibctesting.TestChain, dt time.Duration) {
	c.CurrentHeader = tmproto.Header{
		ChainID:            c.ChainID,
		Height:             c.App.LastBlockHeight() + 1,
		AppHash:            c.App.LastCommitID().Hash,
		Time:               c.CurrentHeader.Time.Add(dt),
		ValidatorsHash:     c.Vals.Hash(),
		NextValidatorsHash: c.NextVals.Hash(),
	}

	_ = c.App.BeginBlock(abci.RequestBeginBlock{Header: c.CurrentHeader})
}

// EndBlock calls app.EndBlock and executes preCommitCallback BEFORE calling app.Commit
// The callback is useful for testing purposes to execute arbitrary code before the
// chain sdk context is cleared in .Commit().
// For example, app.EndBlock may lead to a new state, which you would like to query
// to check that it is correct. However, the sdk context is cleared after .Commit(),
// so you can query the state inside the callback.
//
// NOTE: this method may be used independently of the rest of simibc.
func EndBlock(
	c *ibctesting.TestChain,
	preCommitCallback func(),
) (*ibctmtypes.Header, []channeltypes.Packet) {
	ebRes := c.App.EndBlock(abci.RequestEndBlock{Height: c.CurrentHeader.Height})

	/*
		It is useful to call arbitrary code after ending the block but before
		committing the block because the sdk.Context is cleared after committing.
	*/

	c.App.Commit()

	c.Vals = c.NextVals

	c.NextVals = ibctesting.ApplyValSetChanges(c.T, c.Vals, ebRes.ValidatorUpdates)

	c.LastHeader = c.CurrentTMClientHeader()

	sdkEvts := ABCIToSDKEvents(ebRes.Events)
	packets := ParsePacketsFromEvents(sdkEvts)

	return c.LastHeader, packets
}

// ParsePacketsFromEvents returns all packets found in events.
func ParsePacketsFromEvents(events []sdk.Event) (packets []channeltypes.Packet) {
	for i, ev := range events {
		if ev.Type == channeltypes.EventTypeSendPacket {
			packet, err := ibctesting.ParsePacketFromEvents(events[i:])
			if err != nil {
				panic(err)
			}
			packets = append(packets, packet)
		}
	}
	return
}

// ABCIToSDKEvents converts a list of ABCI events to Cosmos SDK events.
func ABCIToSDKEvents(abciEvents []abci.Event) sdk.Events {
	var events sdk.Events
	for _, evt := range abciEvents {
		var attributes []sdk.Attribute
		for _, attr := range evt.GetAttributes() {
			attributes = append(attributes, sdk.NewAttribute(attr.Key, attr.Value))
		}

		events = events.AppendEvent(sdk.NewEvent(evt.GetType(), attributes...))
	}

	return events
}
