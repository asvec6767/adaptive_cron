import copy
import hashlib
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

HERE = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("gate", HERE / "security_gate.py")
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)


class GateTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.manifest = {"schema": 1, "source_sha256": "a" * 64, "tools": {}}
        self.reports = {
            "build": {"tool": "build", "exit_code": 0},
            "vet": {"tool": "vet", "exit_code": 0},
            "staticcheck": [],
            "gosec": {"Issues": [], "Golang errors": {}, "Stats": {"files": 1}},
        }
        for tool in gate.TOOLS:
            self.write(tool)

    def write(self, tool, rc=0):
        data = json.dumps(self.reports[tool]).encode()
        (self.root / (tool + ".json")).write_bytes(data)
        self.manifest["tools"][tool] = {"completed": True, "exit_code": rc, "command": [tool],
                                         "sha256": hashlib.sha256(data).hexdigest()}

    def result(self):
        (self.root / "manifest.json").write_text(json.dumps(self.manifest))
        return gate.evaluate(self.root, HERE / "config/policy.json")

    def test_clean(self):
        self.assertTrue(self.result()["passed"])

    def test_missing_each_report(self):
        for tool in gate.TOOLS:
            with self.subTest(tool=tool):
                self.write(tool)
                p = self.root / (tool + ".json")
                data = p.read_bytes()
                p.unlink()
                self.assertFalse(self.result()["passed"])
                p.write_bytes(data)

    def test_corrupt_with_valid_hash(self):
        for tool in gate.TOOLS:
            with self.subTest(tool=tool):
                (self.root / (tool + ".json")).write_bytes(b"{")
                self.manifest["tools"][tool]["sha256"] = hashlib.sha256(b"{").hexdigest()
                self.assertFalse(self.result()["passed"])
                self.write(tool)

    def test_hash_mismatch(self):
        (self.root / "gosec.json").write_text("{}")
        self.assertFalse(self.result()["passed"])

    def test_missing_manifest(self):
        self.assertFalse(gate.evaluate(self.root, HERE / "config/policy.json")["passed"])

    def test_empty_schema(self):
        self.reports["gosec"] = {}
        self.write("gosec")
        self.assertFalse(self.result()["passed"])

    def test_analyzer_failure_codes(self):
        for tool in gate.TOOLS:
            for rc in (-1, 1, 2, 127):
                with self.subTest(tool=tool, rc=rc):
                    self.write(tool, rc)
                    self.assertFalse(self.result()["passed"])
            self.write(tool)

    def test_incomplete_execution(self):
        self.manifest["tools"]["gosec"]["completed"] = False
        self.assertFalse(self.result()["passed"])

    def test_missing_tool(self):
        del self.manifest["tools"]["vet"]
        self.assertFalse(self.result()["passed"])

    def test_package_error(self):
        self.reports["gosec"]["Golang errors"] = {"package": [{"error": "load failed"}]}
        self.write("gosec")
        self.assertFalse(self.result()["passed"])

    def test_no_files(self):
        self.reports["gosec"]["Stats"]["files"] = 0
        self.write("gosec")
        self.assertFalse(self.result()["passed"])

    def test_severities(self):
        for severity, expected in (("HIGH", False), ("MEDIUM", False), ("LOW", True), ("NEW", False)):
            self.reports["gosec"]["Issues"] = [{"severity": severity, "confidence": "HIGH", "rule_id": "G204",
                                                "file": "fixture.go", "line": "10", "details": "test"}]
            self.write("gosec", 1)
            self.assertEqual(expected, self.result()["passed"])

    def test_staticcheck_diagnostic(self):
        self.reports["staticcheck"] = [{"code": "SA5011", "location": {"file": "x.go", "line": 2},
                                         "message": "nil dereference", "severity": "error"}]
        self.write("staticcheck", 1)
        self.assertFalse(self.result()["passed"])

    def test_invalid_policy(self):
        p = self.root / "policy.json"
        p.write_text('{"schema": 1, "block_severities": []}')
        self.assertFalse(gate.evaluate(self.root, p)["passed"])

    def test_duplicate_json_key(self):
        data = b'{"Issues": [], "Issues": [], "Golang errors": {}, "Stats": {"files": 1}}'
        (self.root / "gosec.json").write_bytes(data)
        self.manifest["tools"]["gosec"]["sha256"] = hashlib.sha256(data).hexdigest()
        self.assertFalse(self.result()["passed"])


if __name__ == "__main__":
    unittest.main()
