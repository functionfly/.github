"""
Trust policy evaluation helpers for FunctionFly agent integrations.
"""

from __future__ import annotations

from typing import Any, Dict, List, Tuple

from .agent_types import TrustPolicy


def _safe_float(value: Any, default: float = 0.0) -> float:
    try:
        return float(value)
    except (TypeError, ValueError):
        return default


def _extract_capabilities(candidate: Dict[str, Any]) -> List[str]:
    capabilities = candidate.get("capabilities") or []
    if not capabilities:
        manifest = candidate.get("manifest") or {}
        capabilities = manifest.get("capabilities") or []
    return [str(c).strip().lower() for c in capabilities if str(c).strip()]


def evaluate_candidate(policy: TrustPolicy, candidate: Dict[str, Any]) -> Tuple[bool, List[str]]:
    """
    Evaluate a candidate function against trust policy.

    Returns:
        (allowed, reasons)
    """
    reasons: List[str] = []

    verified = bool(candidate.get("verified", False))
    trust_score = _safe_float(candidate.get("trust_score"), default=0.0)
    trust_level = str(candidate.get("trust_level", "")).strip().lower()
    capabilities = _extract_capabilities(candidate)

    if policy.require_verified and not verified:
        reasons.append("candidate is not verified")

    if trust_score < float(policy.min_trust_score):
        reasons.append(
            f"trust_score {trust_score} below minimum {policy.min_trust_score}"
        )

    if policy.required_trust_levels:
        allowed_levels = {lvl.strip().lower() for lvl in policy.required_trust_levels if lvl.strip()}
        if trust_level not in allowed_levels:
            reasons.append(
                f"trust_level '{trust_level or 'unknown'}' not in required levels"
            )

    if policy.capabilities_allow:
        allowed_caps = {cap.strip().lower() for cap in policy.capabilities_allow if cap.strip()}
        if capabilities and not set(capabilities).issubset(allowed_caps):
            disallowed = sorted(set(capabilities) - allowed_caps)
            reasons.append(f"capabilities not allowed: {', '.join(disallowed)}")

    if policy.capabilities_deny:
        denied_caps = {cap.strip().lower() for cap in policy.capabilities_deny if cap.strip()}
        blocked = sorted(set(capabilities).intersection(denied_caps))
        if blocked:
            reasons.append(f"denied capabilities present: {', '.join(blocked)}")

    return (len(reasons) == 0, reasons)


def filter_candidates(
    policy: TrustPolicy,
    candidates: List[Dict[str, Any]],
) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]]]:
    """
    Filter candidate functions using TrustPolicy.

    Returns:
        (allowed_candidates, rejected_candidates_with_reasons)
    """
    allowed: List[Dict[str, Any]] = []
    rejected: List[Dict[str, Any]] = []

    for candidate in candidates:
        is_allowed, reasons = evaluate_candidate(policy, candidate)
        if is_allowed:
            allowed.append(candidate)
            continue
        rejected.append(
            {
                "candidate": candidate,
                "reasons": reasons,
            }
        )

    return allowed, rejected
