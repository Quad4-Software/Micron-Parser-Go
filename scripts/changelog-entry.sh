#!/usr/bin/env bash
# Extract one CHANGELOG.md section for a version tag.
# Accepts v1.2.3 or 1.2.3. Prints the matching ## [ver] section.

set -euo pipefail

if [[ $# -lt 1 ]]; then
	echo "usage: $0 <version|tag> [changelog]" >&2
	exit 2
fi

raw="$1"
changelog="${2:-CHANGELOG.md}"
ver="${raw#v}"

if [[ ! -f "$changelog" ]]; then
	echo "changelog not found: $changelog" >&2
	exit 1
fi

out="$(awk -v ver="$ver" '
	BEGIN { pat = "^## \\[" ver "\\]" }
	$0 ~ pat { printing = 1; found = 1; print; next }
	printing && $0 ~ /^## \[/ { exit }
	printing { print }
	END { if (!found) exit 1 }
' "$changelog")" || {
	echo "no changelog entry for [$ver] in $changelog" >&2
	exit 1
}

printf '%s\n' "$out"
