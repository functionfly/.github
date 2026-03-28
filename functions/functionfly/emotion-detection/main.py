import re

EMOTION_LEXICON = {
    "joy": {"happy","happiness","joy","joyful","joyous","delight","delighted","elated","ecstatic","thrilled","excited","cheerful","glad","pleased","content","satisfied","blissful","euphoric","jubilant","overjoyed","wonderful","fantastic","amazing","love","loving","adore","enjoy","fun","laugh","smile","celebrate","celebration","party","wonderful","great","excellent","awesome","brilliant","superb","magnificent","terrific","splendid","marvelous","radiant","bright","sunny","warm","cozy","comfortable","peaceful","serene","harmonious","grateful","thankful","appreciate","blessed","fortunate","lucky","hopeful","optimistic","confident","proud","accomplished","success","win","victory","triumph"},
    "sadness": {"sad","sadness","unhappy","sorrow","sorrowful","grief","grieving","mourning","mourn","cry","crying","tears","weep","weeping","depressed","depression","melancholy","melancholic","gloomy","gloom","miserable","misery","despair","despairing","hopeless","helpless","lonely","loneliness","alone","isolated","abandoned","rejected","heartbroken","heartbreak","loss","lost","miss","missing","regret","regretful","sorry","apologize","apology","disappointed","disappointment","hurt","pain","suffering","ache","aching","broken","shattered","devastated","devastation","tragic","tragedy","unfortunate","unlucky","unfortunate","pitiful","pathetic","wretched","forlorn","desolate","bereft","inconsolable"},
    "anger": {"angry","anger","furious","fury","rage","enraged","outraged","outrage","mad","livid","irate","irritated","irritation","annoyed","annoyance","frustrated","frustration","hostile","hostility","aggressive","aggression","violent","violence","hate","hatred","despise","loathe","loathing","resent","resentment","bitter","bitterness","vengeful","vengeance","revenge","spite","spiteful","malicious","malice","cruel","cruelty","harsh","harshness","mean","meanness","rude","rudeness","offensive","offend","insult","insulting","disrespect","disrespectful","contempt","contemptuous","scorn","scornful","disdain","disdainful","disgusted","disgust","revolted","revolt","appalled","appalling","outrageous","unacceptable","intolerable","unbearable","infuriating","maddening","exasperating","provoking","provocative"},
    "fear": {"afraid","fear","fearful","scared","terrified","terror","horrified","horror","dread","dreading","anxious","anxiety","worried","worry","nervous","nervousness","panic","panicking","phobia","phobic","threatened","threatening","danger","dangerous","risk","risky","unsafe","insecure","vulnerable","exposed","helpless","powerless","weak","trembling","shaking","shaky","uneasy","unease","apprehensive","apprehension","tense","tension","stressed","stress","overwhelmed","overwhelm","alarmed","alarm","startled","startle","shocked","shock","stunned","stun","paralyzed","paralysis","frozen","petrified","petrify","intimidated","intimidation","cowardly","coward","timid","timidity","shy","shyness"},
    "surprise": {"surprised","surprise","astonished","astonishment","amazed","amazement","astounded","astounding","shocked","shock","stunned","stunning","unexpected","unexpectedly","sudden","suddenly","unbelievable","incredible","extraordinary","remarkable","phenomenal","miraculous","miracle","wonder","wonderful","wow","whoa","oh","ah","gasp","gasping","speechless","dumbfounded","flabbergasted","bewildered","bewilderment","perplexed","perplexity","confused","confusion","disbelief","disbelieve","cannot believe","hard to believe","never expected","did not expect","out of nowhere","out of the blue","caught off guard","taken aback"},
    "disgust": {"disgusted","disgust","revolted","revolt","repulsed","repulsion","repelled","repel","nauseated","nausea","sick","sickened","gross","grossed out","yuck","ew","eww","ugh","horrible","horrid","hideous","ugly","vile","vicious","nasty","filthy","dirty","contaminated","polluted","rotten","putrid","foul","stench","stink","stinking","smelly","offensive","offend","repugnant","repugnance","abhorrent","abhor","abomination","abominable","loathsome","loathe","detestable","detest","despicable","despise","contemptible","contempt","disdainful","disdain","scornful","scorn","revolting","nauseating","sickening","stomach-turning","stomach-churning"},
    "trust": {"trust","trustworthy","reliable","dependable","honest","honesty","faithful","loyalty","loyal","sincere","sincerity","genuine","authentic","real","true","truthful","transparent","open","fair","just","integrity","moral","ethical","principled","responsible","accountable","credible","credibility","believe","belief","faith","confidence","confident","certain","certainty","sure","assured","assurance","secure","security","safe","safety","protected","protection","support","supportive","helpful","helpful","caring","care","compassionate","compassion","empathetic","empathy","understanding","kind","kindness","generous","generosity","benevolent","benevolence","charitable","charity","cooperative","cooperation","collaborative","collaboration","united","unity","solidarity","community","belonging"},
    "anticipation": {"anticipate","anticipation","expect","expectation","hope","hopeful","looking forward","excited","excitement","eager","eagerness","enthusiastic","enthusiasm","curious","curiosity","interested","interest","motivated","motivation","inspired","inspiration","driven","drive","ambitious","ambition","goal","goals","plan","planning","prepare","preparation","ready","readiness","await","awaiting","upcoming","soon","future","next","coming","approaching","imminent","pending","scheduled","arranged","organized","set","determined","resolved","committed","dedicated","focused","concentrated","attentive","alert","watchful","vigilant","anticipatory","preparatory","preliminary","introductory","initial","beginning","starting","launching","initiating","commencing"}
}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        t = text.lower()
        words = set(re.findall(r'\b\w+\b', t))
        scores = {}
        for emotion, lexicon in EMOTION_LEXICON.items():
            count = len(words & lexicon)
            scores[emotion] = count
        total = sum(scores.values()) or 1
        normalized = {e: round(s / total, 4) for e, s in scores.items()}
        dominant = max(scores, key=scores.get)
        if scores[dominant] == 0:
            dominant = "neutral"
            normalized["neutral"] = 1.0
        return {
            "ok": True,
            "result": {"dominant": dominant, "scores": normalized},
            "emotions": normalized,
            "dominant": dominant,
            "raw_counts": scores
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
