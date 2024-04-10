# simibc

## What is this?

A collection of utilities based on [ibc-go/testing](https://github.com/cosmos/ibc-go/tree/main/testing) which make it easier to write test scenarios involving precise orderings of

- BeginBlock, EndBlock on each IBC connected chain
- Packet delivery
- Updating the client

## Why is this useful?

It is very hard to reason about tests written using vanilla [ibc-go/testing](https://github.com/cosmos/ibc-go/tree/main/testing) because the methods included in that library have many side effects. For example, that library has a notion of global time, so calling EndBlock on one chain will influence the future block times of another chain. As another example, sending a packet from chain A to B will automatically progress the block height on chain A. These behaviors make it very hard to understand, especially if your applications have business logic in BeginBlock or EndBlock.

The utilities in simibc do not have any side effects, making it very easy to understand what is happening. It also makes it very easy to write data driven tests (like table tests, model based tests or property based tests).

## How do I use this?

Please see the function docstrings to get an idea of how you could use this package. This README is intentionally short because it is easier to maintain code and docstrings instead of markdown.
