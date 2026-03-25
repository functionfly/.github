CURRENCY_SYMBOLS = {
    "USD": "$", "EUR": "€", "GBP": "£", "JPY": "¥", "CNY": "¥", "KRW": "₩",
    "INR": "₹", "BRL": "R$", "MXN": "$", "CAD": "CA$", "AUD": "A$", "CHF": "CHF",
    "SEK": "kr", "NOK": "kr", "DKK": "kr", "HKD": "HK$", "SGD": "S$", "NZD": "NZ$",
    "ZAR": "R", "RUB": "₽", "PLN": "zł", "TRY": "₺", "THB": "฿", "IDR": "Rp",
    "MYR": "RM", "PHP": "₱", "SAR": "﷼", "AED": "د.إ", "CZK": "Kč", "HUF": "Ft",
}
DECIMAL_CURRENCIES = {"JPY", "KRW", "VND", "BIF", "CLP", "GNF", "ISK", "KMF", "PYG", "RWF", "UGX", "UYI", "XAF", "XOF", "XPF"}


def handler(event):
    amount = event.get("amount") if isinstance(event, dict) else None
    currency = event.get("currency", "USD").upper()
    symbol_position = event.get("symbol_position", "before")
    if amount is None:
        return {"ok": False, "error": "amount is required"}
    try:
        amt = float(amount)
        symbol = CURRENCY_SYMBOLS.get(currency, currency)
        decimals = 0 if currency in DECIMAL_CURRENCIES else 2
        formatted_num = f"{amt:,.{decimals}f}"
        if symbol_position == "after":
            result = f"{formatted_num} {symbol}"
        else:
            result = f"{symbol}{formatted_num}"
        return {"ok": True, "result": result, "amount": amt, "currency": currency, "symbol": symbol}
    except Exception as e:
        return {"ok": False, "error": str(e)}
