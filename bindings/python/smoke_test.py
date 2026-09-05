# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT))

from micron import convert, parse_header_tags, collect_form_fields, build_request_payload


def main() -> None:
    html = convert("> Title\n\nHello <world> & `*bold`*.\n", dark_theme=True, force_monospace=False)
    assert "Hello" in html and "&lt;world&gt;" in html
    assert "bold" in html

    colors = parse_header_tags("#!fg=ccc\n#!bg=222\n\nBody\n")
    assert colors["fg"] == "ccc"
    assert colors["bg"] == "222"

    fields = collect_form_fields(
        [
            {"type": "text", "name": "user", "value": "alice", "checked": False},
            {"type": "checkbox", "name": "opts", "value": "1", "checked": True},
        ]
    )
    assert fields["user"] == "alice"
    assert fields["opts"] == "1"

    payload = build_request_payload(fields, "/page`x=1", "user|opts")
    assert payload["destination"] == "/page"
    assert payload["fields"]["user"] == "alice"
    print("python smoke ok")


if __name__ == "__main__":
    main()
