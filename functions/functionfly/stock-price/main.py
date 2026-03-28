def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    ticker = event.get("ticker")
    if ticker is None:
        return {"ok": False, "error": "ticker is required"}
    # Simulated prices (mock data)
    MOCK_PRICES = {
        "AAPL": 175.0, "MSFT": 380.0, "GOOGL": 140.0, "AMZN": 185.0,
        "NVDA": 500.0, "META": 480.0, "TSLA": 200.0, "BRK.B": 360.0,
        "JPM": 195.0, "V": 270.0, "JNJ": 155.0, "WMT": 60.0,
        "PG": 155.0, "MA": 460.0, "HD": 350.0, "BAC": 35.0,
        "XOM": 105.0, "CVX": 155.0, "ABBV": 165.0, "KO": 60.0
    }
    ticker = str(ticker).upper()
    currency = str(event.get("currency", "USD")).upper()
    price = MOCK_PRICES.get(ticker)
    if price is None:
        return {"ok": False, "error": f"Ticker {ticker} not found in mock data"}
    return {"ok": True, "result": price, "ticker": ticker, "currency": currency, "simulated": True}
