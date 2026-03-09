import base64
import io


def handler(event):
    if isinstance(event, dict):
        pdf = event.get("pdf", event.get("data", ""))
        page_start = event.get("page_start", 1)
        page_end = event.get("page_end")
    else:
        pdf = ""
        page_start = 1
        page_end = None

    if not pdf:
        return {"ok": False, "error": "Input 'pdf' is required"}

    try:
        from pypdf import PdfReader
    except ImportError:
        try:
            from PyPDF2 import PdfReader
        except ImportError:
            return {"ok": False, "error": "pypdf is required; install with: pip install pypdf"}

    try:
        raw = base64.b64decode(str(pdf).strip(), validate=True)
        reader = PdfReader(io.BytesIO(raw))
        total = len(reader.pages)
        i_start = max(0, int(page_start) - 1)
        i_end = min(total, int(page_end)) if page_end is not None else total
        i_end = max(i_start, i_end)
        parts = []
        for i in range(i_start, i_end):
            parts.append(reader.pages[i].extract_text() or "")
        text = "\n\n".join(parts)
        return {"ok": True, "text": text, "pages": i_end - i_start}
    except Exception as e:
        return {"ok": False, "error": str(e)}

