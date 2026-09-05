# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

from __future__ import annotations

import ctypes
import json
import os
import sys
from pathlib import Path
from typing import Any


def _default_lib_names() -> list[str]:
    if sys.platform == "darwin":
        return ["libmicron.dylib"]
    if sys.platform == "win32":
        return ["libmicron.dll", "micron.dll"]
    return ["libmicron.so"]


def _candidate_paths() -> list[Path]:
    names = _default_lib_names()
    here = Path(__file__).resolve().parent
    roots = [
        Path(os.environ["MICRON_LIB_PATH"]) if os.environ.get("MICRON_LIB_PATH") else None,
        here,
        here / "native",
        here.parent.parent.parent / "dist",
        Path.cwd() / "dist",
    ]
    out: list[Path] = []
    for root in roots:
        if root is None:
            continue
        if root.is_file():
            out.append(root)
            continue
        for name in names:
            out.append(root / name)
    return out


def _load_lib() -> ctypes.CDLL:
    last_err: Exception | None = None
    for path in _candidate_paths():
        if not path.is_file():
            continue
        try:
            return ctypes.CDLL(str(path))
        except OSError as exc:
            last_err = exc
    raise FileNotFoundError(
        "libmicron not found; set MICRON_LIB_PATH or place the shared library "
        f"under bindings/python/micron/native or dist/. Last error: {last_err}"
    )


_lib = _load_lib()

_lib.micron_convert.argtypes = [ctypes.c_char_p, ctypes.c_int, ctypes.c_int]
_lib.micron_convert.restype = ctypes.c_void_p

_lib.micron_parse_header_tags.argtypes = [ctypes.c_char_p]
_lib.micron_parse_header_tags.restype = ctypes.c_void_p

_lib.micron_collect_form_fields.argtypes = [ctypes.c_char_p]
_lib.micron_collect_form_fields.restype = ctypes.c_void_p

_lib.micron_build_request_payload.argtypes = [
    ctypes.c_char_p,
    ctypes.c_char_p,
    ctypes.c_char_p,
]
_lib.micron_build_request_payload.restype = ctypes.c_void_p

_lib.micron_free.argtypes = [ctypes.c_void_p]
_lib.micron_free.restype = None


def _take_string(ptr: int | None) -> str:
    if not ptr:
        return ""
    try:
        raw = ctypes.cast(ptr, ctypes.c_char_p).value
        return "" if raw is None else raw.decode("utf-8")
    finally:
        _lib.micron_free(ptr)


def convert(markup: str, dark_theme: bool = True, force_monospace: bool = True) -> str:
    """Convert Micron markup to an HTML fragment."""
    return _take_string(
        _lib.micron_convert(
            markup.encode("utf-8"),
            1 if dark_theme else 0,
            1 if force_monospace else 0,
        )
    )


def parse_header_tags(markup: str) -> dict[str, str]:
    """Return page colors from leading #!fg= / #!bg= lines."""
    data = json.loads(_take_string(_lib.micron_parse_header_tags(markup.encode("utf-8"))))
    return {"fg": data.get("fg", ""), "bg": data.get("bg", "")}


def collect_form_fields(inputs: list[dict[str, Any]]) -> dict[str, str]:
    """Collect form field values from input snapshots."""
    payload = json.dumps(inputs).encode("utf-8")
    return json.loads(_take_string(_lib.micron_collect_form_fields(payload)))


def build_request_payload(
    fields: dict[str, str],
    destination: str,
    fields_spec: str,
) -> dict[str, Any]:
    """Build a Micron-style request payload."""
    return json.loads(
        _take_string(
            _lib.micron_build_request_payload(
                json.dumps(fields).encode("utf-8"),
                destination.encode("utf-8"),
                fields_spec.encode("utf-8"),
            )
        )
    )


__all__ = [
    "convert",
    "parse_header_tags",
    "collect_form_fields",
    "build_request_payload",
]
