# Analysis of GitHub Issue #1671

## Issue Summary
**Title**: "improve error msg when unable to unbond because no state update is submitted for the transfer_proof_height"
**Status**: Open (Created: 2024-12-19, Updated: 2025-05-19)
**Assignee**: danwt

## Problem Description

When a sequencer tries to unbond, the system performs a check to ensure that rotation won't cause a fork before the genesis transfer proof height. If this check fails, the current error message is:

```
"rotation could cause fork before genesis transfer"
```

This error message is confusing because:
1. It doesn't explain what "genesis transfer" means
2. It doesn't tell the user what needs to be done to resolve the issue
3. It doesn't clarify that the issue is specifically about the transfer proof height not being submitted to the settlement layer

## Root Cause Analysis

The error occurs in `/x/sequencer/keeper/msg_server_bond.go` at line 79 in the `Unbond` function:

```go
if !k.rollappKeeper.ForkLatestAllowed(ctx, seq.RollappId) {
    return nil, gerrc.ErrFailedPrecondition.Wrap("rotation could cause fork before genesis transfer")
}
```

The `ForkLatestAllowed` function checks `ForkAllowed` which returns `false` when:
- Genesis bridge is enabled (`IsTransferEnabled() == true`)
- AND the fork would rollback past the `TransferProofHeight` (`TransferProofHeight > lastValidHeight`)

## Technical Context

1. **TransferProofHeight**: The height at which the genesis transfer proof was submitted to the settlement layer
2. **IsTransferEnabled()**: Returns `true` when `TransferProofHeight != 0`
3. **ForkAllowed()**: Ensures that forks don't violate the integrity of already-proven genesis transfers

## Current Status

The issue is still relevant. The assignee (danwt) commented on 2025-05-19: "need to check if still relevant with fork changes"

## Proposed Solution

Replace the current vague error message with a more descriptive one that explains:
- What is being checked
- Why the operation is blocked
- What conditions need to be met for the operation to succeed

## Implementation

The fix involves updating the error message in `x/sequencer/keeper/msg_server_bond.go` to provide better context about the genesis transfer proof height requirement.

## Fixed Implementation

**Location**: `x/sequencer/keeper/msg_server_bond.go:79`

**Old Error Message**:
```
"rotation could cause fork before genesis transfer"
```

**New Error Message**:
```
"cannot unbond: sequencer rotation would cause a fork that affects the genesis transfer proof height. The genesis transfer proof must be submitted to the settlement layer before unbonding can proceed"
```

## Benefits of the Fix

1. **Clarity**: The new message clearly explains what operation is being blocked (unbonding)
2. **Context**: It explains the technical reason (fork affecting genesis transfer proof height)
3. **Actionable**: It tells the user what needs to happen for the operation to succeed (genesis transfer proof submission)
4. **User-friendly**: Uses clear language that helps users understand the prerequisite conditions

## Testing

The fix preserves the existing logic while only improving the error message. The same conditions that triggered the old error will trigger the new error, ensuring no behavioral changes to the system's security mechanisms.

## Status

✅ **FIXED** - Issue #1671 has been resolved by improving the error message to provide clear, actionable information to users encountering this condition.