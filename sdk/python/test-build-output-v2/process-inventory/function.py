@flypy.function(
    name="process-inventory",
    description="Process inventory data with filtering and aggregation",
    deterministic=True
)
def process_inventory(inventory_data: Dict[str, Any]) -> Dict[str, Any]:
    """Process inventory data with filtering and statistics."""
    items = inventory_data.get("items", [])
    filters = inventory_data.get("filters", {})

    # Apply filters
    filtered_items = []
    for item in items:
        # Category filter
        if "category" in filters and item.get("category") != filters["category"]:
            continue

        # Price range filter
        if "min_price" in filters and item.get("price", 0) < filters["min_price"]:
            continue
        if "max_price" in filters and item.get("price", 0) > filters["max_price"]:
            continue

        # Stock level filter
        if "min_stock" in filters and item.get("stock_quantity", 0) < filters["min_stock"]:
            continue

        filtered_items.append(item)

    # Calculate statistics
    if filtered_items:
        prices = [item["price"] for item in filtered_items]
        stocks = [item["stock_quantity"] for item in filtered_items]

        stats = {
            "total_items": len(filtered_items),
            "avg_price": sum(prices) / len(prices),
            "min_price": min(prices),
            "max_price": max(prices),
            "total_value": sum(price * stock for price, stock in zip(prices, stocks)),
            "low_stock_items": len([s for s in stocks if s < 10])
        }
    else:
        stats = {
            "total_items": 0,
            "avg_price": 0.0,
            "min_price": 0.0,
            "max_price": 0.0,
            "total_value": 0.0,
            "low_stock_items": 0
        }

    return {
        "filtered_items": filtered_items,
        "statistics": stats,
        "filters_applied": filters
    }
