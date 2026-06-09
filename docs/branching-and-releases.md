# Branching And Releases

Kumite uses three protected branches:

- `development`: active integration branch. Pull requests into this branch run CI quality and tests only.
- `staging`: release-candidate branch. Pull requests into this branch run the same CI quality and tests.
- `main`: release branch. Pull requests into this branch run CI quality, tests, and build checks. Merges to `main` create a GitHub release from `VERSION`.

All protected branches require:

- pull requests before merge;
- one approving review;
- code owner review by `@OnishellT`;
- stale review dismissal after new commits;
- passing required status checks;
- resolved conversations.

Copilot automatic code review is configured with a repository ruleset when the GitHub account has Copilot code review enabled.

## Release Process

1. Merge feature work into `development`.
2. Promote `development` into `staging` with a pull request.
3. Update `VERSION` on `staging` when preparing a release.
4. Promote `staging` into `main` with a pull request.
5. After merge to `main`, the release workflow builds platform binaries and creates the GitHub release.
