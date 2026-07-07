"""
Currency Converter - Convert amounts between currencies using exchange rates.
Uses Frankfurter API (ECB data, free, no API key required).
"""

import logging
import time
from dataclasses import dataclass, field
from datetime import datetime, timezone
from decimal import Decimal, ROUND_HALF_UP
from typing import TypedDict

logger = logging.getLogger(__name__)

FRANKFURTER_API = "https://api.frankfurter.app"
_CACHE_MAX_AGE_HOURS = 4

EXCHANGE_RATES: dict[str, Decimal] = {
    "USD": Decimal("1.0"),
    "EUR": Decimal("0.85"),
    "GBP": Decimal("0.73"),
    "JPY": Decimal("110.0"),
    "CAD": Decimal("1.25"),
    "AUD": Decimal("1.35"),
    "CHF": Decimal("0.92"),
    "CNY": Decimal("6.45"),
    "INR": Decimal("74.5"),
    "MXN": Decimal("20.0"),
    "BRL": Decimal("5.25"),
    "KRW": Decimal("1180.0"),
    "SGD": Decimal("1.35"),
    "HKD": Decimal("7.78"),
    "NOK": Decimal("8.85"),
    "SEK": Decimal("8.75"),
    "DKK": Decimal("6.35"),
    "NZD": Decimal("1.42"),
    "ZAR": Decimal("15.0"),
}

RATE_LAST_UPDATED = datetime(2026, 1, 1, tzinfo=timezone.utc)


class ConversionInput(TypedDict, total=False):
    amount: float | int | str
    from_currency: str
    to_currency: str


class ConversionResult(TypedDict):
    ok: bool
    original_amount: float
    original_currency: str
    converted_amount: float
    target_currency: str
    exchange_rate: float
    inverse_rate: float
    rate_last_updated: str


class ErrorResult(TypedDict):
    ok: bool
    error: str


def _fetch_rates_from_api() -> tuple[dict[str, Decimal], datetime]:
    """Fetch latest exchange rates from Frankfurter API (ECB data)."""
    try:
        response = requests.get(
            f"{FRANKFURTER_API}/latest",
            params={"from": "USD"},
            timeout=10,
        )
        response.raise_for_status()
        data = response.json()

        rates = {k: Decimal(str(v)) for k, v in data["rates"].items()}
        rates["USD"] = Decimal("1.0")

        updated = datetime.fromisoformat(data["date"]).replace(tzinfo=timezone.utc)

        logger.info("Fetched %d exchange rates from Frankfurter API", len(rates))
        return rates, updated

    except requests.RequestException as e:
        logger.warning("Failed to fetch rates from Frankfurter API: %s", e)
        return EXCHANGE_RATES, RATE_LAST_UPDATED


def _is_cache_stale() -> bool:
    """Check if cached rates need refresh (max age: 4 hours)."""
    age = datetime.now(timezone.utc) - RATE_LAST_UPDATED
    return age.total_seconds() > (_CACHE_MAX_AGE_HOURS * 3600)


def _refresh_rates_if_needed() -> None:
    """Refresh exchange rates if cache is stale."""
    global EXCHANGE_RATES, RATE_LAST_UPDATED

    if not _is_cache_stale():
        return

    new_rates, new_updated = _fetch_rates_from_api()

    EXCHANGE_RATES = new_rates
    RATE_LAST_UPDATED = new_updated


def handler(event: dict) -> ConversionResult | ErrorResult:
    """
    Convert an amount between currencies.

    Args:
        event: Dict with 'amount', 'from_currency' (default USD),
               and 'to_currency' (default EUR)

    Returns:
        ConversionResult on success or ErrorResult on failure.
    """
    _refresh_rates_if_needed()

    if not isinstance(event, dict):
        return ErrorResult(ok=False, error="event must be a dictionary")

    amount_val = event.get("amount")
    if amount_val is None:
        return ErrorResult(ok=False, error="amount is required")

    try:
        amount = Decimal(str(amount_val))
    except (ValueError, TypeError):
        return ErrorResult(ok=False, error="amount must be a number")

    if amount < 0:
        return ErrorResult(ok=False, error="amount cannot be negative")

    from_currency = event.get("from_currency", "USD").upper()
    to_currency = event.get("to_currency", "EUR").upper()

    if from_currency not in EXCHANGE_RATES:
        return ErrorResult(ok=False, error=f"unsupported currency: {from_currency}")
    if to_currency not in EXCHANGE_RATES:
        return ErrorResult(ok=False, error=f"unsupported currency: {to_currency}")

    from_rate = EXCHANGE_RATES[from_currency]
    to_rate = EXCHANGE_RATES[to_currency]

    converted_amount = (amount * to_rate) / from_rate
    exchange_rate = to_rate / from_rate
    inverse_rate = from_rate / to_rate

    return ConversionResult(
        ok=True,
        original_amount=float(amount),
        original_currency=from_currency,
        converted_amount=float(converted_amount.quantize(Decimal("0.01"), rounding=ROUND_HALF_UP)),
        target_currency=to_currency,
        exchange_rate=float(exchange_rate.quantize(Decimal("0.0001"), rounding=ROUND_HALF_UP)),
        inverse_rate=float(inverse_rate.quantize(Decimal("0.0001"), rounding=ROUND_HALF_UP)),
        rate_last_updated=RATE_LAST_UPDATED.isoformat(),
    )
