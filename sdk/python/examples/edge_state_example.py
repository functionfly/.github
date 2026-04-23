"""
Edge State Example - Shopping Cart at the Edge

This example demonstrates how to use the EdgeStateClient from within
a FlyPy function running at the edge (WASM runtime).

The edge state provides low-latency state access for edge functions,
bypassing HTTP API calls and using direct WASM host function calls.
"""

import flypy
from flypy import edge_state


@flypy.function(
    name="shopping-cart-edge",
    description="Shopping cart with edge state storage",
    edge_state_enabled=True,
)
def shopping_cart_handler(event: dict) -> dict:
    """
    Handle shopping cart operations using edge state.

    Args:
        event: Operation details with:
            - action: "get", "add", "remove", "clear", "checkout"
            - user_id: The user identifier
            - item: Item details (for add/remove)

    Returns:
        Cart state or operation result
    """
    action = event.get("action", "get")
    user_id = event.get("user_id")

    if not user_id:
        return {"error": "user_id is required"}

    # Initialize edge state client
    # The fabric_id is auto-detected from environment in edge runtime
    state = edge_state.EdgeStateClient()

    # Build state key for this user's cart
    cart_key = f"carts/{user_id}"

    if action == "get":
        # Get current cart
        cart = state.get(cart_key, default={"items": [], "total": 0})
        return {"cart": cart, "user_id": user_id}

    elif action == "add":
        # Add item to cart
        item = event.get("item")
        if not item:
            return {"error": "item is required for add action"}

        # Get current cart or create new
        cart = state.get(cart_key, default={"items": [], "total": 0})

        # Add item
        cart["items"].append(item)
        cart["total"] = sum(i["price"] * i["quantity"] for i in cart["items"])

        # Save updated cart
        state.set(cart_key, cart)

        return {"success": True, "cart": cart, "user_id": user_id, "action": "add"}

    elif action == "remove":
        # Remove item from cart
        item_id = event.get("item_id")
        if not item_id:
            return {"error": "item_id is required for remove action"}

        # Get current cart
        cart = state.get(cart_key, default={"items": [], "total": 0})

        # Remove item
        cart["items"] = [i for i in cart["items"] if i["id"] != item_id]
        cart["total"] = sum(i["price"] * i["quantity"] for i in cart["items"])

        # Save updated cart
        state.set(cart_key, cart)

        return {"success": True, "cart": cart, "user_id": user_id, "action": "remove"}

    elif action == "clear":
        # Clear cart
        empty_cart = {"items": [], "total": 0}
        state.set(cart_key, empty_cart)

        return {
            "success": True,
            "cart": empty_cart,
            "user_id": user_id,
            "action": "clear",
        }

    elif action == "checkout":
        # Create pre-checkout snapshot
        cart = state.get(cart_key, default={"items": [], "total": 0})

        if not cart["items"]:
            return {"error": "Cannot checkout with empty cart"}

        # Create snapshot before checkout
        snapshot = state.snapshot(cart_key, label=f"pre-checkout-{user_id}")

        return {
            "success": True,
            "cart": cart,
            "snapshot": snapshot,
            "user_id": user_id,
            "action": "checkout",
            "message": "Cart ready for checkout, snapshot created",
        }

    else:
        return {"error": f"Unknown action: {action}"}


@flypy.function(
    name="session-counter-edge",
    description="Session counter using edge state",
)
def session_counter(event: dict) -> dict:
    """
    Simple session counter demonstrating atomic edge state operations.

    Args:
        event: Contains session_id

    Returns:
        Current count for the session
    """
    session_id = event.get("session_id", "default")

    # Use edge state with auto-detected fabric
    count = edge_state.get(f"sessions/{session_id}/count", default=0)

    # Increment count
    new_count = count + 1

    # Save back to state
    edge_state.set(f"sessions/{session_id}/count", new_count)

    return {
        "session_id": session_id,
        "count": new_count,
        "message": f"Session has been accessed {new_count} times",
    }


@flypy.function(
    name="rate-limiter-edge",
    description="Rate limiter using edge state at the edge",
)
def rate_limiter(event: dict) -> dict:
    """
    Simple rate limiter using edge state for request tracking.

    Args:
        event: Contains client_id and optionally max_requests and window_seconds

    Returns:
        Rate limit status
    """
    import time

    client_id = event.get("client_id")
    if not client_id:
        return {"error": "client_id is required"}

    max_requests = event.get("max_requests", 100)
    window_seconds = event.get("window_seconds", 60)

    now = time.time()
    window_start = now - window_seconds

    # Get current request log
    state = edge_state.EdgeStateClient()
    log_key = f"rate_limit/{client_id}"
    request_log = state.get(log_key, default={"requests": [], "count": 0})

    # Filter to current window
    request_log["requests"] = [
        req_time for req_time in request_log["requests"] if req_time > window_start
    ]

    # Check limit
    current_count = len(request_log["requests"])

    if current_count >= max_requests:
        return {
            "allowed": False,
            "client_id": client_id,
            "current_count": current_count,
            "max_requests": max_requests,
            "window_seconds": window_seconds,
            "retry_after": int(request_log["requests"][0] + window_seconds - now),
        }

    # Record request
    request_log["requests"].append(now)
    request_log["count"] = len(request_log["requests"])

    # Save back
    state.set(log_key, request_log)

    return {
        "allowed": True,
        "client_id": client_id,
        "current_count": request_log["count"],
        "max_requests": max_requests,
        "window_seconds": window_seconds,
        "remaining": max_requests - request_log["count"],
    }


# Example of using the EdgeStateManager decorator
manager = edge_state.EdgeStateManager()


@flypy.function(
    name="user-preferences-edge",
    description="User preferences with edge state decorator",
)
@manager.state("preferences", key="user_id", write=True)
def get_user_preferences(user_id: str) -> dict:
    """
    Get user preferences, creating defaults if not exists.

    This uses the EdgeStateManager decorator which automatically:
    - Reads from state when function is called
    - Writes result back to state when function returns
    - Creates state if it doesn't exist (create_if_not_exists=True)

    Args:
        user_id: The user identifier (used as state key)

    Returns:
        Default user preferences
    """
    return {
        "theme": "light",
        "language": "en",
        "notifications": True,
        "timezone": "UTC",
        "created_at": "auto-generated",
    }


# Build and test locally
if __name__ == "__main__":
    # Note: When running locally (not in WASM), the EdgeStateClient
    # falls back to using the HTTP API. For true edge testing,
    # deploy to FunctionFly and run at the edge.

    print("Edge State Example Functions")
    print("=" * 50)

    # Test shopping cart
    print("\n1. Shopping Cart Example:")
    result = shopping_cart_handler(
        {
            "action": "add",
            "user_id": "user123",
            "item": {"id": "item1", "name": "Widget", "price": 9.99, "quantity": 2},
        }
    )
    print(f"   Add result: {result}")

    # Test session counter
    print("\n2. Session Counter Example:")
    for i in range(3):
        result = session_counter({"session_id": "session_abc"})
        print(f"   Call {i + 1}: count = {result['count']}")

    # Test rate limiter
    print("\n3. Rate Limiter Example:")
    for i in range(5):
        result = rate_limiter(
            {"client_id": "client_xyz", "max_requests": 10, "window_seconds": 60}
        )
        print(
            f"   Request {i + 1}: allowed={result['allowed']}, count={result['current_count']}"
        )

    print("\n" + "=" * 50)
    print("Examples complete. Deploy to FunctionFly for edge state access.")
    print("Set FLYPY_STATE_FABRIC_ID environment variable to configure fabric.")
