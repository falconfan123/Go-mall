#!/usr/bin/env python3

"""Render go test -json output into JUnit XML, a text summary, and HTML."""

from __future__ import annotations

import argparse
import html
import json
import sys
import textwrap
import xml.etree.ElementTree as ET
from collections import OrderedDict
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Iterable


@dataclass
class TestCase:
    name: str
    status: str = "running"
    duration: float = 0.0
    output: list[str] = field(default_factory=list)

    @property
    def failure_message(self) -> str:
        for line in reversed(self.output):
            stripped = line.strip()
            if stripped:
                return stripped
        return "test failed"


@dataclass
class PackageResult:
    name: str
    status: str = "unknown"
    duration: float = 0.0
    output: list[str] = field(default_factory=list)
    tests: OrderedDict[str, TestCase] = field(default_factory=OrderedDict)

    def testcase(self, name: str) -> TestCase:
        if name not in self.tests:
            self.tests[name] = TestCase(name=name)
        return self.tests[name]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--suite", required=True)
    parser.add_argument("--title", required=True)
    parser.add_argument("--out-dir", required=True)
    parser.add_argument("--input", default="-")
    return parser.parse_args()


def load_events(stream: Iterable[str]) -> list[dict]:
    events: list[dict] = []
    for raw in stream:
        raw = raw.strip()
        if not raw:
            continue
        try:
            events.append(json.loads(raw))
        except json.JSONDecodeError:
            events.append({"Action": "output", "Output": raw, "_raw": True})
    return events


def collect_packages(events: list[dict]) -> OrderedDict[str, PackageResult]:
    packages: OrderedDict[str, PackageResult] = OrderedDict()

    for event in events:
        package_name = event.get("Package") or ""
        test_name = event.get("Test") or ""

        if package_name and package_name not in packages:
            packages[package_name] = PackageResult(name=package_name)
        package = packages.get(package_name)

        action = event.get("Action")
        output = event.get("Output")
        elapsed = event.get("Elapsed")

        if action == "run":
            if package and test_name:
                package.testcase(test_name)
            continue

        if action == "output" and output is not None:
            if package and test_name:
                package.testcase(test_name).output.append(output)
            elif package:
                package.output.append(output)
            continue

        if action in {"pass", "fail", "skip"}:
            if package and test_name:
                testcase = package.testcase(test_name)
                testcase.status = action
                if isinstance(elapsed, (int, float)):
                    testcase.duration = float(elapsed)
                if action == "fail" and not testcase.output:
                    testcase.output.append(output or f"{test_name} failed")
            elif package:
                package.status = action
                if isinstance(elapsed, (int, float)):
                    package.duration = float(elapsed)
            continue

    for package in packages.values():
        if package.status == "unknown":
            if any(case.status == "fail" for case in package.tests.values()):
                package.status = "fail"
            elif any(case.status == "skip" for case in package.tests.values()):
                package.status = "skip"
            else:
                package.status = "pass" if package.tests else "unknown"

        if package.duration <= 0:
            package.duration = sum(case.duration for case in package.tests.values())

        if package.status == "fail" and not package.tests:
            package.tests["package setup"] = TestCase(
                name="package setup",
                status="fail",
                output=package.output[:] or ["package failed"],
            )

    return packages


def build_summary(packages: OrderedDict[str, PackageResult]) -> dict[str, object]:
    tests = 0
    failures = 0
    skipped = 0
    duration = 0.0
    failed_cases: list[tuple[str, TestCase]] = []

    for package in packages.values():
        duration += package.duration
        for case in package.tests.values():
            tests += 1
            if case.status == "fail":
                failures += 1
                failed_cases.append((package.name, case))
            elif case.status == "skip":
                skipped += 1

    return {
        "tests": tests,
        "failures": failures,
        "skipped": skipped,
        "duration": duration,
        "failed_cases": failed_cases,
    }


def write_junit(out_dir: Path, suite: str, title: str, packages: OrderedDict[str, PackageResult], summary: dict[str, object]) -> None:
    testsuites = ET.Element(
        "testsuites",
        {
            "name": suite,
            "tests": str(summary["tests"]),
            "failures": str(summary["failures"]),
            "skipped": str(summary["skipped"]),
            "time": f'{summary["duration"]:.3f}',
        },
    )

    for package in packages.values():
        package_cases = list(package.tests.values())
        suite_elem = ET.SubElement(
            testsuites,
            "testsuite",
            {
                "name": package.name,
                "tests": str(len(package_cases)),
                "failures": str(sum(1 for case in package_cases if case.status == "fail")),
                "skipped": str(sum(1 for case in package_cases if case.status == "skip")),
                "time": f"{package.duration:.3f}",
                "timestamp": datetime.now(timezone.utc).isoformat(),
            },
        )
        ET.SubElement(suite_elem, "properties")
        for case in package_cases:
            case_elem = ET.SubElement(
                suite_elem,
                "testcase",
                {
                    "classname": package.name,
                    "name": case.name,
                    "time": f"{case.duration:.3f}",
                },
            )
            if case.status == "fail":
                failure = ET.SubElement(case_elem, "failure", {"message": case.failure_message})
                failure.text = "\n".join(case.output) or case.failure_message
            elif case.status == "skip":
                ET.SubElement(case_elem, "skipped")
            if case.output:
                out = ET.SubElement(case_elem, "system-out")
                out.text = "\n".join(case.output)

    tree = ET.ElementTree(testsuites)
    tree.write(out_dir / "junit.xml", encoding="utf-8", xml_declaration=True)


def write_summary(out_dir: Path, title: str, packages: OrderedDict[str, PackageResult], summary: dict[str, object]) -> None:
    lines = [
        f"{title}",
        f"packages: {len(packages)}",
        f"tests: {summary['tests']}",
        f"failures: {summary['failures']}",
        f"skipped: {summary['skipped']}",
        f"duration: {summary['duration']:.3f}s",
        "",
    ]

    failed_cases = summary["failed_cases"]
    if failed_cases:
        lines.append("failed tests:")
        for package, case in failed_cases:
            lines.append(f"- {package}/{case.name}: {case.failure_message}")
    else:
        lines.append("failed tests: none")

    (out_dir / "summary.txt").write_text("\n".join(lines) + "\n", encoding="utf-8")


def render_html(title: str, packages: OrderedDict[str, PackageResult], summary: dict[str, object]) -> str:
    failed_cases = summary["failed_cases"]
    rows = []
    for package in packages.values():
        status = package.status
        if any(case.status == "fail" for case in package.tests.values()):
            status = "fail"
        elif all(case.status == "skip" for case in package.tests.values()) and package.tests:
            status = "skip"

        rows.append(
            "<tr>"
            f"<td>{html.escape(package.name)}</td>"
            f"<td class='status {status}'>{status}</td>"
            f"<td>{len(package.tests)}</td>"
            f"<td>{sum(1 for case in package.tests.values() if case.status == 'fail')}</td>"
            f"<td>{sum(1 for case in package.tests.values() if case.status == 'skip')}</td>"
            f"<td>{package.duration:.3f}s</td>"
            "</tr>"
        )

    failure_blocks = []
    if failed_cases:
        for package, case in failed_cases:
            failure_output = html.escape("\n".join(case.output) or case.failure_message)
            failure_blocks.append(
                "<details open>"
                f"<summary>{html.escape(package)}/{html.escape(case.name)}</summary>"
                f"<pre>{failure_output}</pre>"
                "</details>"
            )
    else:
        failure_blocks.append("<p>No failed tests.</p>")

    return textwrap.dedent(
        f"""
        <!doctype html>
        <html lang="en">
        <head>
          <meta charset="utf-8">
          <meta name="viewport" content="width=device-width, initial-scale=1">
          <title>{html.escape(title)}</title>
          <style>
            body {{ font-family: system-ui, sans-serif; margin: 24px; color: #111; background: #fafafa; }}
            .grid {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 12px; margin-bottom: 20px; }}
            .card {{ background: #fff; border: 1px solid #ddd; border-radius: 10px; padding: 12px 14px; }}
            .label {{ color: #666; font-size: 12px; text-transform: uppercase; letter-spacing: .04em; }}
            .value {{ font-size: 28px; font-weight: 700; margin-top: 4px; }}
            table {{ width: 100%; border-collapse: collapse; background: #fff; border: 1px solid #ddd; }}
            th, td {{ padding: 10px 12px; border-bottom: 1px solid #eee; text-align: left; vertical-align: top; }}
            th {{ background: #f6f6f6; }}
            .status.pass {{ color: #0a7c2f; }}
            .status.fail {{ color: #b42318; }}
            .status.skip {{ color: #9a6700; }}
            details {{ background: #fff; border: 1px solid #ddd; border-radius: 10px; padding: 10px 12px; margin: 10px 0; }}
            pre {{ white-space: pre-wrap; word-break: break-word; margin: 10px 0 0; }}
          </style>
        </head>
        <body>
          <h1>{html.escape(title)}</h1>
          <div class="grid">
            <div class="card"><div class="label">Packages</div><div class="value">{len(packages)}</div></div>
            <div class="card"><div class="label">Tests</div><div class="value">{summary['tests']}</div></div>
            <div class="card"><div class="label">Failures</div><div class="value">{summary['failures']}</div></div>
            <div class="card"><div class="label">Duration</div><div class="value">{summary['duration']:.3f}s</div></div>
          </div>
          <table>
            <thead>
              <tr><th>Package</th><th>Status</th><th>Tests</th><th>Failed</th><th>Skipped</th><th>Duration</th></tr>
            </thead>
            <tbody>
              {''.join(rows)}
            </tbody>
          </table>
          <h2>Failures</h2>
          {''.join(failure_blocks)}
        </body>
        </html>
        """
    ).strip()


def main() -> int:
    args = parse_args()
    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    if args.input == "-":
        events = load_events(sys.stdin)
    else:
        events = load_events(Path(args.input).read_text(encoding="utf-8").splitlines())

    packages = collect_packages(events)
    summary = build_summary(packages)

    write_junit(out_dir, args.suite, args.title, packages, summary)
    write_summary(out_dir, args.title, packages, summary)
    (out_dir / "index.html").write_text(render_html(args.title, packages, summary), encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
