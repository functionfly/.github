LANGUAGE_TRIGRAMS = {
    "en": ["the","and","ing","ion","tha","ent","ati","for","her","ter","hat","his","tio","not","ive","ver","all","ons","nce","our"],
    "es": ["que","las","los","del","una","ión","ent","con","par","ado","nte","est","por","ica","cia","ara","ene","ide","sta","des"],
    "fr": ["les","est","des","que","une","ent","ion","ons","ant","ati","par","our","que","ais","ait","aus","tre","ell","eur","ver"],
    "de": ["die","und","ist","den","ein","für","mit","von","der","das","ich","sie","war","auf","des","wir","ers","nte","eit","hen"],
    "pt": ["que","uma","dos","com","são","ent","ção","ado","eis","nas","sta","mos","nto","tes","por","ara","ção","pre","pro","ver"],
    "it": ["che","per","con","una","del","gli","lla","ent","ion","ato","are","tti","ita","all","nti","tti","ato","esi","rsi","ati"],
    "nl": ["een","van","het","dat","zijn","aan","den","voor","met","die","als","bij","hij","hebben","worden","wij","maar","kan","meer","ijk"],
    "ru": ["ого","ние","ные","ски","что","ения","ного","ане","ыва","ств","при","ред","она","ели","ние","про","пра","ела","тор","ять"],
    "zh": ["的","是","了","在","有","和","不","我","这","个"],
    "ja": ["の","に","は","を","が","で","と","し","て","も"],
    "ar": ["في","من","على","هذا","أن","كان","لما","إلى","كما","هذه"],
    "ko": ["이","는","을","의","가","에","한","하","들","로"],
}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    top_n = int(event.get("top_n", 3))
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text).lower()[:500]
        scores = {}
        for lang, patterns in LANGUAGE_TRIGRAMS.items():
            score = sum(1 for p in patterns if p in t)
            scores[lang] = score
        ranked = sorted(scores.items(), key=lambda x: x[1], reverse=True)
        top = ranked[0][0] if ranked[0][1] > 0 else "unknown"
        confidence = ranked[0][1] / max(len(LANGUAGE_TRIGRAMS.get(top, [])), 1)
        return {
            "ok": True,
            "result": top,
            "language": top,
            "confidence": round(min(1.0, confidence), 4),
            "top_candidates": [{"language": l, "score": s} for l, s in ranked[:top_n]],
            "is_confident": confidence > 0.3
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
