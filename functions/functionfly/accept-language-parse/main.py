def handler(event):
    accept_language = event.get("accept_language")

    if not accept_language:
        return {"ok": False, "error": "accept_language is required"}

    if not isinstance(accept_language, str):
        return {"ok": False, "error": "accept_language must be a string"}

    try:
        languages = []
        quality_values = {}

        # Parse Accept-Language header
        # Format: language-tag;q=quality-value
        parts = accept_language.split(',')

        for part in parts:
            part = part.strip()
            if not part:
                continue

            # Check for quality value
            if ';q=' in part:
                lang, qvalue = part.split(';q=', 1)
                try:
                    quality = float(qvalue.strip())
                    if quality < 0 or quality > 1:
                        quality = 1.0  # Default to 1.0 for invalid values
                except ValueError:
                    quality = 1.0  # Default to 1.0 for invalid values
            else:
                lang = part
                quality = 1.0

            lang = lang.strip()
            if lang:
                languages.append(lang)
                quality_values[lang] = quality

        # Sort by quality value (highest first)
        sorted_languages = sorted(languages, key=lambda x: quality_values[x], reverse=True)

        # Extract language and country codes
        parsed_languages = []
        for lang in languages:
            lang_parts = lang.split('-', 1)
            primary = lang_parts[0].lower()
            country = lang_parts[1].upper() if len(lang_parts) > 1 else None

            parsed_languages.append({
                "full": lang,
                "primary": primary,
                "country": country,
                "quality": quality_values[lang]
            })

        result = {
            "languages": parsed_languages,
            "preferred_language": sorted_languages[0] if sorted_languages else None,
            "language_count": len(languages),
            "has_country_specifier": any(lang.get("country") for lang in parsed_languages),
            "quality_sorted": sorted_languages
        }

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to parse Accept-Language: {str(e)}"}