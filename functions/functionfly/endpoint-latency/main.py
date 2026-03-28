import urllib.request
import urllib.error
import time
import ssl


def measure_once(url, timeout, ctx):
    start = time.time()
    try:
        with urllib.request.urlopen(url, timeout=timeout, context=ctx) as resp:
            resp.read()
        return round((time.time() - start) * 1000, 2), None
    except Exception as e:
        return round((time.time() - start) * 1000, 2), str(e)


def handler(event):
    """Measure HTTP endpoint latency with multiple samples."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        samples_count = min(10, max(1, int(event.get("samples", 3))))
        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        samples = []
        errors = []
        for _ in range(samples_count):
            latency, err = measure_once(url, timeout, ctx)
            samples.append(latency)
            if err:
                errors.append(err)

        if not samples:
            return {"ok": False, "error": "No successful measurements"}

        return {
            "ok": True,
            "min_ms": min(samples),
            "max_ms": max(samples),
            "avg_ms": round(sum(samples) / len(samples), 2),
            "median_ms": sorted(samples)[len(samples) // 2],
            "samples": samples,
            "errors": errors if errors else None,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
