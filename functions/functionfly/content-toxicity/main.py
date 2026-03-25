TOXIC_PATTERNS = ["fuck","shit","ass","bitch","cunt","dick","piss","cock","whore","slut","bastard","damn","hell","crap","idiot","stupid","moron","loser","ugly","fat","dumb","retard","nazi","kill","murder","hate","die","dead","blood","violence","attack","bomb","gun","shoot","stab","racist","sexist","bigot"]
THREAT_PATTERNS = ["kill you","hurt you","attack you","destroy you","bomb","shoot you","stab you","murder","suicide","self-harm","i will","gonna kill","want to hurt"]
IDENTITY_ATTACKS = ["racist","sexist","homophobic","transphobic","antisemitic","islamophobic","bigot","n-word","f-word"]


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text).lower()
        toxic_matches = [w for w in TOXIC_PATTERNS if w in t]
        threat_matches = [p for p in THREAT_PATTERNS if p in t]
        identity_matches = [p for p in IDENTITY_ATTACKS if p in t]
        toxic_score = min(1.0, len(toxic_matches) * 0.15 + len(threat_matches) * 0.4 + len(identity_matches) * 0.3)
        is_toxic = toxic_score >= 0.3
        severity = "high" if toxic_score >= 0.7 else ("medium" if toxic_score >= 0.3 else "low")
        return {
            "ok": True,
            "result": toxic_score,
            "toxicity_score": round(toxic_score, 4),
            "is_toxic": is_toxic,
            "severity": severity,
            "categories": {
                "profanity": len(toxic_matches) > 0,
                "threats": len(threat_matches) > 0,
                "identity_attack": len(identity_matches) > 0,
            },
            "flagged_terms": toxic_matches[:5]
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
