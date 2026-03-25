CURRENCY_SYMBOLS = {
    "AED":"د.إ","AFN":"؋","ALL":"L","AMD":"֏","ANG":"ƒ","AOA":"Kz","ARS":"$","AUD":"A$",
    "AWG":"ƒ","AZN":"₼","BAM":"KM","BBD":"$","BDT":"৳","BGN":"лв","BHD":".د.ب","BIF":"Fr",
    "BMD":"$","BND":"$","BOB":"Bs.","BRL":"R$","BSD":"$","BTN":"Nu","BWP":"P","BYN":"Br",
    "BZD":"$","CAD":"CA$","CDF":"Fr","CHF":"CHF","CLP":"$","CNY":"¥","COP":"$","CRC":"₡",
    "CUP":"$","CVE":"$","CZK":"Kč","DJF":"Fr","DKK":"kr","DOP":"$","DZD":"د.ج","EGP":"£",
    "ERN":"Nfk","ETB":"Br","EUR":"€","FJD":"$","FKP":"£","GBP":"£","GEL":"₾","GHS":"₵",
    "GIP":"£","GMD":"D","GNF":"Fr","GTQ":"Q","GYD":"$","HKD":"HK$","HNL":"L","HRK":"kn",
    "HTG":"G","HUF":"Ft","IDR":"Rp","ILS":"₪","INR":"₹","IQD":"ع.د","IRR":"﷼","ISK":"kr",
    "JMD":"$","JOD":"JD","JPY":"¥","KES":"KSh","KGS":"лв","KHR":"៛","KMF":"Fr","KPW":"₩",
    "KRW":"₩","KWD":"د.ك","KYD":"$","KZT":"₸","LAK":"₭","LBP":"£","LKR":"₨","LRD":"$",
    "LSL":"L","LYD":"ل.د","MAD":"MAD","MDL":"L","MGA":"Ar","MKD":"ден","MMK":"K","MNT":"₮",
    "MOP":"P","MRU":"UM","MUR":"₨","MVR":"Rf","MWK":"MK","MXN":"$","MYR":"RM","MZN":"MT",
    "NAD":"$","NGN":"₦","NIO":"C$","NOK":"kr","NPR":"₨","NZD":"NZ$","OMR":"﷼","PAB":"B/.",
    "PEN":"S/.","PGK":"K","PHP":"₱","PKR":"₨","PLN":"zł","PYG":"₲","QAR":"﷼","RON":"lei",
    "RSD":"din","RUB":"₽","RWF":"Fr","SAR":"﷼","SBD":"$","SCR":"₨","SDG":"£","SEK":"kr",
    "SGD":"S$","SHP":"£","SLL":"Le","SOS":"Sh","SRD":"$","STN":"Db","SVC":"₡","SYP":"£",
    "SZL":"L","THB":"฿","TJS":"SM","TMT":"T","TND":"د.ت","TOP":"T$","TRY":"₺","TTD":"$",
    "TWD":"NT$","TZS":"Sh","UAH":"₴","UGX":"Sh","USD":"$","UYU":"$U","UZS":"лв","VES":"Bs.S",
    "VND":"₫","VUV":"Vt","WST":"T","XAF":"Fr","XCD":"$","XOF":"Fr","XPF":"Fr","YER":"﷼",
    "ZAR":"R","ZMW":"ZK","ZWL":"$",
}


def handler(event):
    code = event.get("code") if isinstance(event, dict) else None
    if not code:
        return {"ok": False, "error": "code is required (ISO 4217 currency code)"}
    upper = str(code).upper().strip()
    symbol = CURRENCY_SYMBOLS.get(upper)
    if symbol:
        return {"ok": True, "result": symbol, "symbol": symbol, "code": upper}
    return {"ok": False, "error": f"unknown currency code: {code}"}
