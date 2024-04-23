<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState
given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## Unreleased

## [v3.1.0](https://github.com/dymensionxyz/dymension/releases/tag/v3.1.0)

### Features

- (delayedack) [#850](https://github.com/dymensionxyz/dymension/issues/850) Add type filter for delayedack
- (delayedack) [#849](https://github.com/dymensionxyz/dymension/issues/849) Add demand order filters: type, rollapp id and limit
- (delayedack) [#728](https://github.com/dymensionxyz/dymension/issues/728) Create eibc order on err ack from rollapp
- (delayedack) [#672](https://github.com/dymensionxyz/dymension/issues/672) Delayedack invariant for finalized and reverted packets
- (evm) [#668](https://github.com/dymensionxyz/dymension/issues/668) Integrate virtual frontier bank contract
- (denommetadata) [#660](https://github.com/dymensionxyz/dymension/issues/660) Add/update multiple denom metadata in same proposal
- (denommetadata) [#659](https://github.com/dymensionxyz/dymension/issues/659) Denommetadata module hook for denom creation and update
- (vfc) [#658](https://github.com/dymensionxyz/dymension/issues/658) VFC should be triggered upon new Denom Metadata registration
- (delayedack) [#655](https://github.com/dymensionxyz/dymension/pull/655) Fix proof height ante decorator
- (delayedack) [#643](https://github.com/dymensionxyz/dymension/issues/643) Validate rollapp IBC state update against current rollapp state
- (ibc) [#636](https://github.com/dymensionxyz/dymension/issues/636) Add ability to query IBC demand orders by status
- (rollapp) [#628](https://github.com/dymensionxyz/dymension/issues/628) Freeze rollapp after fraud
- (vfc) [#627](https://github.com/dymensionxyz/dymension/issues/627) Add VFC Contract for the hub
- (delayedack) [#624](https://github.com/dymensionxyz/dymension/issues/624) Discard pending rollapp ibc packets upon fraud
- (rollapp) [#617](https://github.com/dymensionxyz/dymension/issues/617) Rollapp tokens minting on hub upon rollapp channel creation
- (rollapp) [#615](https://github.com/dymensionxyz/dymension/issues/615) Gov proposal for rollapp fraud event
- (eibc) [#607](https://github.com/dymensionxyz/dymension/issues/607) Add ability to query demand order by id
- (rollapp) [#605](https://github.com/dymensionxyz/dymension/issues/605) Switch the proposing sequencer after unbonding
- (denommetadata) [#60d](https://github.com/dymensionxyz/dymension/issues/604) Create gov proposal for token metadata registration
- (eibc) [#593](https://github.com/dymensionxyz/dymension/issues/593) Release timed out eIBC funds 
- (upgrade) [#572](https://github.com/dymensionxyz/dymension/issues/572) Add upgrade handler for new and modified modules 
- (dependencies) [#525](https://github.com/dymensionxyz/dymension/pull/525) Add Ledger Nano X and S+ support
- (rollapp) [#496](https://github.com/dymensionxyz/dymension/issues/496) Sequencer bonding and Slashing MVP
- (ci) [#444](https://github.com/dymensionxyz/dymension/issues/444) Add e2e IBC Transfer Tests
- (rollapp) [#421](https://github.com/dymensionxyz/dymension/issues/421) Invariants for rollapp module
- (delayedack) [#391](https://github.com/dymensionxyz/dymension/issues/391) Added ante handler to pass proofHeight to middleware

### Bug Fixes

- (rollapp) [#769](https://github.com/dymensionxyz/dymension/issues/769) Rollapp genesis related state shouldn't be imported
- (rollapp) [#767](https://github.com/dymensionxyz/dymension/issues/767) Saved state info index as big endian
- (delayedack) [#764](https://github.com/dymensionxyz/dymension/issues/764) Fix panic on `nil` dereferences if `UpdateRollappPacketWithStatus` errors
- (account) [#762](https://github.com/dymensionxyz/dymension/issues/762) Fix wrong `bech32` prefix for `accountKeeper`
- (ante) [#761](https://github.com/dymensionxyz/dymension/issues/761) Use `UnpackAny` for `ExtensionOptionsWeb3Tx` (audit)
- (eibc) [#760](https://github.com/dymensionxyz/dymension/issues/760) Remove reverted packet to ensure `UnderlyingPacketExistInvariant`
- (sequencer) [#758](https://github.com/dymensionxyz/dymension/issues/758) Fix setting proposer to `false` when `forceUnbonding`
- (delayedack) [#757](https://github.com/dymensionxyz/dymension/issues/757) Fix ibc packet finalization, optimize ibc packet retrieval
- (ante) [#755](https://github.com/dymensionxyz/dymension/issues/755) Add missing ante handler
- (vesting) [#754](https://github.com/dymensionxyz/dymension/issues/754) Removed vesting msgs rejections
- (denommetadata) [#753](https://github.com/dymensionxyz/dymension/issues/753) Fix export genesis of denommetadata module
- (denommetadata) [#750](https://github.com/dymensionxyz/dymension/issues/750) Sync validations between different token metadata components
- (delayedack) [#741](https://github.com/dymensionxyz/dymension/issues/741) Use must unmarshal packet and demand orders
- (dependencies) [#743](https://github.com/dymensionxyz/dymension/issues/743) Update hashicorp go-getter dependency
- (rollapp) [#740](https://github.com/dymensionxyz/dymension/issues/740) Fix `genesisState` of rollapp is non-nullable struct
- (rollapp) [#739](https://github.com/dymensionxyz/dymension/issues/739) Use cached context to avoid panic in finalize queue
- (eibc,delayedack) [#728](https://github.com/dymensionxyz/dymension/issues/728) Create eIBC order upon ackError in the delayed ack middleware
- (vfc) [#726](https://github.com/dymensionxyz/dymension/issues/726) Remove denommetadata ibc middleware and register denoms in genesis event
- (rollapp) [#717](https://github.com/dymensionxyz/dymension/issues/717) Fix EIP155 keys owned by other rollapps can be overwritten
- (sequencer) [#716](https://github.com/dymensionxyz/dymension/issues/716) Sort sequencers by bond when rotating
- (sequencer) [#714](https://github.com/dymensionxyz/dymension/issues/714) Fix broken invariant with unbonding sequencers
- (rollapp) [#710](https://github.com/dymensionxyz/dymension/issues/710) Fix missing `rollappID validation on rollapp creation
- (sequencer) [#708](https://github.com/dymensionxyz/dymension/issues/708) Validate dymint pubkey when creating sequencer
- (denommetadata) [#706](https://github.com/dymensionxyz/dymension/issues/706) Remove redundant logs
- (sequencer) [#703](https://github.com/dymensionxyz/dymension/issues/703) Fix potential int overflow when creating sequencers
- (eibc,rollapp,sequencer) [#700](https://github.com/dymensionxyz/dymension/issues/700) Fix missing invariants wiring
- (rollapp) [#699](https://github.com/dymensionxyz/dymension/pull/699) Validate the IBC client on fraud proposal
- (rollapp) [#691](https://github.com/dymensionxyz/dymension/pull/691) Limit the number of permissioned addresses in MsgCreateRollapp
- (denommetadata) [#694](https://github.com/dymensionxyz/dymension/pull/694) Add token metadata on genesis event
- (rollapp) [#690](https://github.com/dymensionxyz/dymension/pull/690) Fix wrong height in state update in rollapp module invariants test
- (rollapp) [#681](https://github.com/dymensionxyz/dymension/pull/681) Accept rollapp initial state with arbitrary height
- (ibc) [#678](https://github.com/dymensionxyz/dymension/pull/678) Apply a pfm patch
- (rollapp) [#671](https://github.com/dymensionxyz/dymension/pull/671) Fix rollapp genesis token not registered as IBC denom
- (dependencies) [#677](https://github.com/dymensionxyz/dymension/pull/677) Bump cosmos ecosystem dependencies
- (hygiene) [#676](https://github.com/dymensionxyz/dymension/pull/676) Lint tests
- (rollapp) [#657](https://github.com/dymensionxyz/dymension/pull/657) Verification of broken invariant logic
- (rollapp) [#649](https://github.com/dymensionxyz/dymension/pull/649) Fix grace period finalization test
- (rollapp) [#646](https://github.com/dymensionxyz/dymension/pull/646) Fix problem with state info finalization queue
- (eibc) [#644](https://github.com/dymensionxyz/dymension/pull/644) Limit `order_id` length when submitting eIBC order to avoid block spam
- (sequencer) [#625](https://github.com/dymensionxyz/dymension/pull/625) Add events for sequencer module
- (delayedack) [#620](https://github.com/dymensionxyz/dymension/pull/620) Add missing param initialization for delayedAck
- (eibc) [#609](https://github.com/dymensionxyz/dymension/pull/609) DelayedAck panic on PFM memo
- (eibc) [#600](https://github.com/dymensionxyz/dymension/pull/600) Temporarily disable eIBC + PFM for txs initiated on rollapp
- (ibc) [#569](https://github.com/dymensionxyz/dymension/issues/569) Move e2e tests to dymension
- (ibc) [#532](https://github.com/dymensionxyz/dymension/issues/532) Delete RollappPackets after finalization/revert #532
- (rollapp) [#471](https://github.com/dymensionxyz/dymension/issues/471) Validate rollapp token metadata
- (rollapp) [#341](https://github.com/dymensionxyz/dymension/issues/341) Change finalization logic to calculate finalization from the end

___

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0](https://github.com/dymensionxyz/dymension/releases/tag/v0.1.0-alpha)

Initial Release!
