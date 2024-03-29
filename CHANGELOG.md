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

### Features

- (evm) [#668](https://github.com/dymensionxyz/dymension/issues/668) Integrate virtual frontier bank contract
- (delayedack) [#728](https://github.com/dymensionxyz/dymension/issues/728) create eibc order on err ack from rollapp

### Bug Fixes

- (rollapp) [#471](https://github.com/dymensionxyz/dymension/issues/471) Validate rollapp token metadata
- (hygiene) [#676](https://github.com/dymensionxyz/dymension/pull/676) lint tests
- (ibc) [#678](https://github.com/dymensionxyz/dymension/pull/678) apply a pfm patch
- (dependencies) [#677](https://github.com/dymensionxyz/dymension/pull/677) Bump cosmos ecosystem dependencies
- (rollapp) [#739](https://github.com/dymensionxyz/dymension/issues/739) Use cached context to avoid panic in finalize queue
- (vfc) [#726](https://github.com/dymensionxyz/dymension/issues/726) Remove denommetadata ibc middleware and register denoms in genesis event
- (delayedack) [#741](https://github.com/dymensionxyz/dymension/issues/741) Use must unmarshal packet and demand orders

___

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0](https://github.com/dymensionxyz/dymension/releases/tag/v0.1.0-alpha)

Initial Release!
