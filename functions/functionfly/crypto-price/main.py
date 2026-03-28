def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    symbol = event.get("symbol")
    if symbol is None:
        return {"ok": False, "error": "symbol is required"}
    # Simulated prices (mock data)
    MOCK_PRICES = {
        "BTC": 45000.0, "ETH": 2500.0, "SOL": 100.0, "BNB": 300.0,
        "ADA": 0.45, "XRP": 0.55, "DOT": 7.0, "DOGE": 0.08,
        "AVAX": 35.0, "MATIC": 0.85, "LINK": 15.0, "UNI": 6.0,
        "LTC": 70.0, "ATOM": 10.0, "ALGO": 0.15
    }
    symbol = str(symbol).upper()
    currency = str(event.get("currency", "USD")).upper()
    price = MOCK_PRICES.get(symbol)
    if price is None:
        return {"ok": False, "error": f"Symbol {symbol} not found in mock data"}
    return {"ok": True, "result": price, "symbol": symbol, "currency": currency, "simulated": True}
