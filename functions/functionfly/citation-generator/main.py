import re
from datetime import datetime
from typing import Any


def format_apa(source_type: str, title: str, author: str, url: str = None, publish_date: str = None) -> str:
    parts = []
    
    if author:
        author_parts = author.split()
        if len(author_parts) >= 2:
            last_name = author_parts[-1]
            initials = " ".join(f"{n[0]}." for n in author_parts[:-1])
            parts.append(f"{last_name}, {initials}")
        else:
            parts.append(author)
    
    if publish_date:
        try:
            if len(publish_date) == 4 and publish_date.isdigit():
                parts.append(publish_date)
            else:
                for fmt in ["%Y-%m-%d", "%Y/%m/%d", "%d %B %Y", "%B %d, %Y"]:
                    try:
                        dt = datetime.strptime(publish_date, fmt)
                        parts.append(dt.strftime("%Y, %B %d"))
                        break
                    except ValueError:
                        continue
                else:
                    parts.append(f"({publish_date})")
        except Exception:
            parts.append(f"({publish_date})")
    else:
        parts.append("n.d.")
    
    if title:
        parts.append(f"{title}.")
    
    if url:
        parts.append(f"Retrieved from {url}")
    else:
        parts.append("[Source details incomplete]")
    
    return " ".join(parts)


def format_mla(source_type: str, title: str, author: str, url: str = None, publish_date: str = None) -> str:
    parts = []
    
    if author:
        author_parts = author.split()
        if len(author_parts) >= 2:
            last_name = author_parts[-1]
            first_name = " ".join(author_parts[:-1])
            parts.append(f"{last_name}, {first_name}")
        else:
            parts.append(author)
    else:
        parts.append("Anonymous")
    
    if title:
        parts.append(f'"{title}."")
    
    if source_type == "book":
        if publish_date:
            parts.append(f"Published: {publish_date},")
    else:
        if publish_date:
            parts.append(f"{publish_date},")
    
    if url:
        parts.append(f"{url}.")
    
    return ", ".join(parts).replace(", ,", ",").replace(" ,", ",").replace(".,", ".").replace("..", ".")


def format_chicago(source_type: str, title: str, author: str, url: str = None, publish_date: str = None) -> str:
    parts = []
    
    if author:
        author_parts = author.split()
        if len(author_parts) >= 2:
            last_name = author_parts[-1]
            first_name = " ".join(author_parts[:-1])
            parts.append(f"{last_name}, {first_name}")
        else:
            parts.append(author)
    
    if title:
        if source_type == "book":
            parts.append(f"{title}.")
        else:
            parts.append(f'"{title}."')
    
    if publish_date:
        parts.append(f"({publish_date}).")
    
    if url:
        parts.append(f"Accessed: {url}")
    
    return " ".join(parts)


def generate_bibliography_entry(source_type: str, title: str, author: str, url: str = None, publish_date: str = None, format_style: str = "apa") -> str:
    if format_style.lower() == "apa":
        return format_apa(source_type, title, author, url, publish_date)
    elif format_style.lower() == "mla":
        return format_mla(source_type, title, author, url, publish_date)
    elif format_style.lower() == "chicago":
        return format_chicago(source_type, title, author, url, publish_date)
    else:
        return format_apa(source_type, title, author, url, publish_date)


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        source_type = event.get("source_type", "").lower().strip()
        title = event.get("title", "")
        author = event.get("author", "")
        url = event.get("url")
        publish_date = event.get("publish_date")
        format_style = event.get("format", "apa").lower().strip()
        
        if not source_type:
            return {"ok": False, "error": "source_type is required (web/book/article)"}
        
        valid_source_types = ["web", "book", "article", "journal", "magazine", "newspaper"]
        if source_type not in valid_source_types:
            return {"ok": False, "error": f"source_type must be one of: {', '.join(valid_source_types)}"}
        
        if not title:
            return {"ok": False, "error": "title is required"}
        
        if not author:
            return {"ok": False, "error": "author is required"}
        
        valid_formats = ["apa", "mla", "chicago"]
        if format_style not in valid_formats:
            return {"ok": False, "error": f"format must be one of: {', '.join(valid_formats)}"}
        
        citation = generate_bibliography_entry(source_type, title, author, url, publish_date, format_style)
        
        bibliography_entry = citation
        
        return {
            "ok": True,
            "citation": citation,
            "bibliography_entry": bibliography_entry,
            "source_type": source_type,
            "format": format_style
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
