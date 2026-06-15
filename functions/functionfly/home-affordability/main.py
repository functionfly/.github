"""Home Affordability Calculator - Calculate mortgage affordability."""


def handler(event):
    try:
        annual_income = float(event.get("annual_income", 0))
        monthly_debt = float(event.get("monthly_debt", 0))
        down_payment = float(event.get("down_payment", 0))
        home_price = float(event.get("home_price", 0))
        loan_term_years = int(event.get("loan_term_years", 30))
        interest_rate = float(event.get("interest_rate", 0))

        if annual_income <= 0:
            return {"ok": False, "error": "annual_income must be positive"}
        if monthly_debt < 0:
            return {"ok": False, "error": "monthly_debt cannot be negative"}
        if down_payment < 0:
            return {"ok": False, "error": "down_payment cannot be negative"}
        if home_price <= 0:
            return {"ok": False, "error": "home_price must be positive"}
        if down_payment >= home_price:
            return {"ok": False, "error": "down_payment must be less than home_price"}
        if loan_term_years not in [10, 15, 20, 25, 30]:
            return {"ok": False, "error": "loan_term_years must be 10, 15, 20, 25, or 30"}
        if interest_rate < 0 or interest_rate > 30:
            return {"ok": False, "error": "interest_rate must be between 0 and 30"}

        loan_amount = home_price - down_payment
        monthly_rate = (interest_rate / 100.0) / 12.0
        num_payments = loan_term_years * 12

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

        max_home_price = home_price
        if housing_ratio > 28 and housing_ratio > 0:
            max_payment = monthly_income * 0.28
            if monthly_rate > 0:
                max_loan = max_payment * ((1 + monthly_rate)**num_payments - 1) / (monthly_rate * (1 + monthly_rate)**num_payments)
            else:
                max_loan = max_payment * num_payments
            max_home_price = round(max_loan + down_payment, 2)

        return {
            "ok": True,
            "monthly_payment": monthly_payment,
            "loan_amount": round(loan_amount, 2),
            "dti_ratio": dti_ratio,
            "housing_ratio": housing_ratio,
            "affordable": affordable,
            "max_home_price": max_home_price
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
