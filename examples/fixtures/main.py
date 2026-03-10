async def handler(input: str) -> str:
    return input.lower().replace(" ", "-")