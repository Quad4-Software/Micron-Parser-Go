"""
sync_nomadnet_guide.py extracts TOPIC_MARKUP from the upstream NomadNet
Guide.py without importing it (so urwid / RNS / nomadnet need not be
installed). Result is written to nomadnet_guide_official.mu next to this
script.

Usage:
    python3 sync_nomadnet_guide.py [path/to/NomadNet]

If no path is given, NOMADNET_DIR is used, falling back to
/run/media/user1/projects/Reticulum/NomadNet.

This file is intentionally a small, stdlib-only helper. It is not invoked by
the Go test suite at run time; it is run by hand to refresh the committed
snapshot when NomadNet upstream changes.
"""

import argparse
import ast
import os
import sys
from pathlib import Path


DEFAULT_NOMADNET_DIR = "/run/media/user1/projects/Reticulum/NomadNet"


def extract_topic_markup(guide_path: Path) -> str:
    source = guide_path.read_text(encoding="utf-8")
    tree = ast.parse(source, filename=str(guide_path))
    pieces: list[str] = []
    for node in ast.walk(tree):
        if isinstance(node, ast.Assign):
            for target in node.targets:
                if isinstance(target, ast.Name) and target.id == "TOPIC_MARKUP":
                    if not isinstance(node.value, ast.Constant) or not isinstance(node.value.value, str):
                        raise SystemExit("expected TOPIC_MARKUP to be a string literal assignment")
                    pieces.append(node.value.value)
        elif isinstance(node, ast.AugAssign):
            if isinstance(node.target, ast.Name) and node.target.id == "TOPIC_MARKUP":
                if isinstance(node.value, ast.Constant) and isinstance(node.value.value, str):
                    pieces.append(node.value.value)
                else:
                    pieces.append("")
    if not pieces:
        raise SystemExit("TOPIC_MARKUP not found in Guide.py")
    return "".join(pieces)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "nomadnet_dir",
        nargs="?",
        default=os.environ.get("NOMADNET_DIR", DEFAULT_NOMADNET_DIR),
        help="path to a NomadNet checkout",
    )
    args = parser.parse_args()

    guide_path = Path(args.nomadnet_dir) / "nomadnet" / "ui" / "textui" / "Guide.py"
    if not guide_path.is_file():
        print(f"Guide.py not found at {guide_path}", file=sys.stderr)
        return 1

    markup = extract_topic_markup(guide_path)
    out_path = Path(__file__).with_name("nomadnet_guide_official.mu")
    out_path.write_text(markup, encoding="utf-8")
    print(f"wrote {out_path} ({len(markup.encode('utf-8'))} bytes)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
