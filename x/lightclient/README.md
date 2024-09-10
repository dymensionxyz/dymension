# Light client

This module implements the 'canonical light client' concept. Each established Rollapp has an associated canonical light client, which allows safe IBC light clients to be created and operated permissionlessly.

# Operator Info

## Help! My IBC channel isn't working!

This is a help section for operators - people running rollapps against the Hub in any environment (prod, testnet, local or e2e tests).

If you cannot create a useable IBC channel between your rollapp and the Hub, it may be due to the light client for your rollapp on the Hub not being marked as canonical by the Hub state.

### An overview of the protocol

The Hub will allow the creation of new light clients by anyone. The first light client to match the state info's sent by the sequencer to the Hub is marked 'canonical', which will allow it to be used to create IBC channels which use EIBC. That means it's very important that your light client be marked canonical before it is used to create a channel, which in turn requires the state update to have arrived on the Hub from the sequencer for a height which is *at least* the Rollapp height that the light client was created from. 

To summarize, the order of steps is

1. Create light clients 
2. Wait for another state update to arrive on the Hub from the Rollapp sequencer
3. Create IBC transfer channel

The Dymension relayer supports this flow out of the box.

Moreover, it is important to create the light client for the Rollapp on the Hub with the right parameters. The correct parameters can be seen with `dymd q lightclient expected`, and relevant parameters are the trust level, trusting period, unbonding period and max clock drift. The Dymension relayer ensures these parameters have the correct values. If in doubt, compare the output of `dymd q ibc client state 07-tendermint-x` for your light client with the expected values from the Hub.

When combined, this flow implies a few relationships between parameters

```
dymint max idle time < trusting period < rollapp x/sequencers unbonding period = hub x/sequencer unbonding period
```

and additionally, before creating the channel it is wise also set `dymint max batch time` to a small value, since step (2) in the procedure above requires a state update.


### Operator checklist

- [ ] Using latest compatible relayer from https://github.com/dymensionxyz/go-relayer (main-dym) branch
- [ ] Rollapp x/sequencers unbonding period is equal to the Hub x/sequencer unbonding period
- [ ] Dymint idle time is less than trusting period
- [ ] Dymint batch time is short when creating the client and channel
- [ ] Client params equal to expected values (may need to pass `--max-clock-drift` to relayer)

### Additional tips

#### Verifying the result

Check if the light client is canonical with `dymd q lightclient light-client $ROLLAPP_CHAIN_ID`.

#### Small trusting period

Try the relayer `--time-threshold` [flag](https://github.com/cosmos/relayer/blob/main/docs/advanced_usage.md#auto-update-light-client) to make sure the light client does not expire.

