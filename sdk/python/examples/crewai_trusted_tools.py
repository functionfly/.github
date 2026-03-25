#!/usr/bin/env python3
"""
CrewAI trusted tools example for FunctionFly.
"""

import os

from flypy import AgentClient, CrewAIAdapter, TrustPolicy


def main() -> None:
    api_base = os.environ.get("FUNCTIONFLY_API_BASE", "http://localhost:8080")
    api_key = os.environ.get("FUNCTIONFLY_API_KEY")

    client = AgentClient(api_base=api_base, api_key=api_key)
    adapter = CrewAIAdapter(client)
    policy = TrustPolicy(
        min_trust_score=80,
        require_verified=True,
        capabilities_deny=["secrets_read"],
    )

    tools = adapter.build_tools(policy=policy, query="text", limit=5)
    print(f"Loaded {len(tools)} trusted CrewAI tools")
    if not tools:
        print("No tools matched the trust policy")
        return

    tool = tools[0]
    print("Tool:", tool["name"])
    result = tool["func"](text="CrewAI + FunctionFly")
    print("Result:", result)
    print("Execution ID:", result.get("execution_id"))


if __name__ == "__main__":
    main()
