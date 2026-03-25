LOCALE_CONFIG = {
    "en-US": {"decimal": ".", "thousands": ",", "symbol": "$", "position": "before", "code": "USD"},
    "en-GB": {"decimal": ".", "thousands": ",", "symbol": "£", "position": "before", "code": "GBP"},
    "de-DE": {"decimal": ",", "thousands": ".", "symbol": "€", "position": "after", "code": "EUR"},
    "fr-FR": {"decimal": ",", "thousands": " ", "symbol": "€", "position": "after", "code": "EUR"},
    "ja-JP": {"decimal": ".", "thousands": ",", "symbol": "¥", "position": "before", "code": "JPY", "decimals": 0},
    "zh-CN": {"decimal": ".", "thousands": ",", "symbol": "¥", "position": "before", "code": "CNY"},
    "ko-KR": {"decimal": ".", "thousands": ",", "symbol": "₩", "position": "before", "code": "KRW", "decimals": 0},
    "pt-BR": {"decimal": ",", "thousands": ".", "symbol": "R$", "position": "before", "code": "BRL"},
    "es-ES": {"decimal": ",", "thousands": ".", "symbol": "€", "position": "after", "code": "EUR"},
    "it-IT": {"decimal": ",", "thousands": ".", "symbol": "€", "position": "after", "code": "EUR"},
    "nl-NL": {"decimal": ",", "thousands": ".", "symbol": "€", "position": "before", "code": "EUR"},
    "pl-PL": {"decimal": ",", "thousands": " ", "symbol": "zł", "position": "after", "code": "PLN"},
    "ru-RU": {"decimal": ",", "thousands": " ", "symbol": "₽", "position": "after", "code": "RUB"},
    "tr-TR": {"decimal": ",", "thousands": ".", "symbol": "₺", "position": "before", "code": "TRY"},
    "hi-IN": {"decimal": ".", "thousands": ",", "symbol": "₹", "position": "before", "code": "INR"},
    "ar-SA": {"decimal": ".", "thousands": ",", "symbol": "﷼", "position": "after", "code": "SAR"},
    "sv-SE": {"decimal": ",", "thousands": " ", "symbol": "kr", "position": "after", "code": "SEK"},
}


def handler(event):
    amount = event.get("amount") if isinstance(event, dict) else None
    locale = event.get("locale", "en-US")
    if amount is None:
        return {"ok": False, "error": "amount is required"}
    try:
        amt = float(amount)
        cfg = LOCALE_CONFIG.get(locale, LOCALE_CONFIG["en-US"])
        decimals = cfg.get("decimals", 2)
        # Format integer part with thousands separator
        integer_part = str(int(abs(amt)))
        grouped = []
        for i, c in enumerate(reversed(integer_part)):
            if i > 0 and i % 3 == 0:
                grouped.append(cfg["thousands"])
            grouped.append(c)
        integer_formatted = "".join(reversed(grouped))
        if decimals > 0:
            frac = f"{abs(amt):.{decimals}f}".split(".")[1]
            num_str = f"{integer_formatted}{cfg['decimal']}{frac}"
        else:
            num_str = integer_formatted
        if amt < 0:
            num_str = f"-{num_str}"
        sym = cfg["symbol"]
        if cfg["position"] == "after":
            result = f"{num_str} {sym}"
        else:
            result = f"{sym}{num_str}"
        return {"ok": True, "result": result, "locale": locale, "currency": cfg["code"]}
    except Exception as e:
        return {"ok": False, "error": str(e)}
