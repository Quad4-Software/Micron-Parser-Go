#!/usr/bin/env bash
# Stage release artifacts then create an rngit release with CHANGELOG notes.
#
# Usage:
#   scripts/rngit-release.sh v1.1.0
#   scripts/rngit-release.sh v1.1.0 ./dist
#
# If ARTIFACTS is omitted, downloads GitHub release assets for the tag.
# Env:
#   RNS_REMOTE   default rns://06a54b505bb67b25ef3f8097e8001edc/public/micron-parser-go
#   GH_REPO      default Quad4-Software/Micron-Parser-Go
#   CHANGELOG    default CHANGELOG.md

set -euo pipefail

tag="${1:-}"
artifacts_in="${2:-}"

if [[ -z "$tag" ]]; then
	echo "usage: $0 <tag> [artifacts-dir]" >&2
	exit 2
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

RNS_REMOTE="${RNS_REMOTE:-rns://06a54b505bb67b25ef3f8097e8001edc/public/micron-parser-go}"
GH_REPO="${GH_REPO:-Quad4-Software/Micron-Parser-Go}"
CHANGELOG="${CHANGELOG:-CHANGELOG.md}"
ver="${tag#v}"

if ! git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
	echo "local tag not found: $tag" >&2
	exit 1
fi

scripts/changelog-entry.sh "$tag" "$CHANGELOG" >/dev/null

stage=""
cleanup=0
if [[ -n "$artifacts_in" ]]; then
	stage="$(cd "$artifacts_in" && pwd)"
	if [[ -z "$(find "$stage" -maxdepth 1 -type f ! -name '*.rsg' ! -name '*.rsm' | head -1)" ]]; then
		echo "no artifact files in $stage" >&2
		exit 1
	fi
else
	stage="$(mktemp -d "${TMPDIR:-/tmp}/micron-rngit-${ver}.XXXXXX")"
	cleanup=1
	trap '[[ $cleanup -eq 1 ]] && rm -rf "$stage"' EXIT
	echo "Downloading GitHub release assets for $tag..."
	gh release download "$tag" -R "$GH_REPO" -D "$stage"
fi

export RNGIT_RELEASE_TAG="$tag"
export RNGIT_CHANGELOG="$CHANGELOG"
export RNGIT_REPO_ROOT="$root"
export EDITOR="$root/scripts/rngit-release-notes.sh"

echo "Creating rngit release $tag -> $RNS_REMOTE"
echo "Artifacts: $stage"
rngit release "$RNS_REMOTE" create "${tag}:${stage}"
