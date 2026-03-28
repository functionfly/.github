import re

INTENT_PATTERNS = {
    "greeting": [r'\b(hello|hi|hey|howdy|greetings|good morning|good afternoon|good evening|what\'s up|sup|yo)\b'],
    "farewell": [r'\b(bye|goodbye|see you|see ya|later|farewell|take care|good night|ciao|adios|au revoir)\b'],
    "question": [r'^\s*(what|who|where|when|why|how|which|whose|whom|is|are|was|were|do|does|did|can|could|would|should|will|have|has|had)\b', r'\?'],
    "command": [r'^\s*(do|make|create|build|generate|write|send|open|close|start|stop|run|execute|install|delete|remove|add|update|change|set|get|show|display|print|save|load|upload|download|deploy|launch|activate|deactivate|enable|disable|configure|setup|initialize|reset|restart|refresh|reload|clear|clean|fix|repair|check|verify|validate|test|debug|analyze|process|convert|transform|export|import|merge|split|sort|filter|search|find|replace|copy|move|rename|archive|backup|restore)\b'],
    "request": [r'\b(please|could you|can you|would you|will you|i need|i want|i\'d like|i would like|help me|assist me|i\'m looking for|i am looking for|i need help|i need assistance|could you please|can you please|would you please|will you please)\b'],
    "complaint": [r'\b(problem|issue|error|bug|broken|not working|doesn\'t work|doesn\'t work|failed|failure|wrong|incorrect|bad|terrible|awful|horrible|disappointed|frustrating|annoying|unacceptable|this is wrong|this is bad|this is terrible|this is awful|this is horrible|this is broken|this is not working|this is not right|this is not correct|this is not good|this is not acceptable|this is not okay|this is not ok)\b'],
    "compliment": [r'\b(great|excellent|amazing|wonderful|fantastic|love|like|enjoy|happy|pleased|satisfied|perfect|best|awesome|brilliant|outstanding|superb|magnificent|delightful|thank you|thanks|appreciate|well done|good job|nice work|great work|excellent work|amazing work|wonderful work|fantastic work|love it|love this|love the|great job|nice job|good work|well done|kudos|bravo|congrats|congratulations)\b'],
    "information": [r'\b(tell me|inform me|let me know|i want to know|i need to know|what is|what are|who is|who are|where is|where are|when is|when are|how is|how are|explain|describe|define|clarify|elaborate|provide information|give me information|share information|what does|what do|how does|how do)\b'],
    "purchase": [r'\b(buy|purchase|order|shop|checkout|cart|price|cost|how much|payment|pay|subscribe|subscription|plan|pricing|discount|coupon|promo|deal|offer|sale|free trial|trial)\b'],
    "cancel": [r'\b(cancel|cancellation|unsubscribe|stop|end|terminate|close account|delete account|remove account|opt out|opt-out|withdraw|refund|return|chargeback)\b'],
    "support": [r'\b(support|help|assistance|troubleshoot|troubleshooting|fix|repair|resolve|solution|how to|how do i|how can i|step by step|guide|tutorial|documentation|docs|faq|frequently asked questions)\b'],
}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        t = text.lower().strip()
        scores = {intent: 0 for intent in INTENT_PATTERNS}
        for intent, patterns in INTENT_PATTERNS.items():
            for pat in patterns:
                if re.search(pat, t, re.IGNORECASE):
                    scores[intent] += 1
        best_intent = max(scores, key=scores.get)
        best_score = scores[best_intent]
        if best_score == 0:
            best_intent = "unknown"
            confidence = 0.0
        else:
            total = sum(scores.values()) or 1
            confidence = round(best_score / total, 4)
        return {
            "ok": True,
            "result": {"intent": best_intent, "confidence": confidence},
            "intent": best_intent,
            "confidence": confidence,
            "scores": scores
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
