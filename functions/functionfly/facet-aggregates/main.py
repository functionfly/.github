from collections import Counter


def handler(event):
    products = event.get("products") if isinstance(event, dict) else None
    facets = event.get("facets", ["category", "brand", "rating_bucket"])
    if not products:
        return {"ok": False, "error": "products is required"}
    try:
        aggregates = {}
        for facet in facets:
            if facet == "rating_bucket":
                counts = Counter()
                for p in products:
                    r = float(p.get("rating", 0))
                    if r >= 4: bucket = "4+"
                    elif r >= 3: bucket = "3-4"
                    elif r >= 2: bucket = "2-3"
                    else: bucket = "<2"
                    counts[bucket] += 1
                aggregates[facet] = [{"value": k, "count": v} for k, v in sorted(counts.items())]
            elif facet == "price_range":
                counts = Counter()
                for p in products:
                    price = float(p.get("price", 0))
                    if price < 25: bucket = "Under $25"
                    elif price < 50: bucket = "$25-$50"
                    elif price < 100: bucket = "$50-$100"
                    elif price < 250: bucket = "$100-$250"
                    else: bucket = "$250+"
                    counts[bucket] += 1
                aggregates[facet] = [{"value": k, "count": v} for k, v in sorted(counts.items())]
            elif facet == "tags":
                counts = Counter()
                for p in products:
                    for tag in p.get("tags", []):
                        counts[str(tag)] += 1
                aggregates[facet] = [{"value": k, "count": v} for k, v in counts.most_common(20)]
            else:
                values = [p.get(facet) for p in products if p.get(facet) is not None]
                counts = Counter(str(v) for v in values)
                aggregates[facet] = [{"value": k, "count": v} for k, v in counts.most_common(50)]
        return {"ok": True, "result": aggregates, "total": len(products), "facets": facets}
    except Exception as e:
        return {"ok": False, "error": str(e)}
