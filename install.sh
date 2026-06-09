#!/usr/bin/env sh
set -eu

repo_url="${KUMITE_REPO_URL:-https://github.com/OnishellT/kumite.git}"
repo_ref="${KUMITE_REPO_REF:-main}"
bin_dir="${KUMITE_BIN_DIR:-}"
setup_global="${KUMITE_SETUP_GLOBAL:-0}"
force_yes="${KUMITE_YES:-0}"
dry_run="${KUMITE_DRY_RUN:-0}"

need() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "kumite installer requires $1 on PATH" >&2
		exit 1
	fi
}

need go

cleanup_dir=""
if [ -f "./cmd/kumite-installer/main.go" ] && [ -f "./go.mod" ]; then
	source_dir="$(pwd)"
else
	need git
	cleanup_dir="$(mktemp -d)"
	git clone --depth 1 --branch "$repo_ref" "$repo_url" "$cleanup_dir/kumite"
	source_dir="$cleanup_dir/kumite"
fi

set -- --source-dir "$source_dir"
if [ -n "$bin_dir" ]; then
	set -- "$@" --bin-dir "$bin_dir"
fi
if [ "$setup_global" = "1" ]; then
	set -- "$@" --setup-global
fi
if [ "$force_yes" = "1" ] || [ ! -t 0 ] || [ ! -t 1 ]; then
	set -- "$@" --yes
fi
if [ "$dry_run" = "1" ]; then
	set -- "$@" --dry-run
fi

(cd "$source_dir" && go run ./cmd/kumite-installer "$@")

if [ -n "$cleanup_dir" ]; then
	rm -rf "$cleanup_dir"
fi
