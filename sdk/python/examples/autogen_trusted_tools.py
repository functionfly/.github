#!/usr/bin/env python3
"""
AutoGen trusted tools example for FunctionFly.
"""

import os

from flypy import AgentClient, AutoGenAdapter, TrustPolicy


def main() -> None:
    api_base = os.environ.get("FUNCTIONFLY_API_BASE", "http://localhost:8080")
    api_key = os.environ.get("FUNCTIONFLY_API_KEY")

    client = AgentClient(api_base=api_base, api_key=api_key)
    adapter = AutoGenAdapter(client)
    policy = TrustPolicy(min_trust_score=80, require_verified=True)

    tools = adapter.build_tools(policy=policy, query="json", limit=5)
    print(f"Loaded {len(tools)} trusted AutoGen tools")
    if not tools:
        print("No tools matched the trust policy")
        return

    tool = tools[0]
    print("Tool:", tool["name"])
    # Exact args depend on selected FunctionFly function schema.
    result = tool["function"](text='{"hello":"world"}')
    print("Result:", result)
    print("Policy hash:", result.get("metadata", {}).get("policy_hash"))


if __name__ == "__main__":
    main()
