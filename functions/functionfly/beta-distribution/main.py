import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    x = event.get("x")
    alpha = event.get("alpha")
    beta = event.get("beta")
    if x is None or alpha is None or beta is None:
        return {"ok": False, "error": "x, alpha, and beta are required"}
    try:
        x = float(x)
        a = float(alpha)
        b = float(beta)
        if not (0 <= x <= 1):
            return {"ok": False, "error": "x must be between 0 and 1"}
        if a <= 0 or b <= 0:
            return {"ok": False, "error": "alpha and beta must be positive"}
        # PDF using log-gamma
        log_pdf = (math.lgamma(a + b) - math.lgamma(a) - math.lgamma(b)
                   + (a - 1) * math.log(x) if x > 0 else float('-inf')
                   + (b - 1) * math.log(1 - x) if x < 1 else float('-inf'))
        # Regularized incomplete beta function (CDF) via continued fraction
        def beta_cdf(x, a, b):
            if x == 0:
                return 0.0
            if x == 1:
                return 1.0
            lbeta = math.lgamma(a) + math.lgamma(b) - math.lgamma(a + b)
            log_x = a * math.log(x) + b * math.log(1 - x) - lbeta
            # Use regularized incomplete beta via series
            # Simple approximation using scipy-like continued fraction
            def betainc(a, b, x):
                if x < (a + 1) / (a + b + 2):
                    return _betacf(a, b, x) * math.exp(log_x) / a
                else:
                    return 1 - _betacf(b, a, 1 - x) * math.exp(b * math.log(1 - x) + a * math.log(x) - lbeta) / b
            def _betacf(a, b, x):
                MAXIT = 200
                EPS = 3e-7
                qab = a + b
                qap = a + 1
                qam = a - 1
                c = 1.0
                d = 1.0 - qab * x / qap
                if abs(d) < 1e-30:
                    d = 1e-30
                d = 1.0 / d
                h = d
                for m in range(1, MAXIT + 1):
                    m2 = 2 * m
                    aa = m * (b - m) * x / ((qam + m2) * (a + m2))
                    d = 1.0 + aa * d
                    if abs(d) < 1e-30:
                        d = 1e-30
                    c = 1.0 + aa / c
                    if abs(c) < 1e-30:
                        c = 1e-30
                    d = 1.0 / d
                    h *= d * c
                    aa = -(a + m) * (qab + m) * x / ((a + m2) * (qap + m2))
                    d = 1.0 + aa * d
                    if abs(d) < 1e-30:
                        d = 1e-30
                    c = 1.0 + aa / c
                    if abs(c) < 1e-30:
                        c = 1e-30
                    d = 1.0 / d
                    delta = d * c
                    h *= delta
                    if abs(delta - 1.0) < EPS:
                        break
                return h
            return betainc(a, b, x)
        if x == 0 or x == 1:
            pdf = 0.0
        else:
            pdf = math.exp(math.lgamma(a + b) - math.lgamma(a) - math.lgamma(b)
                           + (a - 1) * math.log(x) + (b - 1) * math.log(1 - x))
        cdf = beta_cdf(x, a, b)
        return {"ok": True, "result": {"pdf": round(pdf, 6), "cdf": round(cdf, 6)}}
    except Exception as e:
        return {"ok": False, "error": str(e)}
