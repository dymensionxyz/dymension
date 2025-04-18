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

## [v3.2.0](https://github.com/dymensionxyz/dymension/releases/tag/v3.2.0)


### Bug Fixes

* **rollapp:** allow tokenless on `CreateRollapp` / `UpdateRollapp` with eip ([#1685](https://github.com/dymensionxyz/dymension/issues/1685)) ([abe082a](https://github.com/dymensionxyz/dymension/commit/abe082ab2cdb0363f1075356fc4f18f506065d74))
* **upgrade:** moved old params load to common flow instead of upgrade specific ([#1687](https://github.com/dymensionxyz/dymension/issues/1687)) ([b77a9e1](https://github.com/dymensionxyz/dymension/commit/b77a9e11a4e890135274effd48e318c17e231e25))
* linter ([#1679](https://github.com/dymensionxyz/dymension/issues/1679)) ([95ab385](https://github.com/dymensionxyz/dymension/commit/95ab38589fa69e04ef8116feee263c914bf865d0))
* **migration:** fix setting canonical light clients and gauged denom-metadata ([#1680](https://github.com/dymensionxyz/dymension/issues/1680)) ([a27cb91](https://github.com/dymensionxyz/dymension/commit/a27cb91e5c245025f2da89c6bcddebb31d092935))
* ante handler reject messages recursively ([#1392](https://github.com/dymensionxyz/dymension/issues/1392)) ([5111f26](https://github.com/dymensionxyz/dymension/commit/5111f261887f334e3c3cf7b4c981f3fce92dfb2f))
* **ante:** FeeMarket minGasPrice enforcement ([#1602](https://github.com/dymensionxyz/dymension/issues/1602)) ([b98c7cf](https://github.com/dymensionxyz/dymension/commit/b98c7cfca9ec7b17a589fcb394b968558e966ac8))
* **app:** fix initialization of `transferKeeper` for `x/rollapp` ([#1495](https://github.com/dymensionxyz/dymension/issues/1495)) ([b8a69b8](https://github.com/dymensionxyz/dymension/commit/b8a69b8e5d2a7268185d7b1402b71a0b8d3a7df9))
* **app:** fix missing tx encoder ([#1521](https://github.com/dymensionxyz/dymension/issues/1521)) ([792bec7](https://github.com/dymensionxyz/dymension/commit/792bec773749d1b4851f9652d5625b1e0c15cf13))
* **app:** use the sdk-utils EmitTypeEvent function that doesn't add double quotes to string attributes ([#1123](https://github.com/dymensionxyz/dymension/issues/1123)) ([95c7b5f](https://github.com/dymensionxyz/dymension/commit/95c7b5f6993ceb5d290ec750088282e7412e77e5))
* **bridging_fee:** adding cacheCtx for charging fees ([#1563](https://github.com/dymensionxyz/dymension/issues/1563)) ([65cd290](https://github.com/dymensionxyz/dymension/commit/65cd29029b5af280dccda3eba408550e763ecdcf))
* **bridging_fee:** getting correct denom for incoming transfers ([#1567](https://github.com/dymensionxyz/dymension/issues/1567)) ([31fd2d5](https://github.com/dymensionxyz/dymension/commit/31fd2d565984a824abb6ec1fa30e75435f31969f))
* check for duplicates in genesis info accounts ([#1472](https://github.com/dymensionxyz/dymension/issues/1472)) ([fb0ffb7](https://github.com/dymensionxyz/dymension/commit/fb0ffb7525a086960c3910538bfa535583cfd02f))
* **cli:** use proper context for CLI query client ([#1183](https://github.com/dymensionxyz/dymension/issues/1183)) ([d2f912c](https://github.com/dymensionxyz/dymension/commit/d2f912c33c70ce2c7888358d1c53d52c5d696c88))
* **code standards:** bring over more linters from dymint repo ([#884](https://github.com/dymensionxyz/dymension/issues/884)) ([98faac3](https://github.com/dymensionxyz/dymension/commit/98faac38c13ba06e10466c79a24ad42b739ee66d))
* **code standards:** deletes unused all invariants func ([#894](https://github.com/dymensionxyz/dymension/issues/894)) ([26a8163](https://github.com/dymensionxyz/dymension/commit/26a81635ef53b8f6080882e13fb29a5ada7ce5ce))
* **code standards:** remove calls to suite.SetupTest ([#1052](https://github.com/dymensionxyz/dymension/issues/1052)) ([d2baf5b](https://github.com/dymensionxyz/dymension/commit/d2baf5b8cd6da58936332ead94e5cf0de7259481))
* **code standards:** rename fullfilled -> fulfilled ([#895](https://github.com/dymensionxyz/dymension/issues/895)) ([551a945](https://github.com/dymensionxyz/dymension/commit/551a945337e4864d3f2dafd25148dd915e252072))
* **code standards:** use dymensionxyz/gerrc instead of gerr ([#959](https://github.com/dymensionxyz/dymension/issues/959)) ([a262275](https://github.com/dymensionxyz/dymension/commit/a2622758bb0486d352f312f32e739afe333ee149))
* **code standards:** use https://github.com/dymensionxyz/sdk-utils ([#968](https://github.com/dymensionxyz/dymension/issues/968)) ([d15973b](https://github.com/dymensionxyz/dymension/commit/d15973b715cd87f672c29910b48fc6bfe2a72a02))
* **crisis:** Fix crisis module wiring ([#1424](https://github.com/dymensionxyz/dymension/issues/1424)) ([ea79caf](https://github.com/dymensionxyz/dymension/commit/ea79caf41821e2d23693a491074071f864fe40b4))
* **delayedack:**  use correct port, chan in packet uid ([#963](https://github.com/dymensionxyz/dymension/issues/963)) ([48824ed](https://github.com/dymensionxyz/dymension/commit/48824ed62092782a4e292a08b018588f9907e00e))
* **delayedack:** acknowledgement not written in case of ackerr ([#824](https://github.com/dymensionxyz/dymension/issues/824)) ([d0a5df1](https://github.com/dymensionxyz/dymension/commit/d0a5df18b4f96554f6ce2e1f1fd3181496e2e16d))
* **delayedack:** don't error when deleting packets to prevent endless loop ([#1483](https://github.com/dymensionxyz/dymension/issues/1483)) ([5fc6dd0](https://github.com/dymensionxyz/dymension/commit/5fc6dd0f31ac1f8bda8cc3772bfaf3f8ed70c7c6))
* **delayedack:** handle finalization errors per rollapp ([#966](https://github.com/dymensionxyz/dymension/issues/966)) ([d8de3c6](https://github.com/dymensionxyz/dymension/commit/d8de3c6c29b79f1c77d00aa3ba71ae8e5ff7b602))
* **delayedack:** malformed trackingPacketKey by ensuring base64 encoding/decoding during Genesis export/import ([#1668](https://github.com/dymensionxyz/dymension/issues/1668)) ([12bfcad](https://github.com/dymensionxyz/dymension/commit/12bfcad722ac0dd5ad39ba0e70559ab34ead1924))
* **denom meta:** use correct channels in middlewares arguments ([#1191](https://github.com/dymensionxyz/dymension/issues/1191)) ([b50f568](https://github.com/dymensionxyz/dymension/commit/b50f5687e82d65e0696939b96ccf3abb820bdaa3))
* **denommetadata:** correct base denom format ([#1220](https://github.com/dymensionxyz/dymension/issues/1220)) ([6de2c58](https://github.com/dymensionxyz/dymension/commit/6de2c58ef5c2872e8ec06b97313a720aaa459ba4))
* **denommetadata:** emit events on denom metadata creating and updating ([#1165](https://github.com/dymensionxyz/dymension/issues/1165)) ([ac802eb](https://github.com/dymensionxyz/dymension/commit/ac802ebe23d9b254e9b277e0016e383fa12c9819))
* **deps:** updated osmosis dep to fix taker fee event ([#1616](https://github.com/dymensionxyz/dymension/issues/1616)) ([a401959](https://github.com/dymensionxyz/dymension/commit/a4019592dbe05687cbeb512ad6d803d7c21691d2))
* **deps:** upgraded to ethermint which doesn't emit double events on transfer ([#1534](https://github.com/dymensionxyz/dymension/issues/1534)) ([06ae845](https://github.com/dymensionxyz/dymension/commit/06ae8452b83deabc177c5004948ed741f80a3c63))
* **doc:** improve docstring on proto rollapp LivenessEventHeight ([#1298](https://github.com/dymensionxyz/dymension/issues/1298)) ([2717c18](https://github.com/dymensionxyz/dymension/commit/2717c1847c0375e3c85c2e8c40f42c3ec24cafcd))
* **docs:** adds docstrings for demand order proto ([#896](https://github.com/dymensionxyz/dymension/issues/896)) ([f6630da](https://github.com/dymensionxyz/dymension/commit/f6630da1de694780afdb9d50a94ca4dbdb7e65c1))
* **dymns:** replace incorrect usage of `ErrInternal` ([#1628](https://github.com/dymensionxyz/dymension/issues/1628)) ([0a5b253](https://github.com/dymensionxyz/dymension/commit/0a5b25325b9cf1961455607434904fafbd87eb18))
* **eibc,delaydack:** add validation to eibc and delayedack genesis state ([#967](https://github.com/dymensionxyz/dymension/issues/967)) ([fcd6081](https://github.com/dymensionxyz/dymension/commit/fcd6081202048bd10ff80cf89f1529fd929394c0))
* **eibc:** accept wrong fee ([#1638](https://github.com/dymensionxyz/dymension/issues/1638)) ([b6d70ff](https://github.com/dymensionxyz/dymension/commit/b6d70ffbbc5eca70b3151daf095887f2482fd8b7))
* **eibc:** add `packet_type` and `is_fulfilled` to eibc events ([#1148](https://github.com/dymensionxyz/dymension/issues/1148)) ([8382b00](https://github.com/dymensionxyz/dymension/commit/8382b00b91d764cc008f4fbb6dfea670f4306d79))
* **eibc:** add event for order deleted ([#1664](https://github.com/dymensionxyz/dymension/issues/1664)) ([c710052](https://github.com/dymensionxyz/dymension/commit/c710052d280f9e516603f745b6c4903f53f0708f))
* **eibc:** add test cases for authorized order fulfillment ([#1347](https://github.com/dymensionxyz/dymension/issues/1347)) ([ae279de](https://github.com/dymensionxyz/dymension/commit/ae279de08fb39a611f0566fb4b2ab4853d68dfd9))
* **eIBC:** bridging_fee taken from original recipient and not from fufliller  ([#918](https://github.com/dymensionxyz/dymension/issues/918)) ([7654979](https://github.com/dymensionxyz/dymension/commit/765497923ae16b85a0acc55108a4e8a135238423))
* **eibc:** fix `update-demand-order` CLI ([#992](https://github.com/dymensionxyz/dymension/issues/992)) ([e538ff7](https://github.com/dymensionxyz/dymension/commit/e538ff7cb0f003d2158125ce79ade091f7c26b0f))
* **eibc:** fix demand order fulfill authorized fee check ([#1522](https://github.com/dymensionxyz/dymension/issues/1522)) ([f9d95b0](https://github.com/dymensionxyz/dymension/commit/f9d95b0554ad7e5654aa915534aeab8a863a4ccf))
* **eibc:** have more granularity when emitting eibc order related events ([#1112](https://github.com/dymensionxyz/dymension/issues/1112)) ([1ec00c0](https://github.com/dymensionxyz/dymension/commit/1ec00c066d20dd026497d609b07ee52d7fa88a3f))
* **eibc:** improve eibc memo error handling ([#838](https://github.com/dymensionxyz/dymension/issues/838)) ([9f7fd4e](https://github.com/dymensionxyz/dymension/commit/9f7fd4efcd529318bfaa9a70dc2ef417dbe6d975))
* **eibc:** prevent "finalizable" order due to be fulfilled ([#1435](https://github.com/dymensionxyz/dymension/issues/1435)) ([64b0124](https://github.com/dymensionxyz/dymension/commit/64b0124061180a5a3b770f3b90e78d14894edab6))
* **eibc:** wrong packet written on delayedack  acknowledgment  ([#834](https://github.com/dymensionxyz/dymension/issues/834)) ([4613066](https://github.com/dymensionxyz/dymension/commit/461306673274fad4c729d551c8fcad36bf8122ec))
* **fees:** updated default fees for IRO plan creation and bonding to gauge ([#1583](https://github.com/dymensionxyz/dymension/issues/1583)) ([ef19155](https://github.com/dymensionxyz/dymension/commit/ef191558e879daaec05acb961839c50f0e4c76ee))
* **fork:** make fraud proposal revision into block identifier ([#1478](https://github.com/dymensionxyz/dymension/issues/1478)) ([3e66754](https://github.com/dymensionxyz/dymension/commit/3e6675487df523ac0e62935cc04f1aa0f92f9d0f))
* **genesis bridge:** better UX for chains without genesis transfers ([#961](https://github.com/dymensionxyz/dymension/issues/961)) ([8a232ee](https://github.com/dymensionxyz/dymension/commit/8a232ee4d17c3bb74f65f882391650fd9e01d55c))
* **genesis bridge:** get rollapp by light client ID rather than chain ID in transfer enabled check ([#1339](https://github.com/dymensionxyz/dymension/issues/1339)) ([c9dae97](https://github.com/dymensionxyz/dymension/commit/c9dae97f3641a68cc4886e21240d009b7e8d0cf8))
* **group:** increase group metadata maximum length ([#1606](https://github.com/dymensionxyz/dymension/issues/1606)) ([8441465](https://github.com/dymensionxyz/dymension/commit/8441465010f76df8239fff8621457c5a83fbbff6))
* **ibc:** move the genesis bridge transfer blocker from ante to ics4wrapper ([#1561](https://github.com/dymensionxyz/dymension/issues/1561)) ([6bd4748](https://github.com/dymensionxyz/dymension/commit/6bd4748ea4a06010ce485c2144d2b7982fa4c211))
* **ibc:** remove denom registration fees. validate OnRecvPacket initially ([#1660](https://github.com/dymensionxyz/dymension/issues/1660)) ([b544906](https://github.com/dymensionxyz/dymension/commit/b5449068e9ebf5d9671b18a405c10e7c93f43efa))
* **ibc:** wrap IBC error acknowledgement with an error event ([#1195](https://github.com/dymensionxyz/dymension/issues/1195)) ([206984a](https://github.com/dymensionxyz/dymension/commit/206984aed271acbe4cf4aff66d20fdb5c59e1773))
* **invariant:** no longer have cons state must exist for latest height invariant ([#1633](https://github.com/dymensionxyz/dymension/issues/1633)) ([0dbcf79](https://github.com/dymensionxyz/dymension/commit/0dbcf791124e13edf4cf12b361735ad5ee8cd500))
* IRO claimed amount accounting ([#1620](https://github.com/dymensionxyz/dymension/issues/1620)) ([b44e6bd](https://github.com/dymensionxyz/dymension/commit/b44e6bde45a894d7395efa20420f323f8f71b78a))
* **iro, invariant:** remove non invariant ([#1651](https://github.com/dymensionxyz/dymension/issues/1651)) ([f756f91](https://github.com/dymensionxyz/dymension/commit/f756f91c950b638c3fe02b05b0728771b40d9304))
* **IRO:** added IRO creation event. updated settle event. fixed trade events ([#1316](https://github.com/dymensionxyz/dymension/issues/1316)) ([03ee4bd](https://github.com/dymensionxyz/dymension/commit/03ee4bd730a4a7d13ca16ab2c724e6c9b020f4b6))
* **iro:** correct IRO claim denom ([#1468](https://github.com/dymensionxyz/dymension/issues/1468)) ([d13488c](https://github.com/dymensionxyz/dymension/commit/d13488c8d53e41ea64b878e07c268e990458abda))
* **iro:** creating VFC contracts for IRO tokens ([#1314](https://github.com/dymensionxyz/dymension/issues/1314)) ([7119731](https://github.com/dymensionxyz/dymension/commit/711973139702d025c988f4fac9fd82d406a59d29))
* **IRO:** fixed UT to handle duration for IRO plan ([#1290](https://github.com/dymensionxyz/dymension/issues/1290)) ([ada3fa2](https://github.com/dymensionxyz/dymension/commit/ada3fa2314783a61e9a3da5e1afe20e3aea690c5))
* **iro:** get tradable plans filter ([#1546](https://github.com/dymensionxyz/dymension/issues/1546)) ([f3252be](https://github.com/dymensionxyz/dymension/commit/f3252befd6703be88f902471b8b19f3bead35b24))
* **IRO:** invariant only check sufficiency of IRO tokens after settling ([#1571](https://github.com/dymensionxyz/dymension/issues/1571)) ([5e0a7c7](https://github.com/dymensionxyz/dymension/commit/5e0a7c7ac3a2a291d0d308aeaa62f9269c3e3388))
* **iro:** last plan id invariant fix ([#1568](https://github.com/dymensionxyz/dymension/issues/1568)) ([2c31dd3](https://github.com/dymensionxyz/dymension/commit/2c31dd3858a28afcd1967fd1244a3dc3d7480210))
* **iro:** query for unsettled plans ([#1665](https://github.com/dymensionxyz/dymension/issues/1665)) ([68fd99d](https://github.com/dymensionxyz/dymension/commit/68fd99db7bc2f344db5a5738bf329b7e97ef8737))
* **IRO:** registered missing BuyExactSpend codec register ([#1464](https://github.com/dymensionxyz/dymension/issues/1464)) ([7c2a71f](https://github.com/dymensionxyz/dymension/commit/7c2a71f448790c6541f7f362537034a861b628a8))
* **iro:** removed ibc filter for VFC creation ([#1302](https://github.com/dymensionxyz/dymension/issues/1302)) ([d1125ef](https://github.com/dymensionxyz/dymension/commit/d1125ef233b76e30c7451ac079396eccf6f88a91))
* **iro:** rename IRO token base and display ([#1393](https://github.com/dymensionxyz/dymension/issues/1393)) ([6dcb29a](https://github.com/dymensionxyz/dymension/commit/6dcb29a68f821ef5e3b7a02c3b362011b8caf03a))
* **IRO:** updated IRO's metadata name ([#1508](https://github.com/dymensionxyz/dymension/issues/1508)) ([f8d71e7](https://github.com/dymensionxyz/dymension/commit/f8d71e7239288dbd62b9075c3a9b6bdec9064e9f))
* **iro:** use sdk-utils typed events in IRO ([#1353](https://github.com/dymensionxyz/dymension/issues/1353)) ([1aba71a](https://github.com/dymensionxyz/dymension/commit/1aba71afbee9ffa672c63c4051669352f6ad21b7))
* **light client:** disable misbehavior update for canon client ([#1212](https://github.com/dymensionxyz/dymension/issues/1212)) ([09f56ab](https://github.com/dymensionxyz/dymension/commit/09f56abb2ec9a879cc3919b04939d09eff91638b))
* **light client:** use trust derived by unbonding ([#1210](https://github.com/dymensionxyz/dymension/issues/1210)) ([7326f78](https://github.com/dymensionxyz/dymension/commit/7326f78ce26685ba8ea0128d315cf2e0a1e5cd9c))
* **lightclient:** don't expect signer ([#1463](https://github.com/dymensionxyz/dymension/issues/1463)) ([1ee6efa](https://github.com/dymensionxyz/dymension/commit/1ee6efa7e9c2e9362487b06993ab7ca0997b1e74))
* **lightclient:** make sure to prune signers when setting canonical client for first time ([#1399](https://github.com/dymensionxyz/dymension/issues/1399)) ([870510f](https://github.com/dymensionxyz/dymension/commit/870510fc68b7a102288a72a718a4bd7ac7e354de))
* **lightclient:** prune signers below on setting canonical ([#1398](https://github.com/dymensionxyz/dymension/issues/1398)) ([9cd39de](https://github.com/dymensionxyz/dymension/commit/9cd39de6eb3dca7d9ff78c742d6ba1fda95a81fb))
* **lightclient:** validate state info and timestamp eagerly, prune last valid height signer, removes buggy hard fork in progress state ([#1594](https://github.com/dymensionxyz/dymension/issues/1594)) ([29e5c75](https://github.com/dymensionxyz/dymension/commit/29e5c75735d6f36a30980cd43427aaa6e5f5e2a1))
* **liveness slash:** do not schedule for next block ([#1284](https://github.com/dymensionxyz/dymension/issues/1284)) ([90a8791](https://github.com/dymensionxyz/dymension/commit/90a8791b0a033734b08dd61e89e5d928422bc0f9))
* **mempool:** changed to use no-op app mempool by default to fix evm txs signer ([#1274](https://github.com/dymensionxyz/dymension/issues/1274)) ([a028ec2](https://github.com/dymensionxyz/dymension/commit/a028ec238c28e3531e9fca1fa2ade3afbd1ead3f))
* **migration:** add migration for correcting gamm pools denom metadata ([#1666](https://github.com/dymensionxyz/dymension/issues/1666)) ([f9d073a](https://github.com/dymensionxyz/dymension/commit/f9d073a584c14bf5134451b125ea12241ad12d0a))
* **migration:** adds module account perms changes ([#1525](https://github.com/dymensionxyz/dymension/issues/1525)) ([64fcbad](https://github.com/dymensionxyz/dymension/commit/64fcbad4d58e3a81e47711b69523c8a674f15006))
* **migration:** changed default incentive epoch identifier to be 1 day ([#1641](https://github.com/dymensionxyz/dymension/issues/1641)) ([16b96f8](https://github.com/dymensionxyz/dymension/commit/16b96f86b82fb844e7802ea489c4014dff967ca5))
* **migration:** changed default values for incentive creation to protect against dos ([#1653](https://github.com/dymensionxyz/dymension/issues/1653)) ([887a547](https://github.com/dymensionxyz/dymension/commit/887a5472c3ed40e221d27a1abb401d26a4916a8f))
* **migration:** properly migrates params ([#1459](https://github.com/dymensionxyz/dymension/issues/1459)) ([c7a8a2c](https://github.com/dymensionxyz/dymension/commit/c7a8a2cf6a505d4a61f0611cfae893d518825362))
* refactor usage of err.Error() ([#1286](https://github.com/dymensionxyz/dymension/issues/1286)) ([8d5fe8d](https://github.com/dymensionxyz/dymension/commit/8d5fe8d0e06ec4bbaa41f4a152f543ce20ac5815))
* **rollapp:**  fix packet lookup for non-rollapp chain-id ([#1243](https://github.com/dymensionxyz/dymension/issues/1243)) ([61d5018](https://github.com/dymensionxyz/dymension/commit/61d5018fc28202e3381d711281203da13d2fdfbd))
* **rollapp liveness:** rename last state update height to liveness countdown start height ([#1512](https://github.com/dymensionxyz/dymension/issues/1512)) ([64e217b](https://github.com/dymensionxyz/dymension/commit/64e217b1346c4ad36c2b2d40e7a801574cbc93e5))
* **rollapp,sequencer:** rotation and liveness slash - hook onto all real sequencer changes ([#1477](https://github.com/dymensionxyz/dymension/issues/1477)) ([1d7dad7](https://github.com/dymensionxyz/dymension/commit/1d7dad75e425eb90fcb7df1ba3a8d877b9e82120))
* **rollapp:** add `explorer_url` to rollapp metadata ([#1140](https://github.com/dymensionxyz/dymension/issues/1140)) ([ada0b90](https://github.com/dymensionxyz/dymension/commit/ada0b90ec89c3bd4d2c0df94bd8264a2e479b820))
* **rollapp:** add `fee_base_denom` and `native_base_denom` to rollapp metadata ([#1146](https://github.com/dymensionxyz/dymension/issues/1146)) ([3d67c2d](https://github.com/dymensionxyz/dymension/commit/3d67c2de6f0df92fbfa7d8596b91a60fdd211b85))
* **rollapp:** add `token_symbol` to rollapp metadata ([#1118](https://github.com/dymensionxyz/dymension/issues/1118)) ([ff4a34b](https://github.com/dymensionxyz/dymension/commit/ff4a34b5cd16fc7a5509850976487bb33aa8fff1))
* **rollapp:** add latest height and latest finalized height to rollapp summary ([#1407](https://github.com/dymensionxyz/dymension/issues/1407)) ([b49d4f5](https://github.com/dymensionxyz/dymension/commit/b49d4f5ea55134173e9402143b5cab9dc74fe8c6))
* **rollapp:** add sequence id to Rollapp App ([#1216](https://github.com/dymensionxyz/dymension/issues/1216)) ([ae280ca](https://github.com/dymensionxyz/dymension/commit/ae280ca9ba83226cb28e9b24f76ad03ac483a059))
* **rollapp:** allow empty values when registering rollapp ([#1200](https://github.com/dymensionxyz/dymension/issues/1200)) ([78c49cc](https://github.com/dymensionxyz/dymension/commit/78c49ccb5c7cb6124ad688bedc9eeb6cd9980894))
* **rollapp:** allow genesis sum and bech32 in cli create ([#1099](https://github.com/dymensionxyz/dymension/issues/1099)) ([b8d9823](https://github.com/dymensionxyz/dymension/commit/b8d98231ac3c466ed021df2c3d684058343dd71a))
* **rollapp:** bring back num blocks in state info ([#1276](https://github.com/dymensionxyz/dymension/issues/1276)) ([96d7ac7](https://github.com/dymensionxyz/dymension/commit/96d7ac7a988c973b35fe10e521563fd3620a68b0))
* **rollapp:** change `logo_data_uri` to `logo_url` ([#1133](https://github.com/dymensionxyz/dymension/issues/1133)) ([d1218ca](https://github.com/dymensionxyz/dymension/commit/d1218ca45615e0829b50135ab49b15d1c1051ab1))
* **rollapp:** DRS atomic iteration, DRS import/export genesis ([#1404](https://github.com/dymensionxyz/dymension/issues/1404)) ([b4e5688](https://github.com/dymensionxyz/dymension/commit/b4e5688249ede6a9b53a85e73dbe3392483f95f0))
* **rollapp:** Enforce revision 1 on create rollapps ([#1352](https://github.com/dymensionxyz/dymension/issues/1352)) ([600e3c8](https://github.com/dymensionxyz/dymension/commit/600e3c81cd3a9aeb123e802e1652fa2ae532caca))
* **rollapp:** field image and app_creation_cost renaming ([#1177](https://github.com/dymensionxyz/dymension/issues/1177)) ([1df1e6d](https://github.com/dymensionxyz/dymension/commit/1df1e6d5566bf4d4036beb82ac75dcaca309cc2a))
* **rollapp:** fix expensive iteration when setting canonical light client ([#1575](https://github.com/dymensionxyz/dymension/issues/1575)) ([e50f6de](https://github.com/dymensionxyz/dymension/commit/e50f6deaa6dd9de8a5a282b7fe9591b2395a494e))
* **rollapp:** fix store issue with hooks ([#1489](https://github.com/dymensionxyz/dymension/issues/1489)) ([c0774fc](https://github.com/dymensionxyz/dymension/commit/c0774fc2699290f1e8500b806c9839d14f5ef7e2))
* **rollapp:** fix validation of empty genesis info ([#1273](https://github.com/dymensionxyz/dymension/issues/1273)) ([d85ca7d](https://github.com/dymensionxyz/dymension/commit/d85ca7dddbf42f99e13708b5a902a9ab3645354a))
* **rollapp:** fixed proto contract field name mismatch ([#1042](https://github.com/dymensionxyz/dymension/issues/1042)) ([babc49b](https://github.com/dymensionxyz/dymension/commit/babc49b3e041dd0dda5cfa7d2e9961f9a0bf515a))
* **rollapp:** logo size now 40kib ([#1092](https://github.com/dymensionxyz/dymension/issues/1092)) ([a1395f1](https://github.com/dymensionxyz/dymension/commit/a1395f118339b509e7e94d30421e29e02a41a195))
* **rollapp:** make genesis accounts in genesis info nullable ([#1440](https://github.com/dymensionxyz/dymension/issues/1440)) ([d3f3c65](https://github.com/dymensionxyz/dymension/commit/d3f3c651595083ac877039442681998c4dc1f5e9))
* **rollapp:** make genesis info nullable ([#1263](https://github.com/dymensionxyz/dymension/issues/1263)) ([68f4a75](https://github.com/dymensionxyz/dymension/commit/68f4a75657b19c40efbdf6905738a5221f1b0268))
* **rollapp:** metadata URL validation ([#1667](https://github.com/dymensionxyz/dymension/issues/1667)) ([3395842](https://github.com/dymensionxyz/dymension/commit/3395842542396c564fc50265bc9b0c607619d515))
* **rollapp:** pass client keeper to ra keeper ([#973](https://github.com/dymensionxyz/dymension/issues/973)) ([0f84b52](https://github.com/dymensionxyz/dymension/commit/0f84b52c85e2fa31a5be7421b76fd89d33f19838))
* **rollapp:** permit initial sequencer to be updated to wildcard ([#1175](https://github.com/dymensionxyz/dymension/issues/1175)) ([71d2231](https://github.com/dymensionxyz/dymension/commit/71d2231404a62f159b1253d0e2893faaf48740e7))
* **rollapp:** prevent overflow on rollapp state update ([#960](https://github.com/dymensionxyz/dymension/issues/960)) ([36430d2](https://github.com/dymensionxyz/dymension/commit/36430d210fbbd5ee8f33b090d3e66e5892d30145))
* **rollapp:** properly disambiguate empty/nil genesis accounts ([#1560](https://github.com/dymensionxyz/dymension/issues/1560)) ([5c1c120](https://github.com/dymensionxyz/dymension/commit/5c1c120d9303db71000dcdb43e3d101ea987e3e2))
* **rollapp:** queries now return full rollapp info ([#1096](https://github.com/dymensionxyz/dymension/issues/1096)) ([c9c033e](https://github.com/dymensionxyz/dymension/commit/c9c033e1e0f07867bcc41bd7f6b4647bc6cb96c8))
* **rollapp:** remove `TokenLogoDataUri` from rollapp ([#1124](https://github.com/dymensionxyz/dymension/issues/1124)) ([7883280](https://github.com/dymensionxyz/dymension/commit/788328035dcc6cceec51b30b6787bcf74c4794f7))
* **rollapp:** rename rollapp field sealed to launched ([#1186](https://github.com/dymensionxyz/dymension/issues/1186)) ([686dd4a](https://github.com/dymensionxyz/dymension/commit/686dd4a4652767f8094fcd24617fd9fc30240cfe))
* **rollapp:** restore the ability to import rollapp state ([#883](https://github.com/dymensionxyz/dymension/issues/883)) ([73a8a1d](https://github.com/dymensionxyz/dymension/commit/73a8a1d8da8a48f236ce2503eeac6cd8b13ff4e4))
* **rollapp:** return not found error when state info not found ([d081ce7](https://github.com/dymensionxyz/dymension/commit/d081ce7aab3f2c58ab28f3427e7a634789eac3d7))
* **rollapp:** revert state info created_at removal ([#1321](https://github.com/dymensionxyz/dymension/issues/1321)) ([ee76270](https://github.com/dymensionxyz/dymension/commit/ee76270e1ca3393c07ee66fff6b273bd9b9c6ac4))
* **rollapp:** revert state update pruning feature ([#1312](https://github.com/dymensionxyz/dymension/issues/1312)) ([2cd612a](https://github.com/dymensionxyz/dymension/commit/2cd612aaa6c21b473dbbb7dca9fd03b5aaae6583))
* **rollapp:** track registered rollapp denoms in separate store ([#1344](https://github.com/dymensionxyz/dymension/issues/1344)) ([f80e912](https://github.com/dymensionxyz/dymension/commit/f80e9120cb53e1529220ba0465003bd00bac4c6a))
* **rollapp:** trigger EventTypeStatusChange instead of EventTypeStateUpdate on state update finalization ([#913](https://github.com/dymensionxyz/dymension/issues/913)) ([3e7dfc9](https://github.com/dymensionxyz/dymension/commit/3e7dfc9981ef668dfecb0b3ce20859c7fc7b1377))
* **rollapp:** update rollapp fields add genesis_info ([#1174](https://github.com/dymensionxyz/dymension/issues/1174)) ([1122be6](https://github.com/dymensionxyz/dymension/commit/1122be6a2414b0fc15c0dab2dc9ad5d22914aa68))
* **rollapp:** validate order field when validating update app ([#1351](https://github.com/dymensionxyz/dymension/issues/1351)) ([68c9090](https://github.com/dymensionxyz/dymension/commit/68c9090d371bb8ac80ea9873e780af935f8a5b86))
* **rollapp:** validate owner address ([#1277](https://github.com/dymensionxyz/dymension/issues/1277)) ([e00ef66](https://github.com/dymensionxyz/dymension/commit/e00ef66539ea4ef131aab40b510f75233f15eecd))
* **rollapp:** validating rollapp revision on fraud proposal ([#1520](https://github.com/dymensionxyz/dymension/issues/1520)) ([bc0532f](https://github.com/dymensionxyz/dymension/commit/bc0532f181a5c6aade7ccd40246d030ef821b354))
* **rotation:** complete rotation on `afterStateUpdate` hook ([#1493](https://github.com/dymensionxyz/dymension/issues/1493)) ([f611c49](https://github.com/dymensionxyz/dymension/commit/f611c494d5e98a17e1fbdd179a42240d5a0c0cd4))
* **sequencer, lightclient:** do not prevent unbond when canonical client not found ([#1630](https://github.com/dymensionxyz/dymension/issues/1630)) ([00799a0](https://github.com/dymensionxyz/dymension/commit/00799a0aeaaf3bf70552e03d1cf1ed2a06c8eb2f))
* **sequencer:** add sequencer decrease bond cli command ([#1196](https://github.com/dymensionxyz/dymension/issues/1196)) ([004462d](https://github.com/dymensionxyz/dymension/commit/004462d5f45ba5ba23209c36fca54bc49a3765d8))
* **sequencer:** allow seq creation when awaiting proposer last block ([#1501](https://github.com/dymensionxyz/dymension/issues/1501)) ([1951cbd](https://github.com/dymensionxyz/dymension/commit/1951cbd3c224ab2e082caee59384822d0a94444d))
* **sequencer:** bond reduction cleanup ([#1157](https://github.com/dymensionxyz/dymension/issues/1157)) ([b6020cb](https://github.com/dymensionxyz/dymension/commit/b6020cb9a34054932fd8ffa671c3bdc1ca223907))
* **sequencer:** change metadata gas_price Int to string ([#1670](https://github.com/dymensionxyz/dymension/issues/1670)) ([9f14dac](https://github.com/dymensionxyz/dymension/commit/9f14dac8b7926f1c55e699b1fe404ed4909ade7d))
* **sequencer:** do not allow setting sentinel object ([#1619](https://github.com/dymensionxyz/dymension/issues/1619)) ([240c85e](https://github.com/dymensionxyz/dymension/commit/240c85efc3edf808af7bea6944114d658ed3ad11))
* **sequencer:** do not return sentinel values ([#1639](https://github.com/dymensionxyz/dymension/issues/1639)) ([caca30c](https://github.com/dymensionxyz/dymension/commit/caca30c37030b92fdefe7b1bdc67f47414a0dcb6))
* **sequencer:** ensure sequencer cannot propose twice | graceful handling of non proposer in notice queue ([#1513](https://github.com/dymensionxyz/dymension/issues/1513)) ([bd6e324](https://github.com/dymensionxyz/dymension/commit/bd6e32423e9904563addb15afa96c9ab533ceed0))
* **sequencer:** limit sequencer metadata lists length ([#1149](https://github.com/dymensionxyz/dymension/issues/1149)) ([255f4ee](https://github.com/dymensionxyz/dymension/commit/255f4ee7f0a0e9102effdce3e404c870bec68af1))
* **sequencer:** make create sequencer cli command metadata optional ([#1111](https://github.com/dymensionxyz/dymension/issues/1111)) ([5148d6b](https://github.com/dymensionxyz/dymension/commit/5148d6b8cf0ed272b0fe79c353d2652fc252a676))
* **sequencer:** register PunishSequencerProposal type ([#1656](https://github.com/dymensionxyz/dymension/issues/1656)) ([d0166dd](https://github.com/dymensionxyz/dymension/commit/d0166dd392805c12f2204883aaa197f9945b4230))
* **sequencer:** remove `rollapp_id` from `MsgUpdateSequencerInformation` ([#1110](https://github.com/dymensionxyz/dymension/issues/1110)) ([f0caf5e](https://github.com/dymensionxyz/dymension/commit/f0caf5e3871ac6faf0e0f3e6dc192076a799dfcd))
* **sequencer:** set unbonded on abrupt removal ([#1582](https://github.com/dymensionxyz/dymension/issues/1582)) ([520ac9c](https://github.com/dymensionxyz/dymension/commit/520ac9cf81263f643cabe554a5f5207682c3565c))
* **sequencers:** incorrect sorting mechanism allows manipulation of proposer selection ([#1292](https://github.com/dymensionxyz/dymension/issues/1292)) ([c8b8406](https://github.com/dymensionxyz/dymension/commit/c8b8406eabab9057d7ee7a290b4d40888cdacf04))
* **sequencer:** use AccAddress for operator address ([#1434](https://github.com/dymensionxyz/dymension/issues/1434)) ([4395509](https://github.com/dymensionxyz/dymension/commit/43955096383224cae9248025c8d77a29a7dbcbd9))
* **sponsorship:** missing duplicate vote validation in GenesisState ([#1271](https://github.com/dymensionxyz/dymension/issues/1271)) ([bad4ae3](https://github.com/dymensionxyz/dymension/commit/bad4ae3fa31e3e9792665e50e86712ea462d8264))
* **sponsorship:** remove negative weights from distribution ([#1509](https://github.com/dymensionxyz/dymension/issues/1509)) ([a7a0378](https://github.com/dymensionxyz/dymension/commit/a7a03787193ecdfc8d3d1add134e31f0512c74b2))
* **sponsorship:** removed redundant ceiling while calculating staking power ([#1515](https://github.com/dymensionxyz/dymension/issues/1515)) ([294d8c4](https://github.com/dymensionxyz/dymension/commit/294d8c4325dc62d94272debece489890547c1ac7))
* **streamer:** don't distribute abstained part of sponsored distribution ([#1097](https://github.com/dymensionxyz/dymension/issues/1097)) ([f1df3d7](https://github.com/dymensionxyz/dymension/commit/f1df3d7f7dced4d4ee6505d2bf229d15c4181398))
* tests that run unwanted ([#1132](https://github.com/dymensionxyz/dymension/issues/1132)) ([fe9ba14](https://github.com/dymensionxyz/dymension/commit/fe9ba149a8bade36971a1f787a883858cdd1ad32))
* update messages legacy amino codec registration ([#1403](https://github.com/dymensionxyz/dymension/issues/1403)) ([91f4a2a](https://github.com/dymensionxyz/dymension/commit/91f4a2abed1283f67b9793b9c1edcaf5c5199998))


### Features

* Add canonical light client for Rollapps ([#1098](https://github.com/dymensionxyz/dymension/issues/1098)) ([70b87d7](https://github.com/dymensionxyz/dymension/commit/70b87d7583b75589ae8c52d827b4f24ef74314ef))
* add pagination to sensitive endpoints ([#1601](https://github.com/dymensionxyz/dymension/issues/1601)) ([df0a417](https://github.com/dymensionxyz/dymension/commit/df0a417fc0ab73f40230a6f63ca7f7cbd323a6c0))
* **ante:** allow rejection based on depth ([#1443](https://github.com/dymensionxyz/dymension/issues/1443)) ([17b5be7](https://github.com/dymensionxyz/dymension/commit/17b5be7e8ec03d82d09a1f46175cd5d9c845abba))
* bridging fee middleware ([#899](https://github.com/dymensionxyz/dymension/issues/899)) ([a74ffb0](https://github.com/dymensionxyz/dymension/commit/a74ffb0cec00768bbb8dbe3fd6413e66388010d3))
* change authorized eibc lp fee acceptance criteria ([#1471](https://github.com/dymensionxyz/dymension/issues/1471)) ([80a53c7](https://github.com/dymensionxyz/dymension/commit/80a53c79cf0e69158b2680c8098bae6ced677f84))
* **common:** added more attributes to rollapp packet delayed ack event ([#1381](https://github.com/dymensionxyz/dymension/issues/1381)) ([c150a43](https://github.com/dymensionxyz/dymension/commit/c150a439dd82be0ba838ae64ffc13ff02e97017b))
* **delayedack:** Add type filter for delayedack query ([#860](https://github.com/dymensionxyz/dymension/issues/860)) ([57eca21](https://github.com/dymensionxyz/dymension/commit/57eca21cb5a3b91ebfe08f2e48ee8eeca6534a10))
* **delayedack:** added efficient query for pending packets by addr ([#1385](https://github.com/dymensionxyz/dymension/issues/1385)) ([d577982](https://github.com/dymensionxyz/dymension/commit/d577982f5d915d6184449603ec7b0e1f2f34aad5))
* **delayedack:** finalize rollapp packet by packet base64 key ([#1297](https://github.com/dymensionxyz/dymension/issues/1297)) ([3f08435](https://github.com/dymensionxyz/dymension/commit/3f084355731d2aa946017953c287b53f15cb30cd))
* **delayedack:** fulfill only by packet key and added query for getting recipient packets ([#1338](https://github.com/dymensionxyz/dymension/issues/1338)) ([43d8aca](https://github.com/dymensionxyz/dymension/commit/43d8acabec1be449fda588fecac7ac9b71521744))
* **delayedack:** made packet finalization manual ([#1205](https://github.com/dymensionxyz/dymension/issues/1205)) ([045e6c3](https://github.com/dymensionxyz/dymension/commit/045e6c3b33418d3562c09e1d599933ac578cd15f))
* **delayedack:** paginate rollapp packets when deleting them ([#972](https://github.com/dymensionxyz/dymension/issues/972)) ([1b11625](https://github.com/dymensionxyz/dymension/commit/1b11625498d75a96523a9c49f72a6ebad627c93b))
* **denommetadata:** charge extra fee for IBC denom metadata registration ([#1609](https://github.com/dymensionxyz/dymension/issues/1609)) ([8df230b](https://github.com/dymensionxyz/dymension/commit/8df230b4c7c498ada476ad3a58d113e0aec312b5))
* **denommetadata:** register IBC denom on transfer  ([#956](https://github.com/dymensionxyz/dymension/issues/956)) ([5ba056c](https://github.com/dymensionxyz/dymension/commit/5ba056cb721d23d9cfd2423694dec9d6c3474cca))
* **doc:** add README for canonical light client for operators ([#1213](https://github.com/dymensionxyz/dymension/issues/1213)) ([4782bc4](https://github.com/dymensionxyz/dymension/commit/4782bc4e587fc764f36e1275fd3a25cb0fe82788))
* **dymns:** add Dymension Name Service ([#1007](https://github.com/dymensionxyz/dymension/issues/1007)) ([f7b7296](https://github.com/dymensionxyz/dymension/commit/f7b729652066bb759b72becc8379c4225123429d))
* **eibc:** add EventDemandOrderFulfilledAuthorized ([#1374](https://github.com/dymensionxyz/dymension/issues/1374)) ([f41f1cb](https://github.com/dymensionxyz/dymension/commit/f41f1cb865ed1ebff4ee8ab7c76d529f1cf4a9f7))
* **eibc:** add extra info to demand order ([#950](https://github.com/dymensionxyz/dymension/issues/950)) ([2528e5d](https://github.com/dymensionxyz/dymension/commit/2528e5dfff793605c303f51f5e57475297b326b3))
* **eibc:** add filter by fulfillment to demand orders ([#929](https://github.com/dymensionxyz/dymension/issues/929)) ([87539aa](https://github.com/dymensionxyz/dymension/commit/87539aa943106336ec6687df87104a065006684b))
* **eibc:** add more info to demand order updated event ([#1502](https://github.com/dymensionxyz/dymension/issues/1502)) ([5083e1d](https://github.com/dymensionxyz/dymension/commit/5083e1d86d582fcbcde3fcdc009084e8a676b406))
* **eibc:** add operator address and fee to the eibc authorized event ([#1555](https://github.com/dymensionxyz/dymension/issues/1555)) ([39f197c](https://github.com/dymensionxyz/dymension/commit/39f197cc884df9e8b760314a166b4c224d5947d1))
* **eibc:** auto create eibc demand order for rollapp packets ([#944](https://github.com/dymensionxyz/dymension/issues/944)) ([0274850](https://github.com/dymensionxyz/dymension/commit/027485029f74a9cce3f486a352343a8a1f1a79b5))
* **eibc:** Expand list-demand-orders query with more parameters ([#851](https://github.com/dymensionxyz/dymension/issues/851)) ([c3d9307](https://github.com/dymensionxyz/dymension/commit/c3d93075779ab5fc7fb8ecc5a98c690d220ff59c))
* **eibc:** filtering per rollapp level when granting authorization ([#1375](https://github.com/dymensionxyz/dymension/issues/1375)) ([fca090d](https://github.com/dymensionxyz/dymension/commit/fca090d216d088b5b6734eb7dc45df8c1e7ade41))
* **eibc:** fulfill demand orders with authorization from granter account ([#1326](https://github.com/dymensionxyz/dymension/issues/1326)) ([d9b6cb8](https://github.com/dymensionxyz/dymension/commit/d9b6cb8fd37da12283497550de1063c6b106c08e))
* **eibc:** fulfill order authorization spend limit per rollapp ([#1453](https://github.com/dymensionxyz/dymension/issues/1453)) ([795e23c](https://github.com/dymensionxyz/dymension/commit/795e23cf8896f017af137e4ef2014fd888dedaa9))
* **eibc:** update demand order ([#915](https://github.com/dymensionxyz/dymension/issues/915)) ([6721d98](https://github.com/dymensionxyz/dymension/commit/6721d984c49ba2cca8a7aafd6c2307af0eba758e))
* **endorsement,incentives:** simulation tests ([#1536](https://github.com/dymensionxyz/dymension/issues/1536)) ([ed7dde7](https://github.com/dymensionxyz/dymension/commit/ed7dde7e61772051b5595c38da80404d898be6ae))
* **fork:** restricts to post genesis transfer heights only ([#1600](https://github.com/dymensionxyz/dymension/issues/1600)) ([3e0c553](https://github.com/dymensionxyz/dymension/commit/3e0c5530c1f913b1d52f137e648dfbd499dd27a0))
* **gamm:** added rewarding of OUT denom in case of DYM-rollapp swaps ([#1527](https://github.com/dymensionxyz/dymension/issues/1527)) ([6c6e540](https://github.com/dymensionxyz/dymension/commit/6c6e54005ccbddc11a62a26454a6c30dd3021cb6))
* **genesis bridge:** genesis transfers ([#933](https://github.com/dymensionxyz/dymension/issues/933)) ([4565f34](https://github.com/dymensionxyz/dymension/commit/4565f34fe8a8d2cd9e34a42bd6e3cf71ecbfecd9))
* **genesis_bridge:** json encoding for genesisBridgeData ([#1337](https://github.com/dymensionxyz/dymension/issues/1337)) ([6793f65](https://github.com/dymensionxyz/dymension/commit/6793f65786f3878fef236dd2945f433d67db6370))
* **genesis_bridge:** revised genesis bridge impl ([#1288](https://github.com/dymensionxyz/dymension/issues/1288)) ([c7c6883](https://github.com/dymensionxyz/dymension/commit/c7c688334550ec8c5d7243c3ffd3b3942046eefe))
* **genesistransfer:** open the bridge without waiting for a genesis transfer ([#1227](https://github.com/dymensionxyz/dymension/issues/1227)) ([22a8713](https://github.com/dymensionxyz/dymension/commit/22a8713333fdda5822a1cf211abb423f8929f529))
* **group:** add group module ([#1329](https://github.com/dymensionxyz/dymension/issues/1329)) ([0382bbc](https://github.com/dymensionxyz/dymension/commit/0382bbcc95d439168a3b4686f92aa1834e5385fe))
* **group:** Increased group policy metadata  ([#1613](https://github.com/dymensionxyz/dymension/issues/1613)) ([9c96919](https://github.com/dymensionxyz/dymension/commit/9c9691999600ad5adc2d0ef2771ac5b3bbb1925a))
* **hard-fork:** Implement rollapp hard fork  ([#1354](https://github.com/dymensionxyz/dymension/issues/1354)) ([5deae85](https://github.com/dymensionxyz/dymension/commit/5deae8512b53e41b64c0cba1209151802ada06d9))
* **ibc transfer:** Register IBC denom on transfer ([#900](https://github.com/dymensionxyz/dymension/issues/900)) ([78494bd](https://github.com/dymensionxyz/dymension/commit/78494bd3fec5c1a3a01db7b1c1c0cf2acfff882b))
* **ibc:** add debug log and reason when client fails to be made canonical ([#1349](https://github.com/dymensionxyz/dymension/issues/1349)) ([a41e964](https://github.com/dymensionxyz/dymension/commit/a41e9647a52f1bb0e48712ea7b7ef464aa9fe244))
* **incentives:**  earning events  ([#1545](https://github.com/dymensionxyz/dymension/issues/1545)) ([d2b9920](https://github.com/dymensionxyz/dymension/commit/d2b9920d9496f0408c9ed494c033a8dffe3523ca))
* **incentives:** added fees for adding to gauge and gauge creation ([#1188](https://github.com/dymensionxyz/dymension/issues/1188)) ([7e83549](https://github.com/dymensionxyz/dymension/commit/7e83549dd623a3f5318deff97932c8b1c339ec67))
* **incentives:** feature flag to enable/disable epoch end distribution ([#1648](https://github.com/dymensionxyz/dymension/issues/1648)) ([c3eaedf](https://github.com/dymensionxyz/dymension/commit/c3eaedffb8886fd219819cd3a21304f931352bed))
* **incentives:** rollapp gauges ([#947](https://github.com/dymensionxyz/dymension/issues/947)) ([6483533](https://github.com/dymensionxyz/dymension/commit/648353317fbdecc785cf207e8944e448fbe6237e))
* **incentives:** rollapp gauges upgrade handler and burn creation fee ([#1113](https://github.com/dymensionxyz/dymension/issues/1113)) ([ac7d25e](https://github.com/dymensionxyz/dymension/commit/ac7d25edc665d2c0b53d8b03717ff8f0def0dd3a))
* **incentives:** send rewards to rollapp creator address ([#1047](https://github.com/dymensionxyz/dymension/issues/1047)) ([88a5f9e](https://github.com/dymensionxyz/dymension/commit/88a5f9e0b9c236439536d4f76062d3018eead68e))
* **invariants:** add several invariants across modules ([#1514](https://github.com/dymensionxyz/dymension/issues/1514)) ([3109e57](https://github.com/dymensionxyz/dymension/commit/3109e579353ee0c0a64b3e9e2759b55f74416a80))
* **iro,amm:** adding `closing price` event attribute for trades ([#1359](https://github.com/dymensionxyz/dymension/issues/1359)) ([084f756](https://github.com/dymensionxyz/dymension/commit/084f756abc454ab38bcd3f538de1d53dcd842ac7))
* **iro:** add denom to buy, sell and claim events ([#1460](https://github.com/dymensionxyz/dymension/issues/1460)) ([42e2184](https://github.com/dymensionxyz/dymension/commit/42e21849bf92c0f1f0d5951e51638f615f5a4cd2))
* **iro:** add tradable only filter to IRO plans ([#1358](https://github.com/dymensionxyz/dymension/issues/1358)) ([632c277](https://github.com/dymensionxyz/dymension/commit/632c277be450097ab6088e16a3ec3cedb24bed87))
* **IRO:** allow iro buy with dym amount as input  ([#1318](https://github.com/dymensionxyz/dymension/issues/1318)) ([0a553f9](https://github.com/dymensionxyz/dymension/commit/0a553f990e49eeb42ea2f53907975b7747a2fe6d))
* **iro:** change IRO creation fee to be derived from rollapp token ([#1632](https://github.com/dymensionxyz/dymension/issues/1632)) ([e7da850](https://github.com/dymensionxyz/dymension/commit/e7da850d89c4990908fd7c00379e68756a55c3e0))
* **IRO:** IRO module implementation ([#1201](https://github.com/dymensionxyz/dymension/issues/1201)) ([5601240](https://github.com/dymensionxyz/dymension/commit/5601240fa78d51c73415a9cd84332d05b48907e2))
* **IRO:** IRO Rollapp token creation fee  ([#1333](https://github.com/dymensionxyz/dymension/issues/1333)) ([198f418](https://github.com/dymensionxyz/dymension/commit/198f4187dc9aeeb403f031fa4cb28c86a19a3dd8))
* **iro:** IRO simulation tests ([#1592](https://github.com/dymensionxyz/dymension/issues/1592)) ([30391f3](https://github.com/dymensionxyz/dymension/commit/30391f34c9ba1f838498be0cbdbbbcb4f3038d96))
* **IRO:** no trade allowed before start time ([#1456](https://github.com/dymensionxyz/dymension/issues/1456)) ([366c666](https://github.com/dymensionxyz/dymension/commit/366c666dd6edf70910b3fde0b28ecfd8c1d42c8a))
* **iro:** scale the bonding curve ([#1247](https://github.com/dymensionxyz/dymension/issues/1247)) ([f88af02](https://github.com/dymensionxyz/dymension/commit/f88af02610a0e276c45e54a5d2663a79bdd6a8e5))
* **light client:** adds query for canonical client for rollapp ([#1158](https://github.com/dymensionxyz/dymension/issues/1158)) ([031ecad](https://github.com/dymensionxyz/dymension/commit/031ecadf4249c1c79b93ab85fa8e86072fa5a2a1))
* **light client:** query expected light client ([#1204](https://github.com/dymensionxyz/dymension/issues/1204)) ([f8b420c](https://github.com/dymensionxyz/dymension/commit/f8b420ccd059d2d79a65f3464a3cc9967d8db580))
* **lightclient:** add an event when canonical channel is set ([#1310](https://github.com/dymensionxyz/dymension/issues/1310)) ([0cad2f1](https://github.com/dymensionxyz/dymension/commit/0cad2f1e1554ebb0b67f82ec48d8cd63965fe215))
* **lightclient:** Add query for getting canonical channel  ([#1637](https://github.com/dymensionxyz/dymension/issues/1637)) ([58cbe8d](https://github.com/dymensionxyz/dymension/commit/58cbe8d036dd5cd087a9db07f3a9f771b8f58a15))
* **liveness:** better tests | general code cleanup ([#1532](https://github.com/dymensionxyz/dymension/issues/1532)) ([6734dc3](https://github.com/dymensionxyz/dymension/commit/6734dc37bd1cebb400436b079fdba979c3152e29))
* **liveness:** improve tests ([#1623](https://github.com/dymensionxyz/dymension/issues/1623)) ([de9f463](https://github.com/dymensionxyz/dymension/commit/de9f46340e1ca704e825313909512ce1937f979f))
* **lockup:** added const fee for locking tokens ([#1543](https://github.com/dymensionxyz/dymension/issues/1543)) ([93b6385](https://github.com/dymensionxyz/dymension/commit/93b6385b94a9f3a59df8964dc67e535e92fa8dac))
* **lockup:** moved the module from the Osmosis fork ([#1154](https://github.com/dymensionxyz/dymension/issues/1154)) ([d440dfa](https://github.com/dymensionxyz/dymension/commit/d440dfa78ac44dd9ef739b04f1145d2fff0f9b83))
* **migration:** add migrations for sequencer module ([#1421](https://github.com/dymensionxyz/dymension/issues/1421)) ([3b31d78](https://github.com/dymensionxyz/dymension/commit/3b31d78235879f6253402d173f8f0ca8d04df215))
* **migration:** move module migrations before global migration ([#1425](https://github.com/dymensionxyz/dymension/issues/1425)) ([eb0704f](https://github.com/dymensionxyz/dymension/commit/eb0704fb42f62498a9187502cb036aa40811af45))
* **migrations:** crisis module deprecation ([#1444](https://github.com/dymensionxyz/dymension/issues/1444)) ([cab29c2](https://github.com/dymensionxyz/dymension/commit/cab29c2f3a4526e6d290b7625b22adc4e9a986ed))
* **migrations:** next proposer in state info ([#1482](https://github.com/dymensionxyz/dymension/issues/1482)) ([a308fd0](https://github.com/dymensionxyz/dymension/commit/a308fd03b2c4256d57576bfb21c99ab13f9de263))
* **migrations:** testnet rollapps ([#1475](https://github.com/dymensionxyz/dymension/issues/1475)) ([0bc444c](https://github.com/dymensionxyz/dymension/commit/0bc444c63458e9590b151198868cb353512670f6))
* Register incentives and sponsorship amino message to support eip712 ([#1240](https://github.com/dymensionxyz/dymension/issues/1240)) ([bb2cd89](https://github.com/dymensionxyz/dymension/commit/bb2cd8927211109d225800663ef21ce04c0e1b70))
* removed incentives feature flag and decreased IRO `MinEpochsPaidOver` default value ([#1652](https://github.com/dymensionxyz/dymension/issues/1652)) ([02cbaec](https://github.com/dymensionxyz/dymension/commit/02cbaecdc358b49590b5ff6ab578dea355375dda))
* rollapp royalties ([#1368](https://github.com/dymensionxyz/dymension/issues/1368)) ([a6f9c55](https://github.com/dymensionxyz/dymension/commit/a6f9c558526aeea2391c7a4f91a26798ddc992ee))
* **rollapp:** add apps to rollapp ([#1131](https://github.com/dymensionxyz/dymension/issues/1131)) ([35f7a94](https://github.com/dymensionxyz/dymension/commit/35f7a94c7b58c9d02e8b0454c59713d899ea2cc5))
* **rollapp:** add genesis url to metadata ([#1082](https://github.com/dymensionxyz/dymension/issues/1082)) ([61d5ab1](https://github.com/dymensionxyz/dymension/commit/61d5ab18e17758f88b0b3b3a888304e9da09578c))
* **rollapp:** add nim and mande domain name in migration ([#1498](https://github.com/dymensionxyz/dymension/issues/1498)) ([2295106](https://github.com/dymensionxyz/dymension/commit/2295106b1c87fcedb7d57a9c53eb0998a4471e6c))
* **rollapp:** add tags to rollapp metadata ([#1572](https://github.com/dymensionxyz/dymension/issues/1572)) ([a7df970](https://github.com/dymensionxyz/dymension/commit/a7df970182fdc0a68e09745a2ec0de5fcda146c9))
* **rollapp:** add vm type to rollapp registration ([#1040](https://github.com/dymensionxyz/dymension/issues/1040)) ([2a8d639](https://github.com/dymensionxyz/dymension/commit/2a8d6398690c88f0d7b4b3fe9952a9138d1acfa1))
* **rollapp:** allow rollapps with no native token ([#1654](https://github.com/dymensionxyz/dymension/issues/1654)) ([a9d5d6b](https://github.com/dymensionxyz/dymension/commit/a9d5d6bb0ba1b8102dc92a0a1c3b83b473352bba))
* **rollapp:** allow supplying bech32 and genesis checksum in update rather than create ([#1089](https://github.com/dymensionxyz/dymension/issues/1089)) ([c0216b0](https://github.com/dymensionxyz/dymension/commit/c0216b0341f2762e84231925ad5dd80797b4efbe))
* **rollapp:** create gov prop for updating genesis info ([#1570](https://github.com/dymensionxyz/dymension/issues/1570)) ([81a984c](https://github.com/dymensionxyz/dymension/commit/81a984cb5fa4ec80cbb095559af59818b11677a0))
* **rollapp:** delete stale state updates ([#1176](https://github.com/dymensionxyz/dymension/issues/1176)) ([0f876b3](https://github.com/dymensionxyz/dymension/commit/0f876b3a9fd71a0c27b0e7f68420a86251ad56be))
* **rollapp:** DRS versions ([#1223](https://github.com/dymensionxyz/dymension/issues/1223)) ([05e445c](https://github.com/dymensionxyz/dymension/commit/05e445c00e8cfaff0c8891bde51908e0d80c9c3c))
* **rollapp:** enforce EIP155 chain ID for rollapp creation ([#1020](https://github.com/dymensionxyz/dymension/issues/1020)) ([55a7517](https://github.com/dymensionxyz/dymension/commit/55a7517ceaa151cb22c4f40064f7aa1ae5b89d88))
* **rollapp:** make the test for updating min bond a bit clearer ([#1607](https://github.com/dymensionxyz/dymension/issues/1607)) ([e26b17e](https://github.com/dymensionxyz/dymension/commit/e26b17e6e5adfd75ff081046a8bae0dcecdeab6c))
* **rollapp:** migrate drs from commit string to integer ([#1362](https://github.com/dymensionxyz/dymension/issues/1362)) ([ca94133](https://github.com/dymensionxyz/dymension/commit/ca941338b44335640a4de4ef500ad25d4f64c80e))
* **rollapp:** min bond is now defined on rollapp level ([#1579](https://github.com/dymensionxyz/dymension/issues/1579)) ([89d3abd](https://github.com/dymensionxyz/dymension/commit/89d3abd3607899c32df66fd39df1e431ddeea6b4))
* **rollapp:** move DRS version to be part of the block descriptor ([#1311](https://github.com/dymensionxyz/dymension/issues/1311)) ([8fb2d25](https://github.com/dymensionxyz/dymension/commit/8fb2d25c84c3d9e8cbde28de3f5309269979876c))
* **rollapp:** new rollapp registration flow ([#980](https://github.com/dymensionxyz/dymension/issues/980)) ([afd8254](https://github.com/dymensionxyz/dymension/commit/afd8254a010c0a177c7f9a8896f49d01084101c0))
* **rollapp:** obsolete DRS versions query ([#1445](https://github.com/dymensionxyz/dymension/issues/1445)) ([f621fb1](https://github.com/dymensionxyz/dymension/commit/f621fb11dace75276cf49337d0548e9daaaed644))
* **rollapp:** option to transfer rollapp owner  ([#1039](https://github.com/dymensionxyz/dymension/issues/1039)) ([ce6f078](https://github.com/dymensionxyz/dymension/commit/ce6f0784b3f8cede95eff60aa3054e199ba7d147))
* **rollapp:** prefixed finalization queue with rollappID ([#1390](https://github.com/dymensionxyz/dymension/issues/1390)) ([6a36f7b](https://github.com/dymensionxyz/dymension/commit/6a36f7b37a0ffb9b2b2568d693138d78bdc7beb2))
* **rollapp:** query for genesis bridge validation ([#1564](https://github.com/dymensionxyz/dymension/issues/1564)) ([62fe71c](https://github.com/dymensionxyz/dymension/commit/62fe71c7364005a0aadb8bb422afba9bd9c952ba))
* **rollapp:** refactor rollapp cli to be more useful ([#842](https://github.com/dymensionxyz/dymension/issues/842)) ([90e8b37](https://github.com/dymensionxyz/dymension/commit/90e8b37f92473493a33ee13ef839a2efef90b84a))
* **rollapp:** register rollapp amino messages to support EIP712 ([#1229](https://github.com/dymensionxyz/dymension/issues/1229)) ([f140cd1](https://github.com/dymensionxyz/dymension/commit/f140cd1dd561cefb3e6562cbf4379b88cd16400d))
* **rollapp:** remove alias from rollapp ([#1034](https://github.com/dymensionxyz/dymension/issues/1034)) ([71eac1b](https://github.com/dymensionxyz/dymension/commit/71eac1b3a2a0c8ab4696c36314a3670462c66a23))
* **rollapp:** store all rollapp revisions ([#1476](https://github.com/dymensionxyz/dymension/issues/1476)) ([4cad7dc](https://github.com/dymensionxyz/dymension/commit/4cad7dc256c2835629a39bee431ff1af2a9bfbb6))
* **rollapp:** store rollapp revision history ([#1507](https://github.com/dymensionxyz/dymension/issues/1507)) ([20c6e5a](https://github.com/dymensionxyz/dymension/commit/20c6e5a4a984f9d09edeadf88f80faebf82db0f6))
* **rollapp:** verify genesis checksum is same in hub and rollapp ([#1384](https://github.com/dymensionxyz/dymension/issues/1384)) ([5386029](https://github.com/dymensionxyz/dymension/commit/53860295cce88229f16334e43dd1db3d35f2d4a7))
* **scripts:** updated scripts for local run, added sponsored stream creation ([#1046](https://github.com/dymensionxyz/dymension/issues/1046)) ([c8fc1fd](https://github.com/dymensionxyz/dymension/commit/c8fc1fdd20b6138b4acd7465a1cc3cb18c9fe077))
* **sequencer, rollapp:** liveness slashing and jailing ([#1009](https://github.com/dymensionxyz/dymension/issues/1009)) ([04bd9df](https://github.com/dymensionxyz/dymension/commit/04bd9df0eec8abff6e24fc8e089be7b596227216))
* **sequencer:** add fee denom to sequencer metadata ([#1662](https://github.com/dymensionxyz/dymension/issues/1662)) ([c955648](https://github.com/dymensionxyz/dymension/commit/c955648bacbff2d9ddaf089f1f99a20da0cef65d))
* **sequencer:** add proposal type and handler for punishing sequencer ([#1581](https://github.com/dymensionxyz/dymension/issues/1581)) ([0988440](https://github.com/dymensionxyz/dymension/commit/09884405ba6c52212b54c9eb989cd5de9a2da24a))
* **sequencer:** added reward addr and whitelisted relayers for sequencer ([#1313](https://github.com/dymensionxyz/dymension/issues/1313)) ([8739ceb](https://github.com/dymensionxyz/dymension/commit/8739cebebfacec254e750e798721e5a90bd6c1d8))
* **sequencer:** Allow a sequencer to decrease their bond ([#1031](https://github.com/dymensionxyz/dymension/issues/1031)) ([acd9d13](https://github.com/dymensionxyz/dymension/commit/acd9d13935a98dd43d13affdd440a78c85f3a961))
* **sequencer:** Allow a sequencer to increase their bond  ([#1015](https://github.com/dymensionxyz/dymension/issues/1015)) ([7d27b75](https://github.com/dymensionxyz/dymension/commit/7d27b75aeb8bf7105583ae35d2d02ade6e972afd))
* **sequencer:** conditional unbonding + kick for liveness ([#1343](https://github.com/dymensionxyz/dymension/issues/1343)) ([76aa1e1](https://github.com/dymensionxyz/dymension/commit/76aa1e103e0440d4a96bac27f09ae03a27acf3fc))
* **sequencer:** Defining invariants around bond reductions ([#1144](https://github.com/dymensionxyz/dymension/issues/1144)) ([d986472](https://github.com/dymensionxyz/dymension/commit/d986472a731faaf0653e97e089124cacb5c6a738))
* **sequencer:** enforce endpoints on sequencer registration ([#1043](https://github.com/dymensionxyz/dymension/issues/1043)) ([c2c92a9](https://github.com/dymensionxyz/dymension/commit/c2c92a9274a3c1544b9b232bee16678029f4b682))
* **sequencer:** Implement genesis Import/Export for Bond Reductions ([#1075](https://github.com/dymensionxyz/dymension/issues/1075)) ([155db63](https://github.com/dymensionxyz/dymension/commit/155db63cf42939f948199ea85f66027544df8304))
* **sequencer:** kick rework (dishonor) ([#1647](https://github.com/dymensionxyz/dymension/issues/1647)) ([29d227f](https://github.com/dymensionxyz/dymension/commit/29d227f910f2165c2f4cba5a0d1de7bcc37df5dc))
* **sequencer:** query all proposers ([#1221](https://github.com/dymensionxyz/dymension/issues/1221)) ([e670723](https://github.com/dymensionxyz/dymension/commit/e67072380dec2269b9b6aab4d6160aa5961fab53))
* **sequencers:** sequencer rotation support ([#1006](https://github.com/dymensionxyz/dymension/issues/1006)) ([e50966c](https://github.com/dymensionxyz/dymension/commit/e50966c0c82d340a4d2c6b36ecc354b00a900e09))
* **sequencer:** support rotation misbehavior detection  ([#1345](https://github.com/dymensionxyz/dymension/issues/1345)) ([43db5fd](https://github.com/dymensionxyz/dymension/commit/43db5fdd831f0b78578bdc1a7f0f45f7a43af783))
* **sequencers:** validate sequencer `unbonding_time` greater than `dispute_period` ([#1115](https://github.com/dymensionxyz/dymension/issues/1115)) ([eddfc50](https://github.com/dymensionxyz/dymension/commit/eddfc50581f58ba34c2372b83119333c5d9fd06f))
* **sponsorship:** added proto contracts ([#990](https://github.com/dymensionxyz/dymension/issues/990)) ([54c0fca](https://github.com/dymensionxyz/dymension/commit/54c0fca5fada0292428f1721ff2475d4f8590f1d))
* **sponsorship:** added tests to verify staking power truncation issue ([#1523](https://github.com/dymensionxyz/dymension/issues/1523)) ([c6826d7](https://github.com/dymensionxyz/dymension/commit/c6826d75e77f6e16b09157d05b0caefdd7994698))
* **sponsorship:** added voting mechanisms and staking hooks ([#1044](https://github.com/dymensionxyz/dymension/issues/1044)) ([836e500](https://github.com/dymensionxyz/dymension/commit/836e50018d63d2e3caba522c91f4e932b9c540cd))
* **sponsorship:** fixed revoke vote CLI ([#1270](https://github.com/dymensionxyz/dymension/issues/1270)) ([aa05f98](https://github.com/dymensionxyz/dymension/commit/aa05f982fbb3eb0963dc3af1e9e6c5c4e194318b))
* **sponsorship:** module scaffolding ([#994](https://github.com/dymensionxyz/dymension/issues/994)) ([695583c](https://github.com/dymensionxyz/dymension/commit/695583c7eeeaf828c724f49035f5f328853e7f90))
* **sponsorship:** only allow sponsoring rollapps with bonded sequencers ([#1275](https://github.com/dymensionxyz/dymension/issues/1275)) ([b738cff](https://github.com/dymensionxyz/dymension/commit/b738cff066f1862739a59ae29e8be81f02a20fe9))
* **sponsorship:** remove sequencer bonded enforcement from rollapp gauge ([#1497](https://github.com/dymensionxyz/dymension/issues/1497)) ([f74b114](https://github.com/dymensionxyz/dymension/commit/f74b114a619066d0715bc43791ff5aa4606558d0))
* **sponsorship:** updated module params comment ([#1199](https://github.com/dymensionxyz/dymension/issues/1199)) ([904c0d1](https://github.com/dymensionxyz/dymension/commit/904c0d168216c2e9e6c8270e121a94711693dcaa))
* **sponsorship:** updated the scale system for gauge weights ([#1190](https://github.com/dymensionxyz/dymension/issues/1190)) ([094de0c](https://github.com/dymensionxyz/dymension/commit/094de0cf8601fa2cc2bbf29042bade1e77b6fe06))
* **streamer:** added sponsored distribution support ([#1045](https://github.com/dymensionxyz/dymension/issues/1045)) ([5933d5c](https://github.com/dymensionxyz/dymension/commit/5933d5c8872ca39e01408f6dc7e1d6ef30547502))
* **streamer:** added streamer pagination ([#1100](https://github.com/dymensionxyz/dymension/issues/1100)) ([8a7b79e](https://github.com/dymensionxyz/dymension/commit/8a7b79ec97304376dac52ac5e603498a15b91b8b))
* **streamer:** distribute rewards immediately in the current block ([#1173](https://github.com/dymensionxyz/dymension/issues/1173)) ([14877e9](https://github.com/dymensionxyz/dymension/commit/14877e9ef3b3f404e17f8a85df2bbda22d3ad4d6))
* **swagger:** add make command `proto-swagger-gen` ([#856](https://github.com/dymensionxyz/dymension/issues/856)) ([d79ac7b](https://github.com/dymensionxyz/dymension/commit/d79ac7ba7ce6d629d02420d9be30d3cf07a89bd7))


### Performance Improvements

* use errors.New to replace fmt.Errorf with no parameters ([#857](https://github.com/dymensionxyz/dymension/issues/857)) ([2f0f1e2](https://github.com/dymensionxyz/dymension/commit/2f0f1e2bf0de521da68f61a2c5abe6c2de5f1385))


### Reverts

* **streamer:** don't distribute abstained part of sponsored distri… ([#1117](https://github.com/dymensionxyz/dymension/issues/1117)) ([908cb39](https://github.com/dymensionxyz/dymension/commit/908cb39023ef48f2b2f9ee2627e02d36ae6e6ae9))


## [v3.1.0](https://github.com/dymensionxyz/dymension/releases/tag/v3.1.0)

### Features

- (rollapp) [#999](https://github.com/dymensionxyz/dymension/issues/999) Handle sequencer information updates.
- (rollapp) [#996](https://github.com/dymensionxyz/dymension/issues/996) Handle rollapp information updates.
- (sequencer) [#955](https://github.com/dymensionxyz/dymension/issues/979) Rework the sequencer registration flow.
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
- (delayedack) [#670](https://github.com/dymensionxyz/dymension/issues/670) Finalize error handling per rollapp
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
