POSITIVE = {"excellent","great","amazing","fantastic","wonderful","perfect","love","awesome","best","outstanding","superb","brilliant","incredible","happy","pleased","satisfied","good","nice","quality","beautiful","comfortable","reliable","affordable","worth","helpful","easy","fast","clear","smart","powerful","efficient","innovative","impressive","delightful","enjoy","positive","success","win","achieve","grow","improve","benefit","advantage","opportunity","hope","joy","peace","friendly","welcome","safe","trust","honest","fair","kind","generous","creative","passionate","dedicated","committed","effective","productive","successful","confident","strong","healthy","fresh","clean","bright","warm","fun","exciting","interesting","engaging","inspiring"}
NEGATIVE = {"terrible","horrible","awful","bad","poor","worst","disappointed","disappointing","broken","defective","slow","expensive","overpriced","useless","waste","regret","frustrating","difficult","complicated","unreliable","misleading","fake","cheap","hate","dislike","nasty","problem","issue","bug","error","fail","failure","wrong","mistake","bad","terrible","awful","annoying","confusing","boring","dull","pointless","harmful","dangerous","risk","threat","loss","damage","crisis","conflict","pain","suffering","fear","anger","sad","depressed","anxious","worried","stressed","tired","sick","weak","ugly","dark","cold","empty","broken","lost","failed"}
NEGATIONS = {"not","no","never","nothing","neither","nor","hardly","barely","doesn't","don't","isn't","wasn't","won't","can't","cannot","couldn't"}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        words = str(text).lower().split()
        pos, neg = 0, 0
        pos_words, neg_words = [], []
        for i, word in enumerate(words):
            clean = word.strip(".,!?\"'()[]{}:;")
            negated = i > 0 and words[i-1].strip(".,!?") in NEGATIONS
            if clean in POSITIVE:
                if negated: neg += 1; neg_words.append(f"not {clean}")
                else: pos += 1; pos_words.append(clean)
            elif clean in NEGATIVE:
                if negated: pos += 1; pos_words.append(f"not {clean}")
                else: neg += 1; neg_words.append(clean)
        total = pos + neg or 1
        score = round((pos - neg) / total, 4)
        sentiment = "positive" if score > 0.1 else ("negative" if score < -0.1 else "neutral")
        return {"ok": True, "result": sentiment, "sentiment": sentiment, "score": score, "positive_count": pos, "negative_count": neg, "positive_words": pos_words[:10], "negative_words": neg_words[:10]}
    except Exception as e:
        return {"ok": False, "error": str(e)}
