#!/usr/bin/env python3

import json
import pathlib
import sys


def load_entries(path: pathlib.Path) -> list[str]:
    entries: list[str] = []
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        entries.append(line)
    return entries


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: ci_matrix.py <list-file>", file=sys.stderr)
        return 1

    path = pathlib.Path(sys.argv[1])
    if not path.is_file():
        print(f"list file not found: {path}", file=sys.stderr)
        return 1

    entries = load_entries(path)
    if not entries:
        print(f"no entries found in {path}", file=sys.stderr)
        return 1

    print(json.dumps(entries, separators=(",", ":")))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
