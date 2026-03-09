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
        from pypdf import PdfReader, PdfWriter
    except ImportError:
        try:
            from PyPDF2 import PdfReader, PdfWriter
        except ImportError:
            return {"ok": False, "error": "pypdf is required; install with: pip install pypdf"}

    try:
        raw = base64.b64decode(str(pdf).strip(), validate=True)
        reader = PdfReader(io.BytesIO(raw))
        pages = []
        for page in reader.pages:
            writer = PdfWriter()
            writer.add_page(page)
            buf = io.BytesIO()
            writer.write(buf)
            pages.append(base64.b64encode(buf.getvalue()).decode("ascii"))
        return {"ok": True, "pages": pages, "count": len(pages)}
    except Exception as e:
        return {"ok": False, "error": str(e)}

