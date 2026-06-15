"""Tax Calculator - Calculate income tax."""


def handler(event):
    try:
        annual_income = float(event.get("annual_income", 0))
        filing_status = event.get("filing_status", "single").lower()
        state = event.get("state")
        deductions = float(event.get("deductions", 14600 if filing_status == "single" else 29200))

        if annual_income < 0:
            return {"ok": False, "error": "annual_income cannot be negative"}
        if filing_status not in ["single", "married"]:
            return {"ok": False, "error": "filing_status must be single/married"}
        if state is not None and (len(state) != 2 or not state.isalpha()):
            return {"ok": False, "error": "state must be a 2-letter code"}
        if deductions < 0:
            return {"ok": False, "error": "deductions cannot be negative"}

        taxable_income = max(0, annual_income - deductions)

        brackets = [(11600, 0.10), (47150, 0.12), (100525, 0.22), (191950, 0.24),
                    (243725, 0.32), (609350, 0.35), (float('inf'), 0.37)]
        if filing_status == "married":
            brackets = [(23200, 0.10), (94300, 0.12), (201050, 0.22), (383900, 0.24),
                        (487450, 0.32), (731200, 0.35), (float('inf'), 0.37)]

        federal_tax = 0.0
        prev_limit = 0
        marginal_rate = 0.10
        for limit, rate in brackets:
            if taxable_income > limit:
                taxable_in_bracket = limit - prev_limit
                federal_tax += taxable_in_bracket * rate
                prev_limit = limit
            else:
                taxable_in_bracket = taxable_income - prev_limit
                federal_tax += taxable_in_bracket * rate
                marginal_rate = rate
                break

        federal_tax = round(federal_tax, 2)

        result = {
            "ok": True,
            "federal_tax": federal_tax,
            "total_tax": federal_tax,
            "effective_rate": round((federal_tax / annual_income) * 100, 2) if annual_income > 0 else 0,
            "marginal_rate": round(marginal_rate * 100, 1),
            "after_tax_income": round(annual_income - federal_tax, 2)
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
            state_tax = round(annual_income * state_rate, 2)
            result["state_tax"] = state_tax
            result["total_tax"] = round(federal_tax + state_tax, 2)
            result["effective_rate"] = round((result["total_tax"] / annual_income) * 100, 2) if annual_income > 0 else 0
            result["after_tax_income"] = round(annual_income - result["total_tax"], 2)

        return result
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
