#!/usr/bin/env python3
"""
LangChain trusted tools example for FunctionFly.
"""

import os

from flypy import AgentClient, LangChainAdapter, TrustPolicy


def main() -> None:
    api_base = os.environ.get("FUNCTIONFLY_API_BASE", "http://localhost:8080")
    api_key = os.environ.get("FUNCTIONFLY_API_KEY")

    client = AgentClient(api_base=api_base, api_key=api_key)
    adapter = LangChainAdapter(client)
    policy = TrustPolicy(
        min_trust_score=80,
        require_verified=True,
        required_trust_levels=["high"],
    )

    tools = adapter.build_tools(
        policy=policy,
        query="text transform",
        category=None,
        limit=5,
    )
    print(f"Loaded {len(tools)} trusted LangChain tools")

    if not tools:
        print("No tools matched the trust policy")
        return

    first_tool = tools[0]
    if isinstance(first_tool, dict):
        result = first_tool["callable"](text="hello functionfly")
        print("Tool:", first_tool["name"])
    else:
        # LangChain StructuredTool path
        result = first_tool.invoke({"text": "hello functionfly"})
        print("Tool:", getattr(first_tool, "name", "unknown"))

    print("Result:", result)


if __name__ == "__main__":
    main()
