import os

def handler(input_data):
    """
    Perform sentiment analysis using AI capability.
    """
    try:
        model = os.getenv('AI_MODEL', 'sentiment')
        text = input_data.get('text', '')

        if not text:
            return {
                "status": "error",
                "message": "No text provided for analysis",
                "usage": "Provide 'text' field in input data"
            }

        # In a real implementation, this would call the AI host function
        # For now, we'll use a simple rule-based sentiment analysis

        positive_words = ['good', 'great', 'excellent', 'amazing', 'wonderful', 'fantastic', 'love', 'like', 'happy', 'joy']
        negative_words = ['bad', 'terrible', 'awful', 'hate', 'dislike', 'sad', 'angry', 'horrible', 'worst', 'poor']

        text_lower = text.lower()
        positive_count = sum(1 for word in positive_words if word in text_lower)
        negative_count = sum(1 for word in negative_words if word in text_lower)

        if positive_count > negative_count:
            sentiment = "positive"
            confidence = min(0.9, positive_count / max(1, len(text.split())))
        elif negative_count > positive_count:
            sentiment = "negative"
            confidence = min(0.9, negative_count / max(1, len(text.split())))
        else:
            sentiment = "neutral"
            confidence = 0.5

        result = {
            "status": "success",
            "model": model,
            "text": text,
            "sentiment": sentiment,
            "confidence": round(confidence, 2),
            "analysis": {
                "positive_indicators": positive_count,
                "negative_indicators": negative_count,
                "text_length": len(text)
            }
        }

        return result

    except Exception as e:
        return {
            "status": "error",
            "message": f"AI analysis failed: {str(e)}"
        }