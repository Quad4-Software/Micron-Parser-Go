#!/usr/bin/env python3
# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD
"""
python_oracle.py emits a semantic fingerprint of Micron markup for differential
tests against micron-parser-go.

Authority: NomadNet Python (MicronParser.py / Guide.py) defines the dialect.
Stock MicronParser.markup_to_attrmaps cannot run headlessly without urwid,
NomadNetworkApp, RNS, and MarkdownToMicron. This oracle therefore does NOT call
the Urwid path. It emits a token-style fingerprint aligned with MicronParser
line/tag semantics (sections, links, fields, partials, headers) via a stdlib
scanner so CI works without NomadNet installed.

Optional: if the nomadnet package imports, nomadnet_installed is true in the
JSON (presence check only). The fingerprint method remains semantic_scan.

Usage:
    python3 python_oracle.py < markup.mu
    python3 python_oracle.py path/to/file.mu

Output JSON:
    {
      "authority": "stdlib",
      "method": "semantic_scan",
      "nomadnet_installed": true|false,
      "sections": [depth, ...],
      "links": [{"url": "...", "label": "..."}],
      "fields": [{"name": "...", "kind": "..."}],
      "partials": [{"destination": "..."}],
      "headers": {"fg": "...", "bg": "..."}
    }
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path


HEADER_FG = re.compile(r"^#!fg=([0-9a-fA-F]{3}|[0-9a-fA-F]{6})\s*$")
HEADER_BG = re.compile(r"^#!bg=([0-9a-fA-F]{3}|[0-9a-fA-F]{6})\s*$")
SECTION = re.compile(r"^(>+)")
LINK = re.compile(r"`\[([^\]`]*)`([^\]`]*)(?:`([^\]`]*))?\]")
LINK_BARE = re.compile(r"`\[([^\]`]+)\]")
FIELD = re.compile(r"`<([^>`]*)`([^>]*)>")
PARTIAL = re.compile(r"`\{([^}]*)\}")


def nomadnet_installed() -> bool:
    try:
        import nomadnet  # noqa: F401

        return True
    except Exception:
        return False


def fingerprint(markup: str) -> dict:
    sections: list[int] = []
    links: list[dict] = []
    fields: list[dict] = []
    partials: list[dict] = []
    headers = {"fg": "", "bg": ""}
    literal = False

    for raw in markup.split("\n"):
        line = raw.rstrip("\r")
        t = line.strip()
        if t == "`=":
            literal = not literal
            continue
        if not literal and t.startswith("#!"):
            m = HEADER_FG.match(t)
            if m:
                headers["fg"] = m.group(1)
                continue
            m = HEADER_BG.match(t)
            if m:
                headers["bg"] = m.group(1)
                continue
        if literal or (t.startswith("#") and not t.startswith("#!")):
            continue
        if t.startswith(">"):
            m = SECTION.match(t)
            if m:
                depth = min(len(m.group(1)), 16)
                rest = t[len(m.group(1)) :].strip()
                if rest:
                    sections.append(depth)
            continue
        if t.startswith("<"):
            continue
        for m in LINK.finditer(line):
            label, url = m.group(1), m.group(2)
            if not label:
                label = url
            links.append({"url": url, "label": label})
        if "`[" in line and not any(True for _ in LINK.finditer(line)):
            for m in LINK_BARE.finditer(line):
                url = m.group(1)
                links.append({"url": url, "label": url})
        for m in FIELD.finditer(line):
            spec, data = m.group(1), m.group(2)
            kind = "text"
            name = spec
            if "|" in spec:
                flags, rest = spec.split("|", 1)
                name = rest.split("|", 1)[0]
                if "^" in flags:
                    kind = "radio"
                elif "?" in flags:
                    kind = "checkbox"
                elif "!" in flags:
                    kind = "masked"
            fields.append({"name": name, "kind": kind, "data": data})
        for m in PARTIAL.finditer(line):
            dest = m.group(1).split("`", 1)[0].strip()
            if dest:
                partials.append({"destination": dest})

    return {
        "authority": "stdlib",
        "method": "semantic_scan",
        "nomadnet_installed": nomadnet_installed(),
        "sections": sections,
        "links": links,
        "fields": fields,
        "partials": partials,
        "headers": headers,
    }


def main() -> int:
    if len(sys.argv) > 1:
        markup = Path(sys.argv[1]).read_text(encoding="utf-8")
    else:
        markup = sys.stdin.read()
    json.dump(fingerprint(markup), sys.stdout, separators=(",", ":"), ensure_ascii=False)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
