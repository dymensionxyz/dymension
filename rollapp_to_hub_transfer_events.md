# Rollapp to Hub Transfer - Expected Events Analysis

## Overview

A rollapp to hub transfer in the Dymension protocol is an IBC transfer with custom middleware handling on the hub, including delayedAck and EIBC (Enhanced IBC) functionality. The transfer is added to a queue until fulfilled or claimed once the rollapp height is finalized.

## Event Categories

### **Expected Events (High Probability)**

#### Standard IBC Events
- **✅ `recv_packet`**: This event is **expected** as it's emitted when the hub receives the IBC packet from the rollapp
- **✅ `send_packet`**: This event is **expected** for the initial packet being sent from the rollapp to the hub
- **⚠️ `write_acknowledgement`**: This event is **conditionally expected** - it depends on whether the delayedAck middleware delays the acknowledgement or writes it immediately
- **⚠️ `acknowledge_packet`**: This event is **conditionally expected** - it will be emitted when the acknowledgement is finally processed, but may be delayed

#### Dymension Custom Events
- **✅ `delayedack`**: This event is **expected** as it's part of Dymension's custom middleware that handles the delayed acknowledgement mechanism for rollapp transfers

### **Conditionally Expected Events (Medium Probability)**

#### EIBC Events
- **🔄 `eibc`**: This event is **conditionally expected** - it depends on whether Enhanced IBC functionality is triggered for this specific transfer
- **🔄 `dymensionxyz.dymension.eibc.EventDemandOrderCreated`**: This event is **conditionally expected** - it's emitted when an EIBC demand order is created for immediate liquidity
- **🔄 `dymensionxyz.dymension.eibc.EventDemandOrderFulfilled`**: This event is **conditionally expected** - it's emitted when an EIBC demand order is fulfilled by a market maker
- **🔄 `dymensionxyz.dymension.eibc.EventDemandOrderFulfilledAuthorized`**: This event is **conditionally expected** - it's emitted when an authorized fulfillment occurs
- **🔄 `dymensionxyz.dymension.eibc.EventDemandOrderFeeUpdated`**: This event is **conditionally expected** - it's emitted when fees are updated for a demand order

### **Unlikely Events (Low Probability)**

- **❌ `timeout_packet`**: This event is **unlikely** under normal circumstances, as it would indicate the transfer failed due to timeout

## Event Flow Explanation

### Normal Flow (No EIBC)
1. **`send_packet`** - Rollapp initiates the transfer
2. **`recv_packet`** - Hub receives the packet
3. **`delayedack`** - Delayed acknowledgement middleware processes the packet
4. **`write_acknowledgement`** - Eventually written after rollapp height finalization
5. **`acknowledge_packet`** - Finally acknowledged back to rollapp

### EIBC Enhanced Flow
1. **`send_packet`** - Rollapp initiates the transfer
2. **`recv_packet`** - Hub receives the packet
3. **`delayedack`** - Delayed acknowledgement middleware processes the packet
4. **`eibc`** - EIBC middleware is triggered
5. **`dymensionxyz.dymension.eibc.EventDemandOrderCreated`** - Demand order created for immediate liquidity
6. **`dymensionxyz.dymension.eibc.EventDemandOrderFulfilled`** - Market maker fulfills the order
7. **`write_acknowledgement`** - Acknowledgement written after processing
8. **`acknowledge_packet`** - Final acknowledgement

## Key Considerations

1. **Timing**: Events may not appear immediately due to the delayed acknowledgement mechanism
2. **Order**: The order of events depends on whether EIBC is triggered and rollapp height finalization
3. **Conditional Logic**: EIBC events only appear if the transfer qualifies for enhanced IBC processing
4. **Finalization**: Some events are delayed until the rollapp height is finalized on the hub

## Conclusion

For a typical rollapp to hub transfer, you should expect:
- **Definitely**: `send_packet`, `recv_packet`, `delayedack`
- **Likely**: `write_acknowledgement`, `acknowledge_packet` (after delay)
- **Conditionally**: EIBC events (if enhanced IBC is triggered)
- **Unlikely**: `timeout_packet` (only on failure)

The specific events you'll see depend on the transfer configuration, amount, and whether EIBC enhanced processing is enabled and triggered for the particular transfer.