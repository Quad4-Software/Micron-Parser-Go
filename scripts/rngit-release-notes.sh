#!/usr/bin/env bash
# EDITOR wrapper for rngit release create.
# Overwrites the notes tempfile with the CHANGELOG section for RNGIT_RELEASE_TAG.

set -euo pipefail

notes_file="${1:-}"
if [[ -z "$notes_file" ]]; then
	echo "usage: $0 <notes-file>" >&2
	exit 2
fi

tag="${RNGIT_RELEASE_TAG:-}"
changelog="${RNGIT_CHANGELOG:-CHANGELOG.md}"
root="${RNGIT_REPO_ROOT:-.}"

if [[ -z "$tag" ]]; then
	echo "RNGIT_RELEASE_TAG is required" >&2
	exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$script_dir/changelog-entry.sh" "$tag" "$root/$changelog" >"$notes_file"
