"""Collect real tool outputs; never execute application or fixture functions."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

HERE = Path(__file__).resolve().parent


def source_hash(root):
    digest = hashlib.sha256()
    for p in sorted(root.rglob("*")):
        if p.is_file() and (p.suffix == ".go" or p.name in ("go.mod", "go.sum")):
            if ".git" in p.parts or "vendor" in p.parts:
                continue
            digest.update(p.relative_to(root).as_posix().encode())
            digest.update(p.read_bytes())
    return digest.hexdigest()


def run(source, output, tools):
    source, output, tools = source.resolve(), output.resolve(), tools.resolve()
    output.mkdir(parents=True, exist_ok=True)
    policy = json.loads((HERE / "config/policy.json").read_text())
    env = os.environ.copy()
    env.update(GOTOOLCHAIN=policy["go_toolchain"], GOWORK="off", GOFLAGS="-mod=readonly")
    suffix = ".exe" if os.name == "nt" else ""
    def binary(name):
        local = tools / (name + suffix)
        return str(local) if local.exists() else (shutil.which(name) or name)
    commands = {
        "build": ["go", "build", "./..."],
        "vet": ["go", "vet", "-json", "./..."],
        "staticcheck": [binary("staticcheck"), "-checks=" + policy["staticcheck_checks"], "-f", "json", "./..."],
        "gosec": [binary("gosec"), "-no-fail=false", "-nosec=true", "-conf", str(HERE / "config/gosec.json"), "-fmt=json", "-out", str(output / "gosec.json"), "./..."],
    }
    manifest = {"schema": 1, "source_sha256": source_hash(source), "source": str(source),
                "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(),
                "platform": sys.platform, "policy": policy, "tools": {}}
    # Remove stale reports so failed executions cannot reuse previous success.
    for name in commands:
        (output / (name + ".json")).unlink(missing_ok=True)
    for name, command in commands.items():
        completed, rc, stdout, stderr = False, -1, "", ""
        try:
            result = subprocess.run(command, cwd=source, env=env, capture_output=True, text=True,
                                    encoding="utf-8", errors="replace", timeout=600)
            completed, rc, stdout, stderr = True, result.returncode, result.stdout, result.stderr
        except (OSError, subprocess.TimeoutExpired) as exc:
            stderr = str(exc)
        (output / (name + ".stdout.log")).write_text(stdout, encoding="utf-8")
        (output / (name + ".stderr.log")).write_text(stderr, encoding="utf-8")
        report = output / (name + ".json")
        if name in ("build", "vet"):
            report.write_text(json.dumps({"tool": name, "exit_code": rc, "stdout": stdout, "stderr": stderr}, indent=2), encoding="utf-8")
        elif name == "staticcheck":
            try:
                issues = [json.loads(line) for line in stdout.splitlines() if line.strip()]
                report.write_text(json.dumps(issues, indent=2), encoding="utf-8")
            except ValueError:
                report.write_text(stdout, encoding="utf-8")
        entry = {"command": command, "completed": completed, "exit_code": rc,
                 "sha256": hashlib.sha256(report.read_bytes()).hexdigest() if report.exists() else None}
        manifest["tools"][name] = entry
        print(name, rc, flush=True)
    manifest["source_unchanged"] = manifest["source_sha256"] == source_hash(source)
    if not manifest["source_unchanged"]:
        manifest["tools"]["build"]["completed"] = False
    (output / "manifest.json").write_text(json.dumps(manifest, indent=2), encoding="utf-8")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--tools", type=Path, default=HERE / "tools")
    args = parser.parse_args()
    run(args.source, args.output, args.tools)
    # Collection is not approval. Always invoke security_gate.py afterwards.
