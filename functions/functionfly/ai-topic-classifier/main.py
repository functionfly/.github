"""AI Topic Classifier - Classify text into categories using keyword matching."""
import re
from collections import Counter


DEFAULT_CATEGORIES = {
    "technology": ["software", "hardware", "computer", "digital", "tech", "app", "application", "device", "system", "data", "cloud", "ai", "machine learning", "web", "internet", "online", "computer", "programming", "code", "developer", "api", "database", "server", "network", "cybersecurity", "privacy", "software", "app", "mobile", "smartphone", "computer", "laptop", "tablet"],
    "business": ["company", "startup", "enterprise", "business", "revenue", "profit", "market", "industry", "investment", "investor", "funding", "ceo", "management", "strategy", "growth", "sales", "marketing", "customer", "client", "product", "service", "corporate", "organization", "business", "entrepreneur", "small business", "corporation", "llc", "inc"],
    "marketing": ["marketing", "advertising", "brand", "branding", "social media", "content", "seo", "campaign", "promotion", "digital marketing", "email marketing", "influencer", "engagement", "audience", "reach", "conversion", "leads", "advertisement", "ads", "facebook", "instagram", "twitter", "linkedin", "google", "analytics", "advertising", "campaign", "promotion"],
    "healthcare": ["health", "medical", "healthcare", "doctor", "patient", "hospital", "treatment", "diagnosis", "health", "wellness", "disease", "illness", "symptom", "therapy", "medicine", "pharmaceutical", "drug", "surgery", "clinic", "healthcare", "insurance", "patient", "nurse", "physician", "health", "wellness", "fitness", "exercise", "diet", "nutrition"],
    "finance": ["finance", "banking", "investment", "stock", "market", "trading", "financial", "money", "fund", "portfolio", "asset", "liability", "debt", "equity", "bond", "interest", "rate", "credit", "loan", "mortgage", "bank", "investor", "wall street", "fintech", "cryptocurrency", "bitcoin", "blockchain", "crypto", "trading", "shares", "dividend"],
    "education": ["education", "learning", "school", "university", "college", "student", "teacher", "course", "training", "degree", "academic", "study", "knowledge", "skill", "online course", "elearning", "tutoring", "classroom", "lesson", "curriculum", "assignment", "exam", "grades", "education", "teaching", "instruction", "development", "workshop", "seminar"],
    "food": ["food", "restaurant", "cooking", "recipe", "meal", "chef", "dining", "cuisine", "ingredient", "diet", "nutrition", "food", "eating", "restaurant", "kitchen", "baking", "delicious", "tasty", "flavor", "dish", "gourmet", "foodie", "yummy", "recipe", "chef", "dinner", "lunch", "breakfast", "snack", "drink", "beverage"],
    "travel": ["travel", "vacation", "trip", "destination", "hotel", "flight", "tourism", "airline", "booking", "holiday", "adventure", "explore", "journey", "itinerary", "luggage", "passport", "beach", "mountain", "city", "country", "travel", "tourist", "sightseeing", "resort", "hostel", "airbnb", "cruise", "backpacking", "tour", "vacation"],
    "fashion": ["fashion", "clothing", "style", "apparel", "designer", "trend", "wardrobe", "outfit", "fashion", "wear", "dress", "shoes", "accessories", "luxury", "brand", "runway", "collection", "season", "fashion", "stylish", "elegant", "chic", "vogue", "haute couture", "apparel", "textile", "fabric", "fashion"],
    "sports": ["sports", "game", "team", "player", "athlete", "championship", "tournament", "score", "win", "loss", "match", "league", "fitness", "exercise", "workout", "training", "coach", "stadium", "competition", "racing", "football", "basketball", "soccer", "baseball", "tennis", "golf", "olympics", "sports"],
    "entertainment": ["entertainment", "movie", "film", "music", "celebrity", "tv", "show", "streaming", "netflix", "series", "episode", "actor", "actress", "director", "hollywood", "concert", "album", "song", "band", "entertainment", "drama", "comedy", "action", "horror", "documentary", "theater", "performance", "star", "fame"],
    "real_estate": ["real estate", "property", "home", "house", "apartment", "condo", "mortgage", "rent", "landlord", "tenant", "real estate", "listing", "realtor", "broker", "investment", "commercial", "residential", "construction", "renovation", "interior", "design", "architecture", "building", "sqft", "bedroom", "bathroom", "yard", "pool", "neighborhood"],
    "science": ["science", "research", "study", "experiment", "scientific", "discovery", "laboratory", "data", "hypothesis", "theory", "physics", "chemistry", "biology", "genetics", "astronomy", "scientist", "peer reviewed", "journal", "publication", "research", "academic", "analysis", "experiment", "laboratory", "scientific", "breakthrough", "innovation", "technology"],
    "politics": ["politics", "government", "election", "vote", "campaign", "policy", "congress", "senate", "president", "political", "democracy", "republican", "democrat", "law", "legislation", "politician", "candidate", "voting", "political", "governance", "public", "officials", "administration", "regulation", "political", "campaign", "rally", "debate", "convention"]
}


def calculate_category_scores(text, categories):
    clean_text = re.sub(r'[^\w\s]', ' ', text.lower())
    words = clean_text.split()
    word_count = Counter(words)

    scores = {}
    for category, keywords in categories.items():
        score = 0
        matched_keywords = []
        for keyword in keywords:
            keyword_words = keyword.split()
            if len(keyword_words) == 1:
                if keyword in word_count:
                    score += word_count[keyword] * 2
                    matched_keywords.append(keyword)
            else:
                phrase = " ".join(keyword_words)
                if phrase in clean_text:
                    score += 3
                    matched_keywords.append(keyword)

        if words:
            scores[category] = round(score / len(words), 4)
        else:
            scores[category] = 0

    return scores


def handler(event):
    try:
        text = event.get("text", "")
        categories_input = event.get("categories")

        if not text:
            return {"ok": False, "error": "text is required"}

        if categories_input:
            if isinstance(categories_input, list):
                categories = {cat: [] for cat in categories_input}
            elif isinstance(categories_input, dict):
                categories = categories_input
            else:
                return {"ok": False, "error": "categories must be a list or dict"}
        else:
            categories = DEFAULT_CATEGORIES

        scores = calculate_category_scores(text, categories)

        sorted_categories = sorted(scores.items(), key=lambda x: x[1], reverse=True)

        primary_category = sorted_categories[0][0] if sorted_categories else None
        secondary_category = sorted_categories[1][0] if len(sorted_categories) > 1 and sorted_categories[1][1] > 0 else None

        confidence = min(sorted_categories[0][1] * 2, 1.0) if sorted_categories else 0

        return {
            "ok": True,
            "primary_category": primary_category,
            "secondary_category": secondary_category,
            "category_scores": dict(sorted_categories),
            "confidence": round(confidence, 2),
            "text_length": len(text.split())
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
