"""Fail-closed validation of reports produced by run_checks.py (stdlib only)."""
import argparse
import hashlib
import json
from pathlib import Path

TOOLS = ("build", "vet", "staticcheck", "gosec")


def load(path):
    def unique(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise ValueError("duplicate JSON key: " + key)
            result[key] = value
        return result
    return json.loads(Path(path).read_text(encoding="utf-8"), object_pairs_hook=unique)


def evaluate(directory, policy_path):
    root = Path(directory)
    reasons, findings = [], []
    try:
        policy = load(policy_path)
        if policy["schema"] != 1 or policy["block_severities"] != ["HIGH", "MEDIUM"]:
            raise ValueError("unsupported policy")
        manifest = load(root / "manifest.json")
        if manifest["schema"] != 1 or set(manifest["tools"]) != set(TOOLS):
            raise ValueError("incomplete manifest")
        if not isinstance(manifest["source_sha256"], str) or len(manifest["source_sha256"]) != 64:
            raise ValueError("missing source identity")
        for tool in TOOLS:
            entry = manifest["tools"][tool]
            rc = entry["exit_code"]
            if (type(rc) is not int or entry["completed"] is not True
                    or not isinstance(entry["command"], list) or not entry["command"]):
                raise ValueError(tool + ": incomplete execution")
            path = root / (tool + ".json")
            if hashlib.sha256(path.read_bytes()).hexdigest() != entry["sha256"]:
                raise ValueError(tool + ": report hash mismatch")
            report = load(path)
            if tool in ("build", "vet"):
                if report["tool"] != tool or report["exit_code"] != rc or rc != 0:
                    reasons.append(tool + ": compilation / vet failed")
            elif tool == "staticcheck":
                if not isinstance(report, list) or rc not in (0, 1) or (rc == 1 and not report):
                    raise ValueError("staticcheck: analyzer failure")
                for issue in report:
                    if not all(key in issue for key in ("code", "location", "message", "severity")):
                        raise ValueError("invalid staticcheck finding")
                    if not isinstance(issue["location"], dict) or not issue["code"]:
                        raise ValueError("invalid staticcheck location/code")
                    findings.append({"tool": tool, **issue})
                if report:
                    reasons.append("staticcheck: diagnostics present")
            else:
                if not isinstance(report, dict) or not all(k in report for k in ("Issues", "Golang errors", "Stats")):
                    raise ValueError("invalid gosec schema")
                issues = report["Issues"]
                if not isinstance(issues, list) or not isinstance(report["Stats"], dict):
                    raise ValueError("invalid gosec issues/stats")
                if type(report["Stats"].get("files")) is not int or report["Stats"]["files"] < 1:
                    raise ValueError("gosec: no files analyzed")
                if report["Golang errors"] not in (None, {}):
                    raise ValueError("gosec: package loading errors")
                if rc not in (0, 1) or (rc == 1 and not issues):
                    raise ValueError("gosec: analyzer failure")
                for issue in issues:
                    if not all(k in issue for k in ("severity", "confidence", "rule_id", "file", "line", "details")):
                        raise ValueError("invalid gosec finding")
                    if issue["severity"] not in ("HIGH", "MEDIUM", "LOW"):
                        raise ValueError("unknown gosec severity")
                    findings.append({"tool": tool, **issue})
                    if issue["severity"] in policy["block_severities"]:
                        reasons.append("gosec: " + issue["rule_id"] + " " + issue["severity"])
    except (OSError, ValueError, KeyError, TypeError, AttributeError) as exc:
        reasons.append("invalid or missing evidence: " + str(exc))
    return {"schema": 1, "passed": not reasons, "reasons": reasons, "findings": findings}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("directory", type=Path)
    parser.add_argument("--policy", type=Path, default=Path(__file__).parent / "config/policy.json")
    args = parser.parse_args()
    result = evaluate(args.directory, args.policy)
    args.directory.mkdir(parents=True, exist_ok=True)
    text = json.dumps(result, ensure_ascii=False, indent=2)
    (args.directory / "gate.json").write_text(text, encoding="utf-8")
    print(text)
    return 0 if result["passed"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
