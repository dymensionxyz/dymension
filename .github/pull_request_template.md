# PR Standards

## Opening a pull request should be able to meet the following requirements:

---

For Author:

- [ ]  Added a relevant changelog entry to the `Unreleased` section in `CHANGELOG.md`
- [ ]  Targeted PR against the correct branch
- [ ]  included the correct [type prefix](https://github.com/commitizen/conventional-commit-types/blob/v3.0.0/index.json) in the PR title
- [ ]  Linked to the GitHub issue with discussion and accepted design
- [ ]  Targets only one GitHub issue
- [ ]  Wrote unit and integration tests
- [ ]  Wrote relevant migration scripts if necessary
- [ ]  All CI checks have passed
- [ ]  Added relevant `godoc` [comments](https://blog.golang.org/godoc-documenting-go-code)

---

For Reviewer:

- [ ]  Confirmed the correct [type prefix](https://github.com/commitizen/conventional-commit-types/blob/v3.0.0/index.json) in the PR title
- [ ]  Reviewers assigned
- [ ]  Confirmed all author checklist items have been addressed

---

After reviewer approval:

- [ ]  In case the PR targets the main branch, PR should not be squash merge in order to keep meaningful git history.
- [ ]  In case the PR targets a release branch, PR must be rebased.
