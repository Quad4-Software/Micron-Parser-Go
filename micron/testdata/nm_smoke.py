# SPDX-License-Identifier: 0BSD
"""Smoke-test that the latest pypi nomadnet MicronParser accepts folding markup."""
from __future__ import annotations

import sys


def main() -> int:
    import nomadnet.ui.textui.MicronParser as MP

    # Avoid needing a running UI for palette lookup.
    MP.SELECTED_STYLES = MP.STYLES_DARK
    MP.ensure_selected_styles = lambda: None
    MP.make_style = lambda state: "plain"

    state = MP.default_state("ddd", "default")
    samples = [
        "> plain heading",
        "`+> open folding heading",
        "`-> closed folding heading",
        "#!fold v >",
    ]
    for s in samples:
        out = MP.parse_line(s, state, None)
        if out is None and s.startswith("#!fold"):
            # fold directive is a comment, None is expected.
            continue
        if not out:
            print(f"FAIL: nomadnet did not return output for {s!r}", file=sys.stderr)
            return 1
        print(f"ok: {s!r}")
    print("nomadnet-smoke ok")
    return 0


if __name__ == "__main__":
    sys.exit(main())
