CATEGORIES = {
    "technology": ["software","hardware","computer","phone","app","tech","digital","code","programming","ai","machine","learning","data","cloud","internet","web","api","algorithm","network","device","robot","automation","security","cyber","blockchain","crypto","virtual","augmented"],
    "sports": ["game","team","player","score","match","win","lose","championship","league","season","coach","athlete","stadium","tournament","sport","football","basketball","soccer","baseball","tennis","golf","swimming","running","cycling","fitness","gym","exercise","training","workout"],
    "finance": ["money","bank","invest","stock","market","economy","financial","trade","currency","revenue","profit","loss","budget","fund","capital","asset","portfolio","dividend","interest","loan","mortgage","tax","insurance","crypto","bitcoin","ethereum","dollar","euro"],
    "health": ["health","medical","doctor","hospital","patient","disease","treatment","medicine","drug","therapy","surgery","diagnosis","symptom","mental","fitness","nutrition","diet","wellness","care","clinic","vaccine","pandemic","virus","bacteria","cancer","heart","blood","brain"],
    "politics": ["government","election","vote","president","congress","senate","policy","law","political","democracy","republican","democrat","liberal","conservative","minister","parliament","legislation","constitution","rights","freedom","protest","campaign","candidate","party"],
    "entertainment": ["movie","film","music","song","artist","actor","celebrity","show","tv","series","album","concert","festival","award","streaming","netflix","youtube","podcast","dance","theater","comedy","drama","action","romance","horror","entertainment"],
    "science": ["science","research","study","discovery","experiment","theory","physics","chemistry","biology","astronomy","space","planet","evolution","genetics","climate","environment","ecology","quantum","particle","molecule","element","compound","cell","organism"],
    "business": ["business","company","startup","entrepreneur","ceo","executive","product","market","customer","client","sale","revenue","growth","strategy","management","brand","marketing","advertising","supply","demand","industry","corporate","enterprise"],
    "travel": ["travel","trip","vacation","holiday","hotel","flight","destination","tourist","country","city","beach","mountain","culture","food","adventure","explore","visit","tour","visa","passport","airport","resort","cruise"],
    "food": ["food","recipe","cook","restaurant","meal","dish","ingredient","flavor","taste","cuisine","chef","menu","drink","coffee","tea","wine","beer","vegetarian","vegan","organic","nutrition","calorie","diet","breakfast","lunch","dinner"],
}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    top_n = int(event.get("top_n", 3))
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        words = set(str(text).lower().split())
        scores = {}
        for category, keywords in CATEGORIES.items():
            scores[category] = sum(1 for kw in keywords if kw in words)
        ranked = sorted(scores.items(), key=lambda x: x[1], reverse=True)
        top_cat = ranked[0][0] if ranked[0][1] > 0 else "general"
        return {
            "ok": True,
            "result": top_cat,
            "category": top_cat,
            "top_categories": [{"category": c, "score": s} for c, s in ranked[:top_n] if s > 0],
            "scores": scores
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
