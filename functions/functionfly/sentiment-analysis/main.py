import re

POSITIVE = {"good","great","excellent","amazing","wonderful","fantastic","love","like","enjoy","happy","pleased","satisfied","perfect","best","awesome","brilliant","outstanding","superb","magnificent","delightful","positive","nice","beautiful","helpful","useful","effective","efficient","impressive","remarkable","exceptional","terrific","splendid","marvelous","joyful","cheerful","glad","thrilled","excited","grateful","thankful","appreciate","recommend","success","successful","win","winning","benefit","beneficial","improve","improvement","better","best","top","superior","quality","reliable","trustworthy","honest","fair","kind","generous","friendly","warm","caring","supportive","innovative","creative","smart","intelligent","clever","talented","skilled","expert","professional","clean","clear","simple","easy","fast","quick","smooth","comfortable","convenient","affordable","valuable","worthy","worthy","fun","entertaining","interesting","engaging","inspiring","motivating","encouraging","empowering","uplifting","refreshing","energizing","vibrant","lively","dynamic","powerful","strong","robust","solid","stable","secure","safe","healthy","fresh","natural","pure","authentic","genuine","real","true","honest","transparent","open","accessible","inclusive","diverse","innovative","progressive","forward","advance","growth","opportunity","potential","promise","hope","optimistic","confident","proud","accomplished","achieved","succeed","thrive","flourish","prosper","excel","shine","glow","radiant","bright","light","positive","upbeat","enthusiastic","passionate","dedicated","committed","loyal","faithful","reliable","dependable","consistent","steady","calm","peaceful","serene","harmonious","balanced","wholesome","nourishing","enriching","fulfilling","rewarding","satisfying","meaningful","purposeful","valuable","precious","cherished","beloved","adored","treasured"}

NEGATIVE = {"bad","terrible","awful","horrible","dreadful","hate","dislike","poor","worst","disappointing","disappointed","frustrating","frustrated","annoying","annoyed","angry","upset","sad","unhappy","miserable","depressed","worried","anxious","scared","afraid","fearful","disgusting","disgusted","revolting","repulsive","offensive","rude","mean","cruel","harsh","unfair","unjust","wrong","incorrect","false","fake","dishonest","deceptive","misleading","confusing","complicated","difficult","hard","slow","broken","damaged","defective","faulty","unreliable","untrustworthy","dangerous","harmful","toxic","poisonous","destructive","devastating","catastrophic","disastrous","tragic","terrible","horrible","dreadful","awful","appalling","shocking","outrageous","unacceptable","intolerable","unbearable","painful","suffering","agony","misery","despair","hopeless","helpless","powerless","weak","fragile","vulnerable","exposed","threatened","endangered","at risk","failure","fail","failed","losing","lost","defeat","defeated","problem","issue","trouble","difficulty","challenge","obstacle","barrier","hindrance","setback","delay","waste","wasteful","inefficient","ineffective","useless","worthless","pointless","meaningless","empty","hollow","shallow","superficial","trivial","insignificant","irrelevant","unnecessary","redundant","excessive","extreme","radical","dangerous","risky","uncertain","unstable","unreliable","inconsistent","unpredictable","chaotic","messy","dirty","polluted","contaminated","corrupt","broken","damaged","destroyed","ruined","spoiled","rotten","expired","outdated","obsolete","inferior","substandard","mediocre","average","ordinary","bland","boring","dull","tedious","monotonous","repetitive","stale","tired","exhausted","drained","depleted","overwhelmed","stressed","pressured","burdened","overloaded","swamped","drowning","struggling","suffering","enduring","tolerating","putting up with","dealing with","coping with","managing","surviving"}

NEGATORS = {"not","no","never","neither","nor","nobody","nothing","nowhere","hardly","barely","scarcely","without","lack","lacking","absent","missing","void","empty","free from","devoid","cannot","can't","won't","wouldn't","shouldn't","couldn't","didn't","doesn't","don't","isn't","aren't","wasn't","weren't","haven't","hasn't","hadn't"}

INTENSIFIERS = {"very","extremely","absolutely","completely","totally","utterly","highly","deeply","strongly","greatly","incredibly","remarkably","exceptionally","particularly","especially","quite","rather","fairly","somewhat","slightly","a bit","a little","pretty","really","truly","genuinely","sincerely","honestly","definitely","certainly","surely","undoubtedly","unquestionably","without doubt","beyond doubt","clearly","obviously","evidently","apparently","seemingly","supposedly","allegedly","reportedly","supposedly"}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        t = text.lower()
        words = re.findall(r'\b\w+\b', t)
        pos_score = 0.0
        neg_score = 0.0
        for i, word in enumerate(words):
            multiplier = 1.0
            # Check for negation in previous 3 words
            prev_words = words[max(0, i-3):i]
            negated = any(n in prev_words for n in NEGATORS)
            # Check for intensifier
            if i > 0 and words[i-1] in INTENSIFIERS:
                multiplier = 1.5
            if word in POSITIVE:
                if negated:
                    neg_score += multiplier
                else:
                    pos_score += multiplier
            elif word in NEGATIVE:
                if negated:
                    pos_score += multiplier * 0.5
                else:
                    neg_score += multiplier
        total = pos_score + neg_score or 1.0
        compound = round((pos_score - neg_score) / total, 4)
        if compound > 0.1:
            sentiment = "positive"
        elif compound < -0.1:
            sentiment = "negative"
        else:
            sentiment = "neutral"
        return {
            "ok": True,
            "result": {"sentiment": sentiment, "score": compound},
            "sentiment": sentiment,
            "score": compound,
            "positive": round(pos_score / total, 4),
            "negative": round(neg_score / total, 4),
            "neutral": round(1 - abs(compound), 4)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
