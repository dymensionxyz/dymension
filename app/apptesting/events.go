package apptesting

import (
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AssertEventEmitted asserts that ctx's event manager has emitted the given number of events
// of the given type.
func (s *KeeperTestHelper) AssertEventEmitted(ctx sdk.Context, eventTypeExpected string, numEventsExpected int) {
	allEvents := ctx.EventManager().Events()
	// filter out other events
	actualEvents := make([]sdk.Event, 0)
	for _, event := range allEvents {
		if event.Type == eventTypeExpected {
			actualEvents = append(actualEvents, event)
		}
	}
	s.Require().Equal(numEventsExpected, len(actualEvents))
}

func (s *KeeperTestHelper) FindEvent(events []sdk.Event, name string) sdk.Event {
	index := slices.IndexFunc(events, func(e sdk.Event) bool { return e.Type == name })
	if index == -1 {
		return sdk.Event{}
	}
	return events[index]
}

// FindLastEventOfType returns the last event of the given type.
func (s *KeeperTestHelper) FindLastEventOfType(events []sdk.Event, eventType string) (sdk.Event, bool) {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Type == eventType {
			return events[i], true
		}
	}
	return sdk.Event{}, false
}

func (s *KeeperTestHelper) ExtractAttributes(event sdk.Event) map[string]string {
	attrs := make(map[string]string)
	if event.Attributes == nil {
		return attrs
	}
	for _, a := range event.Attributes {
		attrs[a.Key] = a.Value
	}
	return attrs
}

func (s *KeeperTestHelper) AssertAttributes(event sdk.Event, eventAttributes []sdk.Attribute) {
	attrs := s.ExtractAttributes(event)
	for _, attr := range eventAttributes {
		s.Equal(attr.Value, attrs[attr.Key])
	}
}
