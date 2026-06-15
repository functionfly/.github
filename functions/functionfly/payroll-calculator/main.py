"""Payroll Calculator - Calculate payroll with taxes."""


def handler(event):
    try:
        gross_salary = float(event.get("gross_salary", 0))
        pay_period = event.get("pay_period", "monthly").lower()
        filing_status = event.get("filing_status", "single").lower()
        allowances = int(event.get("allowances", 0))
        state = event.get("state")

        if gross_salary <= 0:
            return {"ok": False, "error": "gross_salary must be positive"}
        if pay_period not in ["weekly", "biweekly", "monthly"]:
            return {"ok": False, "error": "pay_period must be weekly/biweekly/monthly"}
        if filing_status not in ["single", "married"]:
            return {"ok": False, "error": "filing_status must be single/married"}
        if allowances < 0:
            return {"ok": False, "error": "allowances cannot be negative"}
        if state is not None and (len(state) != 2 or not state.isalpha()):
            return {"ok": False, "error": "state must be a 2-letter code"}

        periods_per_year = {"weekly": 52, "biweekly": 26, "monthly": 12}
        periods = periods_per_year[pay_period]
        annual_gross = gross_salary * periods

        standard_deduction = 14600 if filing_status == "single" else 29200
        taxable_income = max(0, annual_gross - standard_deduction - (allowances * 4300))

        brackets = [(11600, 0.10), (47150, 0.12), (100525, 0.22), (191950, 0.24),
                    (243725, 0.32), (609350, 0.35), (float('inf'), 0.37)]
        if filing_status == "married":
            brackets = [(23200, 0.10), (94300, 0.12), (201050, 0.22), (383900, 0.24),
                        (487450, 0.32), (731200, 0.35), (float('inf'), 0.37)]

        federal_tax = 0.0
        prev_limit = 0
        for limit, rate in brackets:
            if taxable_income > prev_limit:
                taxable_in_bracket = min(taxable_income, limit) - prev_limit
                federal_tax += taxable_in_bracket * rate
            prev_limit = limit
        federal_tax = round(federal_tax / periods, 2)

        social_security = round(min(annual_gross, 168600) * 0.062, 2)
        medicare = round(annual_gross * 0.0145, 2)
        if annual_gross > 200000:
            medicare += round((annual_gross - 200000) * 0.0035, 2)

        result = {
            "ok": True,
            "federal_tax": federal_tax,
            "social_security": social_security,
            "medicare": medicare,
            "net_salary": round(gross_salary - federal_tax - social_security - medicare, 2),
            "effective_tax_rate": 0.0
        }

        if state is not None:
            state = state.upper()
            state_taxes = {
                "AL": 0.05, "AK": 0.0, "AZ": 0.025, "AR": 0.044, "CA": 0.0725,
                "CO": 0.044, "CT": 0.0699, "DE": 0.066, "FL": 0.0, "GA": 0.0549,
                "HI": 0.0825, "ID": 0.058, "IL": 0.0495, "IN": 0.0317, "IA": 0.06,
                "KS": 0.057, "KY": 0.045, "LA": 0.0425, "ME": 0.0715, "MD": 0.0575,
                "MA": 0.05, "MI": 0.0425, "MN": 0.0985, "MS": 0.05, "MO": 0.0495,
                "MT": 0.059, "NE": 0.0664, "NV": 0.0, "NH": 0.0, "NJ": 0.0897,
                "NM": 0.059, "NY": 0.0685, "NC": 0.0525, "ND": 0.029, "OH": 0.04,
                "OK": 0.0475, "OR": 0.099, "PA": 0.0307, "RI": 0.0599, "SC": 0.065,
                "SD": 0.0, "TN": 0.0, "TX": 0.0, "UT": 0.0485, "VT": 0.0875,
                "VA": 0.0575, "WA": 0.0, "WV": 0.065, "WI": 0.0765, "WY": 0.0
            }
            state_rate = state_taxes.get(state, 0.05)
            state_tax = round(annual_gross * state_rate / periods, 2)
            result["state_tax"] = state_tax
            result["net_salary"] = round(result["net_salary"] - state_tax, 2)

        total_tax = federal_tax + social_security + medicare
        if "state_tax" in result:
            total_tax += result["state_tax"]
        result["effective_tax_rate"] = round((total_tax / gross_salary) * 100, 2) if gross_salary > 0 else 0

        return result
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
