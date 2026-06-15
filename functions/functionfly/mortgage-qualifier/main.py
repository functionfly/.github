"""Mortgage Qualifier - Qualify for mortgage loan."""


def handler(event):
    try:
        annual_income = float(event.get("annual_income", 0))
        monthly_debt = float(event.get("monthly_debt", 0))
        loan_amount = float(event.get("loan_amount", 0))
        interest_rate = float(event.get("interest_rate", 0))
        term_years = int(event.get("term_years", 30))

        if annual_income <= 0:
            return {"ok": False, "error": "annual_income must be positive"}
        if monthly_debt < 0:
            return {"ok": False, "error": "monthly_debt cannot be negative"}
        if loan_amount <= 0:
            return {"ok": False, "error": "loan_amount must be positive"}
        if interest_rate < 0 or interest_rate > 30:
            return {"ok": False, "error": "interest_rate must be between 0 and 30"}
        if term_years <= 0 or term_years > 50:
            return {"ok": False, "error": "term_years must be between 1 and 50"}

        monthly_rate = (interest_rate / 100.0) / 12.0
        num_payments = term_years * 12

        if monthly_rate > 0:
            monthly_payment = loan_amount * (monthly_rate * (1 + monthly_rate)**num_payments) / ((1 + monthly_rate)**num_payments - 1)
        else:
            monthly_payment = loan_amount / num_payments
        monthly_payment = round(monthly_payment, 2)

        total_monthly_debt = monthly_debt + monthly_payment
        monthly_income = annual_income / 12.0

        dti_ratio = round((total_monthly_debt / monthly_income) * 100, 2)
        housing_ratio = round((monthly_payment / monthly_income) * 100, 2)

        affordable = dti_ratio <= 43 and housing_ratio <= 28

        required_income = round((total_monthly_debt / 0.43) * 12, 2) if dti_ratio > 0 else annual_income

        max_loan = loan_amount
        if housing_ratio > 28 and housing_ratio > 0:
            max_payment = monthly_income * 0.28
            if monthly_rate > 0:
                max_loan = max_payment * ((1 + monthly_rate)**num_payments - 1) / (monthly_rate * (1 + monthly_rate)**num_payments)
            else:
                max_loan = max_payment * num_payments
            max_loan = round(max_loan, 2)

        return {
            "ok": True,
            "affordable": affordable,
            "dti_ratio": dti_ratio,
            "housing_ratio": housing_ratio,
            "required_income": required_income,
            "max_loan": max_loan
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
