#!/usr/bin/env python3
"""Generate Python gRPC stubs from flymind.proto.

Run from the ai-service directory using the project venv (uv):
  uv run python scripts/generate_grpc.py

Do not use system pip/python; use the project's environment so grpcio-tools
and protobuf from pyproject.toml are used. The proto imports google.protobuf
types, so the script adds the site-packages root to the protoc include path.
"""

import os
import sys
import subprocess

# ai-service root (parent of scripts/)
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
GRPC_DIR = os.path.join(ROOT, "src", "grpc")
PROTO_FILE = os.path.join(GRPC_DIR, "flymind.proto")


def main():
    if not os.path.isfile(PROTO_FILE):
        print(f"Proto not found: {PROTO_FILE}", file=sys.stderr)
        sys.exit(1)

    # Include path for google/protobuf/*.proto (site-packages)
    try:
        import google.protobuf
        site = os.path.dirname(os.path.dirname(google.protobuf.__file__))
    except ImportError:
        print(
            "Missing protobuf. Use the project venv: uv sync && uv run python scripts/generate_grpc.py",
            file=sys.stderr,
        )
        sys.exit(1)

    cmd = [
        sys.executable, "-m", "grpc_tools.protoc",
        "-I", GRPC_DIR,
        "-I", site,
        "--python_out", GRPC_DIR,
        "--grpc_python_out", GRPC_DIR,
        os.path.basename(PROTO_FILE),
    ]
    env = os.environ.copy()
    env["PYTHONPATH"] = os.path.join(ROOT, "src") + os.pathsep + env.get("PYTHONPATH", "")

    result = subprocess.run(cmd, cwd=GRPC_DIR, env=env)
    if result.returncode != 0:
        sys.exit(result.returncode)
    print("Generated flymind_pb2.py and flymind_pb2_grpc.py in src/grpc/")


if __name__ == "__main__":
    main()
