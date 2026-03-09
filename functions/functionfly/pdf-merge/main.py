import base64
import io


def handler(event):
    if isinstance(event, dict):
        pdfs = event.get("pdfs", event.get("documents", []))
    else:
        pdfs = []

    if not pdfs or not isinstance(pdfs, list):
        return {"ok": False, "error": "Input 'pdfs' must be a non-empty array of base64 PDF strings"}

    try:
        from pypdf import PdfReader, PdfWriter
    except ImportError:
        try:
            from PyPDF2 import PdfReader, PdfWriter
        except ImportError:
            return {"ok": False, "error": "pypdf is required; install with: pip install pypdf"}

    try:
        writer = PdfWriter()
        for b64 in pdfs:
            raw = base64.b64decode(str(b64).strip(), validate=True)
            reader = PdfReader(io.BytesIO(raw))
            for page in reader.pages:
                writer.add_page(page)
        buf = io.BytesIO()
        writer.write(buf)
        out_b64 = base64.b64encode(buf.getvalue()).decode("ascii")
        return {"ok": True, "pdf_base64": out_b64, "pages_total": len(writer.pages)}
    except Exception as e:
        return {"ok": False, "error": str(e)}

