"""Reproduce both real fixture analyses and test the fail-closed gate."""
import argparse
import datetime
import io
import json
from pathlib import Path
import subprocess
import sys
import unittest

HERE = Path(__file__).resolve().parent


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--reuse", action="store_true", help="validate existing reports without rescanning")
    args = parser.parse_args()
    out = HERE / "artifacts"
    out.mkdir(exist_ok=True)
    stream = io.StringIO()
    result = unittest.TextTestRunner(stream=stream, verbosity=2).run(unittest.defaultTestLoader.discover(str(HERE / "tests")))
    (out / "gate-tests.log").write_text(stream.getvalue(), encoding="utf-8")
    summary = {"timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(),
               "unit_tests": {"run": result.testsRun, "failures": len(result.failures), "errors": len(result.errors),
                              "passed": result.wasSuccessful()}, "fixtures": {}}
    for name, expected in (("vulnerable", 1), ("fixed", 0)):
        dest = out / name
        if not args.reuse:
            subprocess.run([sys.executable, str(HERE / "run_checks.py"), "--source", str(HERE / "fixtures" / name),
                            "--output", str(dest)], check=True)
        checked = subprocess.run([sys.executable, str(HERE / "security_gate.py"), str(dest)], capture_output=True,
                                 encoding="utf-8", errors="replace")
        (dest / "gate.log").write_text(checked.stdout + checked.stderr, encoding="utf-8")
        decision = json.loads((dest / "gate.json").read_text(encoding="utf-8"))
        rules = sorted(i["rule_id"] for i in decision["findings"] if i["tool"] == "gosec")
        expected_rules = ["G204", "G402"] if name == "vulnerable" else []
        ok = checked.returncode == expected and rules == expected_rules
        summary["fixtures"][name] = {"gate_exit": checked.returncode, "expected_exit": expected,
                                      "rules": rules, "expected_rules": expected_rules, "passed": ok}
    summary["passed"] = result.wasSuccessful() and all(x["passed"] for x in summary["fixtures"].values())
    (out / "validation.json").write_text(json.dumps(summary, indent=2), encoding="utf-8")
    print(json.dumps(summary, indent=2))
    return 0 if summary["passed"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
