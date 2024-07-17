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

### State Machine Breaking

- (dependencies) [#970](https://github.com/dymensionxyz/dymension/pull/970) Bump dependencies cosmos-sdk to v0.47.12 

### Features

- (swagger) [#856](https://github.com/dymensionxyz/dymension/issues/856) Add make command `proto-swagger-gen`
- (delayedack) [#825](https://github.com/dymensionxyz/dymension/issues/825) Add query for rollapp packets using CLI

### Bug Fixes

## [v3.1.0](https://github.com/dymensionxyz/dymension/releases/tag/v3.1.0)

### Features

- (app) [#972](https://github.com/dymensionxyz/dymension/pull/972) Refactor upgrade handlers. 
- (delayedack) [#972](https://github.com/dymensionxyz/dymension/pull/972) Use pagination when deleting rollapp packets.
- (denommetadata) [#955](https://github.com/dymensionxyz/dymension/issues/955) Add IBC middleware to create denom metadata from rollapp, on IBC transfer.
- (genesisbridge) [#932](https://github.com/dymensionxyz/dymension/issues/932) Adds ibc module and ante handler to stop transfers to/from rollapp that has an incomplete genesis bridge (transfersEnabled)
- (genesisbridge) [#932](https://github.com/dymensionxyz/dymension/issues/932) Adds a new temporary ibc module to set the canonical channel id, since we no longer do that using a whitelisted addr
- (genesisbridge) [#932](https://github.com/dymensionxyz/dymension/issues/932) Adds a new ibc module to handle incoming 'genesis transfers'. It validates the special memo and registers a denom. It will not allow any regular transfers if transfers are not enabled
- (rollapp) [#932](https://github.com/dymensionxyz/dymension/issues/932) Renames is_genesis_event on the rollapp genesis state to 'transfers_enabled' this is backwards compatible
- (rollapp) [#932](https://github.com/dymensionxyz/dymension/issues/932) Removes concept of passing genesis accounts and denoms in the create rollapp message
- (rollapp) [#932](https://github.com/dymensionxyz/dymension/issues/932) Adds a transfersenabled flag to createRollapp (might be changed in future)
- (delayedack) [#932](https://github.com/dymensionxyz/dymension/issues/932) Adds the notion of skipctx, to skip it with a special sdk context value
- (code standards) [#932](https://github.com/dymensionxyz/dymension/issues/932) Adds a gerr (google error ) and derr (dymension error) packages for idiomatic error handling. (In future we will consolidate across dymint/rdk)
- (denommetadata) [#907](https://github.com/dymensionxyz/dymension/issues/907) Add IBC middleware to migrate denom metadata to rollappp, remove `CreateDenomMetadata` and `UpdateDenomMetadata` tx handlers
- (eibc) [#873](https://github.com/dymensionxyz/dymension/issues/873) Add `FulfillerAddress` to `DemandOrder` and its event
- (delayedack) [#849](https://github.com/dymensionxyz/dymension/issues/849) Add demand order filters: type, rollapp id and limit
- (delayedack) [#850](https://github.com/dymensionxyz/dymension/issues/850) Add type filter for delayedack
- (rollapp) [#829](https://github.com/dymensionxyz/dymension/issues/829) Refactor rollapp cli to be more useful
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

- (eibc,delayedack) [#942](https://github.com/dymensionxyz/dymension/issues/942) Add missing genesis validation
- (rollapp) [#317](https://github.com/dymensionxyz/research/issues/317) Prevent overflow on rollapp state update
- (code standards) [#932](https://github.com/dymensionxyz/dymension/issues/932) Dry out existing middlewares to make use of new .GetValidTransfer* functions which take care of parsing and validating the fungible packet, and querying and validating any associated rollapp and finalizations
- (code standards) [#932](https://github.com/dymensionxyz/dymension/issues/932) Removes the obsolete ValidateRollappId func and sub routines
- (code standards) [#932](https://github.com/dymensionxyz/dymension/issues/932) Simplify GetAllBlockHeightToFinalizationQueue
- (code standards) [#932](https://github.com/dymensionxyz/dymension/issues/932) Fixes naming for our 'middlewares' to make them clearly one of ibc module / ics4 wrapper / middleware
- (code standards) [#932](https://github.com/dymensionxyz/dymension/issues/932) Moves our various utils to properly namespaced packages under utils/
- (rollapp) [#839](https://github.com/dymensionxyz/dymension/issues/839) Remove rollapp deprecated fields
- (eibc) [#836](https://github.com/dymensionxyz/dymension/issues/836) Improve eibc memo error handling
- (eibc) [#830](https://github.com/dymensionxyz/dymension/issues/830) Invalid tx should return ackErr
- (eibc) [#828](https://github.com/dymensionxyz/dymension/issues/828) Wrong packet written on delayedack acknowledgement
- (delayedack) [#822](https://github.com/dymensionxyz/dymension/issues/822) Acknowledgement not written in case of ackerr
- (rollapp) [#820](https://github.com/dymensionxyz/dymension/issues/820) Invariant block-height-to-finalization-queue fix for freezing rollapp
- (delayedack) [#814](https://github.com/dymensionxyz/dymension/issues/814) Proof height ante handler doesn't gurantee uniqueness
- (fraud) [#811](https://github.com/dymensionxyz/dymension/issues/811) Refunding pending outgoing packets
- (delayedack) [#810](https://github.com/dymensionxyz/dymension/issues/810) Wrong denom metadata created for eIBC on delayedack timeout and ack
- (delayedack) [#809](https://github.com/dymensionxyz/dymension/issues/809) Delayed ack wrong channels on timeout and ack
- (rollapp) [#807](https://github.com/dymensionxyz/dymension/issues/807) Allow creating rollapp same eip155 when forking
- (delayedack) [#799](https://github.com/dymensionxyz/dymension/issues/799) Do not create eibc order on timeout/errAck if fee is not positive
- (delayedack) [#794](https://github.com/dymensionxyz/dymension/issues/794) Fix missing validation of channel id when validating rollapp packet
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
