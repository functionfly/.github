import email.header
import email.charset


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    charset = event.get("charset", "utf-8")
    encoding = event.get("encoding", "q")  # 'q' (quoted-printable) or 'b' (base64)

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        enc = "quoted-printable" if encoding.lower() in ("q", "quoted-printable") else "base64"
        cs = email.charset.Charset(charset)
        cs.header_encoding = email.charset.QP if enc == "quoted-printable" else email.charset.BASE64
        header = email.header.Header(str(data), charset=cs)
        result = header.encode()
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
