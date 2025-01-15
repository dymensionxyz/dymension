package uevent

import (
	"encoding/json"
	"slices"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"golang.org/x/exp/maps"
)

// EmitTypedEvent takes a typed event and emits it.
// The original EmitTypedEvent from cosmos-sdk adds double quotes around the string attributes,
// which makes it difficult, if not impossible to query/subscribe to those events.
// See https://github.com/cosmos/cosmos-sdk/issues/12592 and
// https://github.com/dymensionxyz/sdk-utils/pull/5#discussion_r1724688379
func EmitTypedEvent(ctx sdk.Context, tev proto.Message) error {
	event, err := TypedEventToEvent(tev)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(event)
	return nil
}

// TypedEventToEvent takes typed event and converts to Event object
func TypedEventToEvent(tev proto.Message) (ev sdk.Event, err error) {
	evtType := proto.MessageName(tev)

	var evtJSON []byte
	evtJSON, err = codec.ProtoMarshalJSON(tev, nil)
	if err != nil {
		return
	}

	var attrMap map[string]json.RawMessage
	if err = json.Unmarshal(evtJSON, &attrMap); err != nil {
		return
	}

	// sort the keys to ensure the order is always the same
	keys := maps.Keys(attrMap)
	slices.Sort(keys)

	attrs := make([]abci.EventAttribute, 0, len(attrMap))
	for _, k := range keys {
		v := attrMap[k]
		attrs = append(attrs, abci.EventAttribute{
			Key:   k,
			Value: removeSurroundingQuotes(v),
		})
	}

	ev = sdk.Event{
		Type:       evtType,
		Attributes: attrs,
	}
	return
}

func removeSurroundingQuotes(bz []byte) string {
	const dquote = 34
	if len(bz) > 1 && bz[0] == dquote && bz[len(bz)-1] == dquote {
		return string(bz[1 : len(bz)-1])
	}
	return string(bz)
}
