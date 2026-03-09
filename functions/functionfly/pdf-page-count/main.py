import base64
import io


def handler(event):
    if isinstance(event, dict):
        pdf = event.get("pdf", event.get("data", ""))
    else:
        pdf = ""

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
        return {"ok": True, "pages": len(reader.pages)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
