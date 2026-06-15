"""
Currency Converter - Convert amounts between currencies using exchange rates.
"""


# Static exchange rates (base: USD) - rates should be refreshed in production
EXCHANGE_RATES = {
    "USD": 1.0,
    "EUR": 0.85,
    "GBP": 0.73,
    "JPY": 110.0,
    "CAD": 1.25,
    "AUD": 1.35,
    "CHF": 0.92,
    "CNY": 6.45,
    "INR": 74.5,
    "MXN": 20.0,
    "BRL": 5.25,
    "KRW": 1180.0,
    "SGD": 1.35,
    "HKD": 7.78,
    "NOK": 8.85,
    "SEK": 8.75,
    "DKK": 6.35,
    "NZD": 1.42,
    "ZAR": 15.0,
    "RUB": 73.5,
}


def handler(event):
    if isinstance(event, dict):
        amount = event.get("amount")
        from_currency = event.get("from_currency", "USD")
        to_currency = event.get("to_currency", "EUR")
    else:
        amount, from_currency, to_currency = None, "USD", "EUR"

    if amount is None:
        return {"ok": False, "error": "amount is required"}

    try:
        amount = float(amount)
    except (ValueError, TypeError):
        return {"ok": False, "error": "amount must be a number"}

    if amount < 0:
        return {"ok": False, "error": "amount cannot be negative"}

    from_currency = from_currency.upper()
    to_currency = to_currency.upper()

    if from_currency not in EXCHANGE_RATES:
        return {"ok": False, "error": f"unsupported currency: {from_currency}"}
    if to_currency not in EXCHANGE_RATES:
        return {"ok": False, "error": f"unsupported currency: {to_currency}"}

    try:
        # Convert to USD first, then to target currency
        usd_amount = amount / EXCHANGE_RATES[from_currency]
        converted_amount = usd_amount * EXCHANGE_RATES[to_currency]

        return {
            "ok": True,
            "original_amount": amount,
            "original_currency": from_currency,
            "converted_amount": round(converted_amount, 2),
            "target_currency": to_currency,
            "exchange_rate": round(EXCHANGE_RATES[to_currency] / EXCHANGE_RATES[from_currency], 4),
            "inverse_rate": round(EXCHANGE_RATES[from_currency] / EXCHANGE_RATES[to_currency], 4),
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}