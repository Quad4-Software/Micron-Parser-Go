"""
sync_nomadnet_guide.py extracts TOPIC_MARKUP from the upstream NomadNet
Guide.py without importing it (so urwid / RNS / nomadnet need not be
installed). Result is written to nomadnet_guide_official.mu and mirrored to
the playground/bench seed copies.

Usage:
    python3 sync_nomadnet_guide.py [path/to/NomadNet]

If no path is given, NOMADNET_DIR is used, falling back to
/run/media/user1/projects/Reticulum/NomadNet.

This file is intentionally a small, stdlib-only helper. It is not invoked by
the Go test suite at run time; it is run by hand to refresh the committed
snapshot when NomadNet upstream changes.
"""

import argparse
import os
import sys
from pathlib import Path


DEFAULT_NOMADNET_DIR = "/run/media/user1/projects/Reticulum/NomadNet"


def extract_topic_markup(guide_path: Path) -> str:
    """Evaluate TOPIC_MARKUP the same way Guide.py builds it at import time.

    NomadNet assigns the main body, then appends an escaped self-copy inside a
    literal block (TOPIC_MARKUP.replace), then Closing Remarks. String-slicing
    the assignment alone drops that self-embed.
    """
    source = guide_path.read_text(encoding="utf-8")
    start = source.find("TOPIC_MARKUP =")
    if start < 0:
        raise SystemExit("TOPIC_MARKUP assignment not found in Guide.py")
    end = source.find("\nTOPICS =", start)
    if end < 0:
        raise SystemExit("TOPICS dict not found after TOPIC_MARKUP in Guide.py")
    block = source[start:end]
    ns: dict[str, object] = {}
    exec(block, {"__builtins__": {}}, ns)
    markup = ns.get("TOPIC_MARKUP")
    if not isinstance(markup, str) or markup == "":
        raise SystemExit("TOPIC_MARKUP did not evaluate to a non-empty string")
    return markup


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
    data = markup.encode("utf-8")
    testdata = Path(__file__).resolve().parent
    repo = testdata.parent.parent
    outputs = [
        testdata / "nomadnet_guide_official.mu",
        testdata / "nomadnet_guide.mu",
        repo / "web" / "static" / "data" / "nomadnet_guide.mu",
    ]
    for out_path in outputs:
        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_bytes(data)
        print(f"wrote {out_path} ({len(data)} bytes)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
