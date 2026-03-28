import re

def _parse_markdown_table(lines):
    """Parse markdown-style table."""
    rows = []
    for line in lines:
        if "|" in line and not re.match(r"^\s*\|[-:|\s]+\|\s*$", line):
            cells = [c.strip() for c in line.split("|")]
            cells = [c for c in cells if c]
            if cells:
                rows.append(cells)
    return rows

def _parse_csv_table(lines):
    """Parse CSV-style table."""
    rows = []
    for line in lines:
        if "," in line:
            cells = [c.strip().strip('"') for c in line.split(",")]
            if cells:
                rows.append(cells)
    return rows

def _parse_tsv_table(lines):
    """Parse TSV-style table."""
    rows = []
    for line in lines:
        if "\t" in line:
            cells = [c.strip() for c in line.split("\t")]
            if cells:
                rows.append(cells)
    return rows

def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        fmt = event.get("format", "auto")
        lines = text.strip().split("\n")
        rows = []
        detected_format = fmt
        if fmt == "auto":
            if any("|" in l for l in lines):
                rows = _parse_markdown_table(lines)
                detected_format = "markdown"
            elif any("\t" in l for l in lines):
                rows = _parse_tsv_table(lines)
                detected_format = "tsv"
            elif any("," in l for l in lines):
                rows = _parse_csv_table(lines)
                detected_format = "csv"
        elif fmt == "markdown":
            rows = _parse_markdown_table(lines)
        elif fmt == "csv":
            rows = _parse_csv_table(lines)
        elif fmt == "tsv":
            rows = _parse_tsv_table(lines)
        if not rows:
            return {"ok": True, "result": [], "tables": [], "count": 0}
        # First row as header
        header = rows[0] if rows else []
        data_rows = rows[1:] if len(rows) > 1 else []
        table = {
            "headers": header,
            "rows": data_rows,
            "row_count": len(data_rows),
            "column_count": len(header),
            "format": detected_format
        }
        # Convert to list of dicts
        records = []
        for row in data_rows:
            record = {}
            for j, cell in enumerate(row):
                key = header[j] if j < len(header) else f"col_{j}"
                record[key] = cell
            records.append(record)
        table["records"] = records
        return {"ok": True, "result": [table], "tables": [table], "count": 1}
    except Exception as e:
        return {"ok": False, "error": str(e)}
