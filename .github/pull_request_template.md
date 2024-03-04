## Description

<!-- Add a description of the changes that this PR introduces and the files that
are the most critical to review.
-->

----

Closes #XXX

**All** items are required. Please add a note to the item if the item is not applicable and
please add links to any relevant follow-up issues.

PR review checkboxes:

I have...

- [ ]  Added a relevant changelog entry to the `Unreleased` section in `CHANGELOG.md`
- [ ]  Targeted PR against the correct branch
- [ ]  included the correct [type prefix](https://github.com/commitizen/conventional-commit-types/blob/v3.0.0/index.json) in the PR title
- [ ]  Linked to the GitHub issue with discussion and accepted design
- [ ]  Targets only one GitHub issue
- [ ]  Wrote unit and integration tests
- [ ]  Wrote relevant migration scripts if necessary
- [ ]  All CI checks have passed
- [ ]  Added relevant `godoc` [comments](https://blog.golang.org/godoc-documenting-go-code)
- [ ]  Updated the scripts for local run, e.g genesis_config_commands.sh if the PR changes parameters
- [ ]  Add an issue in the [e2e-tests repo](https://github.com/dymensionxyz/e2e-tests) if necessary

SDK Checklist
- [ ] Import/Export Genesis
- [ ] Registered Invariants
- [ ] Registered Events
- [ ] Updated openapi.yaml
- [ ] No usage of go `map`
- [ ] No usage of `time.Now()`
- [ ] Used fixed point arithmetic and not float arithmetic
- [ ] Avoid panicking in Begin/End block as much as possible
- [ ] No unexpected math Overflow
- [ ] Used `sendCoin` and not `SendCoins`
- [ ] Out-of-block compute is bounded
- [ ] No serialized ID at the end of store keys
- [ ] UInt to byte conversion should use BigEndian

Full security checklist [here](https://www.faulttolerant.xyz/2024-01-16-cosmos-security-1/)


----;

For Reviewer:

- [ ]  Confirmed the correct [type prefix](https://github.com/commitizen/conventional-commit-types/blob/v3.0.0/index.json) in the PR title
- [ ]  Reviewers assigned
- [ ]  Confirmed all author checklist items have been addressed

---;

After reviewer approval:

- [ ]  In case the PR targets the main branch, PR should not be squash merge in order to keep meaningful git history.
- [ ]  In case the PR targets a release branch, PR must be rebased.
