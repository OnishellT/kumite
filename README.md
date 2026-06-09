# Kumite

Kumite installs and configures the Pi agent coding harness for projects that use the kumite subagent workflow.

## Install

Install from GitHub:

```sh
curl -fsSL https://raw.githubusercontent.com/OnishellT/kumite/main/install.sh | sh
```

Install and immediately run global setup:

```sh
curl -fsSL https://raw.githubusercontent.com/OnishellT/kumite/main/install.sh | KUMITE_SETUP_GLOBAL=1 sh
```

Install from a local checkout:

```sh
git clone https://github.com/OnishellT/kumite.git
cd kumite
./install.sh
```

Install from a branch or a fork:

```sh
KUMITE_REPO_REF=development \
KUMITE_REPO_URL=https://github.com/OnishellT/kumite.git \
sh -c "$(curl -fsSL https://raw.githubusercontent.com/OnishellT/kumite/main/install.sh)"
```

Useful installer options:

```sh
KUMITE_BIN_DIR="$HOME/.local/bin" ./install.sh
KUMITE_SETUP_GLOBAL=1 ./install.sh
KUMITE_YES=1 ./install.sh
```

The installer requires `go` on `PATH`. When it is not run from a checkout, it also requires `git` so it can clone the repository before building the `kumite` binary.

## Daily Use

Install global Pi extensions and static-analysis tooling:

```sh
kumite setup --global --keep-going
```

Initialize a project with kumite Pi agents, skills, chain, MCP config, `agents.md` index, and memory docs:

```sh
cd /path/to/project
kumite init
```

`kumite init` also upserts `.pi/settings.json` with `npm:pi-kumite` so Pi loads the Kumite extension in future sessions. For local extension development, use `kumite init --pi-package /path/to/pi-kumite`.

`kumite init` upserts `agents.md`, the curated agent entry point and index for project memory. It links to deeper memory files instead of duplicating them.

Run the full setup for the current project, including global tooling and project files:

```sh
kumite setup --keep-going
```

## Branches

- `development`: active integration branch. Pull requests into this branch run code quality and tests, but not release builds.
- `staging`: release-candidate branch. Pull requests into this branch run code quality and tests, but not release builds.
- `main`: release branch. Pull requests into this branch run code quality, tests, and build checks. Merges to `main` create a GitHub release from `VERSION`.

All protected branches require an approving review from `@OnishellT`. See [Branching And Releases](docs/branching-and-releases.md).

## Development Checks

Install local Git hooks:

```sh
make install-go-tools
python -m pip install pre-commit
pre-commit install
```

Run the same checks used by GitHub Actions:

```sh
make ci
```

GitHub Actions runs quality and test checks on pull requests into `development`, `staging`, and `main`. Pull requests into `main` also run build checks. After a merge to `main`, the release workflow builds Linux, macOS, and Windows binaries and creates a GitHub release.
