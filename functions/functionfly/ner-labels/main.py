NER_SCHEMAS = {
    "standard": {
        "PERSON": {"description": "People, including fictional characters", "examples": ["Albert Einstein", "Sherlock Holmes", "Marie Curie"]},
        "ORGANIZATION": {"description": "Companies, agencies, institutions, and other groups", "examples": ["Google", "United Nations", "Harvard University"]},
        "LOCATION": {"description": "Countries, cities, states, mountains, bodies of water", "examples": ["New York", "Mount Everest", "Pacific Ocean"]},
        "DATE": {"description": "Absolute or relative dates or periods", "examples": ["January 1, 2024", "last year", "the 1990s"]},
        "TIME": {"description": "Times smaller than a day", "examples": ["3:00 PM", "noon", "midnight"]},
        "MONEY": {"description": "Monetary values, including unit", "examples": ["$100", "€50 million", "five dollars"]},
        "PERCENT": {"description": "Percentage, including '%'", "examples": ["20%", "fifty percent", "0.5%"]},
        "MISC": {"description": "Miscellaneous entities that don't fit other categories", "examples": ["Nobel Prize", "World Cup", "COVID-19"]},
    },
    "ontonotes": {
        "PERSON": {"description": "People, including fictional", "examples": ["Barack Obama", "Harry Potter"]},
        "NORP": {"description": "Nationalities, religious or political groups", "examples": ["American", "Buddhist", "Republican"]},
        "FAC": {"description": "Buildings, airports, highways, bridges", "examples": ["Eiffel Tower", "JFK Airport", "Golden Gate Bridge"]},
        "ORG": {"description": "Companies, agencies, institutions", "examples": ["Apple Inc.", "FBI", "MIT"]},
        "GPE": {"description": "Countries, cities, states", "examples": ["France", "Tokyo", "California"]},
        "LOC": {"description": "Non-GPE locations, mountain ranges, bodies of water", "examples": ["Alps", "Amazon River", "Sahara Desert"]},
        "PRODUCT": {"description": "Objects, vehicles, foods, etc.", "examples": ["iPhone", "Tesla Model 3", "Coca-Cola"]},
        "EVENT": {"description": "Named hurricanes, battles, wars, sports events", "examples": ["World War II", "Super Bowl", "Hurricane Katrina"]},
        "WORK_OF_ART": {"description": "Titles of books, songs, etc.", "examples": ["The Great Gatsby", "Bohemian Rhapsody", "Mona Lisa"]},
        "LAW": {"description": "Named documents made into laws", "examples": ["First Amendment", "GDPR", "Roe v. Wade"]},
        "LANGUAGE": {"description": "Any named language", "examples": ["English", "Mandarin", "Spanish"]},
        "DATE": {"description": "Absolute or relative dates or periods", "examples": ["January 2024", "last week", "the Renaissance"]},
        "TIME": {"description": "Times smaller than a day", "examples": ["2:30 PM", "morning", "at dawn"]},
        "PERCENT": {"description": "Percentage", "examples": ["15%", "half", "a third"]},
        "MONEY": {"description": "Monetary values", "examples": ["$500", "€1 billion", "ten euros"]},
        "QUANTITY": {"description": "Measurements, as of weight or distance", "examples": ["5 kg", "100 miles", "3 liters"]},
        "ORDINAL": {"description": "First, second, etc.", "examples": ["first", "3rd", "twenty-first"]},
        "CARDINAL": {"description": "Numerals that do not fall under another type", "examples": ["one", "42", "a dozen"]},
    },
    "conll": {
        "PER": {"description": "Person names", "examples": ["John Smith", "Angela Merkel"]},
        "ORG": {"description": "Organization names", "examples": ["Microsoft", "Red Cross"]},
        "LOC": {"description": "Location names", "examples": ["London", "Mount Fuji"]},
        "MISC": {"description": "Miscellaneous named entities", "examples": ["English", "World Cup", "Nobel Prize"]},
    }
}


def handler(event):
    if not isinstance(event, dict):
        event = {}
    try:
        schema = event.get("schema", "standard")
        if schema not in NER_SCHEMAS:
            schema = "standard"
        labels = NER_SCHEMAS[schema]
        label_list = [
            {"label": label, "description": info["description"], "examples": info["examples"]}
            for label, info in labels.items()
        ]
        return {
            "ok": True,
            "result": label_list,
            "labels": label_list,
            "schema": schema,
            "available_schemas": list(NER_SCHEMAS.keys()),
            "count": len(label_list)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
