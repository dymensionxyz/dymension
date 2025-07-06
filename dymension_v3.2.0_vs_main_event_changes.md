# Dymension Event Changes Analysis: v3.2.0 vs Main Branch

## Overview

This document outlines how to identify changes in events between the Dymension v3.2.0 tag and the main branch, particularly focusing on rollapp to hub transfer events and IBC-related functionality.

## Key Areas to Examine

### 1. Event Definition Changes
Check for modifications in event definitions across these modules:

#### Core IBC Events
- **Location**: `x/ibc/` modules
- **Files to check**:
  - `types/events.go`
  - `keeper/` files with event emissions
  - Event constants and type definitions

#### DelayedAck Module
- **Location**: `x/delayedack/`
- **Files to check**:
  - `types/events.go`
  - `keeper/msg_server.go`
  - Middleware implementations
  - Event emission patterns

#### EIBC Module
- **Location**: `x/eibc/`
- **Files to check**:
  - `types/events.go`
  - `keeper/demand_order.go`
  - `keeper/msg_server.go`
  - Event type definitions:
    - `EventDemandOrderCreated`
    - `EventDemandOrderFulfilled`
    - `EventDemandOrderFulfilledAuthorized`
    - `EventDemandOrderFeeUpdated`

### 2. How to Check for Changes

#### Method 1: Direct Repository Comparison
```bash
# Clone the repository
git clone https://github.com/dymensionxyz/dymension.git
cd dymension

# Check out both versions
git fetch --all --tags

# Compare event-related files
git diff v3.2.0..main -- x/delayedack/types/events.go
git diff v3.2.0..main -- x/eibc/types/events.go
git diff v3.2.0..main -- x/ibc/
git diff v3.2.0..main -- "**/*events*.go"
```

#### Method 2: GitHub Web Interface
Visit: `https://github.com/dymensionxyz/dymension/compare/v3.2.0...main`

Filter results by:
- Files containing "events"
- Changes in `x/eibc/` directory
- Changes in `x/delayedack/` directory
- Modifications to IBC-related modules

#### Method 3: Search for Event-Related Commits
```bash
# Search commit messages for event-related changes
git log v3.2.0..main --grep="event" --oneline
git log v3.2.0..main --grep="eibc" --oneline
git log v3.2.0..main --grep="delayedack" --oneline
git log v3.2.0..main --grep="ibc" --oneline
```

### 3. Specific Event Types to Track

#### Standard IBC Events
- `send_packet`
- `recv_packet`
- `write_acknowledgement`
- `acknowledge_packet`
- `timeout_packet`

#### Dymension Custom Events
- `delayedack` events
- `eibc` events

#### EIBC Specific Events
- `dymensionxyz.dymension.eibc.EventDemandOrderCreated`
- `dymensionxyz.dymension.eibc.EventDemandOrderFulfilled`
- `dymensionxyz.dymension.eibc.EventDemandOrderFulfilledAuthorized`
- `dymensionxyz.dymension.eibc.EventDemandOrderFeeUpdated`

### 4. What Changes to Look For

#### Event Structure Changes
- New event fields added
- Event fields removed or renamed
- Changes in event data types
- Modifications to event attribute names

#### Event Emission Changes
- New events being emitted
- Events no longer being emitted
- Changes in when events are triggered
- Modifications to event emission logic

#### Event Flow Changes
- Changes in the order of events
- New conditional logic affecting event emission
- Modifications to middleware that affects events

### 5. Files That Commonly Contain Event Changes

```
x/eibc/types/events.go
x/eibc/keeper/demand_order.go
x/eibc/keeper/msg_server.go
x/delayedack/types/events.go
x/delayedack/keeper/
x/rollapp/types/events.go
x/sequencer/types/events.go
app/app.go (middleware setup)
```

### 6. Testing Event Changes

#### Check Integration Tests
- Look for changes in test files that verify events
- Check e2e tests for event validation
- Review test expectations for event sequences

#### Validate Event Schemas
- Ensure event schemas are backward compatible
- Check if new events require client updates
- Verify event parsing logic still works

### 7. Documentation and Changelog

#### Check Documentation Updates
- Look for event-related documentation changes
- Review API documentation updates
- Check if new events are documented

#### Examine Changelog
- Review `CHANGELOG.md` for event-related entries
- Look for breaking changes in events
- Check for new features that might emit events

## Expected Impact on Rollapp to Hub Transfers

### Potential Changes to Monitor
1. **New EIBC Features**: Additional events for enhanced IBC functionality
2. **DelayedAck Improvements**: Modified acknowledgment handling events
3. **Middleware Updates**: Changes to IBC middleware affecting event flow
4. **Security Enhancements**: New events for security features
5. **Performance Optimizations**: Events related to performance improvements

### Backward Compatibility Concerns
- **Event Schema Changes**: Breaking changes to existing events
- **Event Timing Changes**: Events emitted at different times
- **New Required Events**: Events that clients must now handle
- **Deprecated Events**: Events no longer emitted

## Conclusion

To identify specific changes between v3.2.0 and main:

1. **Clone the repository** and use git diff commands
2. **Focus on event-related files** in key modules
3. **Check commit history** for event-related changes
4. **Review test files** for event validation changes
5. **Examine documentation** for new event descriptions

The most critical areas for rollapp to hub transfers are the EIBC and DelayedAck modules, as these directly impact the events you listed in your original question.

## Next Steps

1. Execute the git commands above to identify specific changes
2. Review any found differences for impact on your use case
3. Test event handling with both versions if significant changes are found
4. Update client code if new events are introduced or existing events are modified