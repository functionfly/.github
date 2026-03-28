#!/usr/bin/env python3
"""Generate all 75 Finance & Analytics functions for FunctionFly."""

import os
import json

BASE_DIR = "functions/functionfly"

FUNCTIONS = [
    # (name, title, description, category, tags, input_props, required, example_input, example_output_extra)
    (
        "pv-calculate",
        "Present Value Calculate",
        "Calculate the present value of a future sum of money given a discount rate and number of periods.",
        "finance",
        ["present-value", "finance", "time-value-of-money"],
        {
            "future_value": {"type": "number", "description": "Future value amount", "default": 0},
            "rate": {"type": "number", "description": "Interest rate per period (decimal, e.g. 0.05)"},
            "periods": {"type": "number", "description": "Number of periods"},
            "payment": {"type": "number", "description": "Periodic payment (default 0)", "default": 0}
        },
        ["rate", "periods"],
        {"future_value": 1000, "rate": 0.05, "periods": 10},
        {"result": 613.91}
    ),
    (
        "fv-calculate",
        "Future Value Calculate",
        "Calculate the future value of an investment given present value, interest rate, and number of periods.",
        "finance",
        ["future-value", "finance", "investment"],
        {
            "present_value": {"type": "number", "description": "Present value amount"},
            "rate": {"type": "number", "description": "Interest rate per period (decimal)"},
            "periods": {"type": "number", "description": "Number of periods"},
            "payment": {"type": "number", "description": "Periodic payment (default 0)", "default": 0}
        },
        ["present_value", "rate", "periods"],
        {"present_value": 1000, "rate": 0.05, "periods": 10},
        {"result": 1628.89}
    ),
    (
        "npv-calculate",
        "Net Present Value Calculate",
        "Calculate the net present value of a series of cash flows discounted at a given rate.",
        "finance",
        ["npv", "net-present-value", "finance", "investment"],
        {
            "rate": {"type": "number", "description": "Discount rate per period (decimal)"},
            "cash_flows": {"type": "array", "items": {"type": "number"}, "description": "List of cash flows (first is initial investment, negative)"}
        },
        ["rate", "cash_flows"],
        {"rate": 0.1, "cash_flows": [-1000, 300, 400, 500]},
        {"result": 8.43}
    ),
    (
        "irr-calculate",
        "Internal Rate of Return Calculate",
        "Calculate the internal rate of return for a series of cash flows.",
        "finance",
        ["irr", "internal-rate-of-return", "finance"],
        {
            "cash_flows": {"type": "array", "items": {"type": "number"}, "description": "List of cash flows (first is initial investment, negative)"},
            "guess": {"type": "number", "description": "Initial guess for IRR (default 0.1)", "default": 0.1}
        },
        ["cash_flows"],
        {"cash_flows": [-1000, 300, 400, 500]},
        {"result": 0.1234}
    ),
    (
        "pmt-calculate",
        "Payment Amount Calculate",
        "Calculate the periodic payment for a loan or annuity given rate, number of periods, and present value.",
        "finance",
        ["payment", "pmt", "loan", "annuity"],
        {
            "rate": {"type": "number", "description": "Interest rate per period (decimal)"},
            "periods": {"type": "number", "description": "Number of periods"},
            "present_value": {"type": "number", "description": "Present value (loan amount)"},
            "future_value": {"type": "number", "description": "Future value (default 0)", "default": 0}
        },
        ["rate", "periods", "present_value"],
        {"rate": 0.005, "periods": 360, "present_value": 200000},
        {"result": -1199.1}
    ),
    (
        "nper-calculate",
        "Number of Periods Calculate",
        "Calculate the number of periods required to pay off a loan or reach a savings goal.",
        "finance",
        ["nper", "periods", "loan", "savings"],
        {
            "rate": {"type": "number", "description": "Interest rate per period (decimal)"},
            "payment": {"type": "number", "description": "Periodic payment amount"},
            "present_value": {"type": "number", "description": "Present value"},
            "future_value": {"type": "number", "description": "Future value (default 0)", "default": 0}
        },
        ["rate", "payment", "present_value"],
        {"rate": 0.005, "payment": -1199.1, "present_value": 200000},
        {"result": 360.0}
    ),
    (
        "rate-calculate",
        "Interest Rate Calculate",
        "Calculate the interest rate per period for a loan or investment.",
        "finance",
        ["rate", "interest-rate", "loan", "finance"],
        {
            "periods": {"type": "number", "description": "Number of periods"},
            "payment": {"type": "number", "description": "Periodic payment amount"},
            "present_value": {"type": "number", "description": "Present value"},
            "future_value": {"type": "number", "description": "Future value (default 0)", "default": 0},
            "guess": {"type": "number", "description": "Initial guess (default 0.01)", "default": 0.01}
        },
        ["periods", "payment", "present_value"],
        {"periods": 360, "payment": -1199.1, "present_value": 200000},
        {"result": 0.005}
    ),
    (
        "loan-payment",
        "Loan Payment Calculate",
        "Calculate the monthly payment for a loan given principal, annual interest rate, and term in years.",
        "finance",
        ["loan", "payment", "mortgage", "finance"],
        {
            "principal": {"type": "number", "description": "Loan principal amount"},
            "annual_rate": {"type": "number", "description": "Annual interest rate (decimal, e.g. 0.06)"},
            "years": {"type": "number", "description": "Loan term in years"}
        },
        ["principal", "annual_rate", "years"],
        {"principal": 200000, "annual_rate": 0.06, "years": 30},
        {"result": 1199.1, "total_payment": 431676.0, "total_interest": 231676.0}
    ),
    (
        "amortization-schedule",
        "Amortization Schedule",
        "Generate a full amortization schedule for a loan showing principal, interest, and balance for each period.",
        "finance",
        ["amortization", "loan", "schedule", "finance"],
        {
            "principal": {"type": "number", "description": "Loan principal amount"},
            "annual_rate": {"type": "number", "description": "Annual interest rate (decimal)"},
            "years": {"type": "number", "description": "Loan term in years"},
            "periods_per_year": {"type": "number", "description": "Payment periods per year (default 12)", "default": 12}
        },
        ["principal", "annual_rate", "years"],
        {"principal": 10000, "annual_rate": 0.06, "years": 1},
        {"result": []}
    ),
    (
        "mortgage-payment",
        "Mortgage Payment Calculate",
        "Calculate monthly mortgage payment including principal and interest.",
        "finance",
        ["mortgage", "payment", "real-estate", "finance"],
        {
            "home_price": {"type": "number", "description": "Home purchase price"},
            "down_payment": {"type": "number", "description": "Down payment amount"},
            "annual_rate": {"type": "number", "description": "Annual interest rate (decimal)"},
            "years": {"type": "number", "description": "Loan term in years"}
        },
        ["home_price", "down_payment", "annual_rate", "years"],
        {"home_price": 300000, "down_payment": 60000, "annual_rate": 0.065, "years": 30},
        {"result": 1516.42, "loan_amount": 240000}
    ),
    (
        "investment-return",
        "Investment Return Calculate",
        "Calculate the total return and annualized return on an investment.",
        "finance",
        ["investment", "return", "roi", "finance"],
        {
            "initial_value": {"type": "number", "description": "Initial investment value"},
            "final_value": {"type": "number", "description": "Final investment value"},
            "years": {"type": "number", "description": "Investment period in years", "default": 1}
        },
        ["initial_value", "final_value"],
        {"initial_value": 10000, "final_value": 15000, "years": 5},
        {"result": 0.5, "annualized_return": 0.0845}
    ),
    (
        "roi-calculate",
        "Return on Investment Calculate",
        "Calculate return on investment (ROI) as a percentage.",
        "finance",
        ["roi", "return-on-investment", "finance"],
        {
            "gain": {"type": "number", "description": "Net gain from investment"},
            "cost": {"type": "number", "description": "Cost of investment"}
        },
        ["gain", "cost"],
        {"gain": 5000, "cost": 10000},
        {"result": 0.5, "result_pct": 50.0}
    ),
    (
        "cagr-calculate",
        "Compound Annual Growth Rate Calculate",
        "Calculate the compound annual growth rate (CAGR) of an investment.",
        "finance",
        ["cagr", "growth-rate", "investment", "finance"],
        {
            "beginning_value": {"type": "number", "description": "Beginning value"},
            "ending_value": {"type": "number", "description": "Ending value"},
            "years": {"type": "number", "description": "Number of years"}
        },
        ["beginning_value", "ending_value", "years"],
        {"beginning_value": 10000, "ending_value": 20000, "years": 5},
        {"result": 0.1487}
    ),
    (
        "compound-growth",
        "Compound Growth Calculate",
        "Calculate the future value using compound growth formula.",
        "finance",
        ["compound-growth", "finance", "investment"],
        {
            "initial_value": {"type": "number", "description": "Initial value"},
            "growth_rate": {"type": "number", "description": "Growth rate per period (decimal)"},
            "periods": {"type": "number", "description": "Number of periods"}
        },
        ["initial_value", "growth_rate", "periods"],
        {"initial_value": 1000, "growth_rate": 0.1, "periods": 5},
        {"result": 1610.51}
    ),
    (
        "depreciation-straight-line",
        "Straight-Line Depreciation",
        "Calculate straight-line depreciation for an asset.",
        "finance",
        ["depreciation", "straight-line", "accounting", "finance"],
        {
            "cost": {"type": "number", "description": "Asset cost"},
            "salvage_value": {"type": "number", "description": "Salvage value at end of life"},
            "useful_life": {"type": "number", "description": "Useful life in years"}
        },
        ["cost", "salvage_value", "useful_life"],
        {"cost": 10000, "salvage_value": 1000, "useful_life": 5},
        {"result": 1800.0, "annual_depreciation": 1800.0}
    ),
    (
        "depreciation-ddb",
        "Double Declining Balance Depreciation",
        "Calculate depreciation using the double declining balance method.",
        "finance",
        ["depreciation", "ddb", "double-declining", "accounting"],
        {
            "cost": {"type": "number", "description": "Asset cost"},
            "salvage_value": {"type": "number", "description": "Salvage value at end of life"},
            "useful_life": {"type": "number", "description": "Useful life in years"},
            "period": {"type": "number", "description": "Period to calculate depreciation for (1-based)"}
        },
        ["cost", "salvage_value", "useful_life", "period"],
        {"cost": 10000, "salvage_value": 1000, "useful_life": 5, "period": 1},
        {"result": 4000.0}
    ),
    (
        "depreciation-sum-of-years",
        "Sum of Years Digits Depreciation",
        "Calculate depreciation using the sum of years digits method.",
        "finance",
        ["depreciation", "sum-of-years", "accounting", "finance"],
        {
            "cost": {"type": "number", "description": "Asset cost"},
            "salvage_value": {"type": "number", "description": "Salvage value at end of life"},
            "useful_life": {"type": "number", "description": "Useful life in years"},
            "period": {"type": "number", "description": "Period to calculate depreciation for (1-based)"}
        },
        ["cost", "salvage_value", "useful_life", "period"],
        {"cost": 10000, "salvage_value": 1000, "useful_life": 5, "period": 1},
        {"result": 3000.0}
    ),
    (
        "break-even",
        "Break-Even Analysis",
        "Calculate the break-even point in units and revenue.",
        "finance",
        ["break-even", "cost-analysis", "finance", "accounting"],
        {
            "fixed_costs": {"type": "number", "description": "Total fixed costs"},
            "variable_cost_per_unit": {"type": "number", "description": "Variable cost per unit"},
            "price_per_unit": {"type": "number", "description": "Selling price per unit"}
        },
        ["fixed_costs", "variable_cost_per_unit", "price_per_unit"],
        {"fixed_costs": 10000, "variable_cost_per_unit": 5, "price_per_unit": 15},
        {"result": 1000.0, "break_even_units": 1000.0, "break_even_revenue": 15000.0}
    ),
    (
        "profit-margin",
        "Profit Margin Calculate",
        "Calculate net profit margin as a percentage of revenue.",
        "finance",
        ["profit-margin", "profitability", "finance"],
        {
            "net_profit": {"type": "number", "description": "Net profit"},
            "revenue": {"type": "number", "description": "Total revenue"}
        },
        ["net_profit", "revenue"],
        {"net_profit": 20000, "revenue": 100000},
        {"result": 0.2, "result_pct": 20.0}
    ),
    (
        "gross-margin",
        "Gross Margin Calculate",
        "Calculate gross margin as a percentage of revenue.",
        "finance",
        ["gross-margin", "profitability", "finance"],
        {
            "revenue": {"type": "number", "description": "Total revenue"},
            "cost_of_goods_sold": {"type": "number", "description": "Cost of goods sold (COGS)"}
        },
        ["revenue", "cost_of_goods_sold"],
        {"revenue": 100000, "cost_of_goods_sold": 60000},
        {"result": 0.4, "result_pct": 40.0, "gross_profit": 40000.0}
    ),
    (
        "operating-margin",
        "Operating Margin Calculate",
        "Calculate operating margin (EBIT / Revenue) as a percentage.",
        "finance",
        ["operating-margin", "profitability", "ebit", "finance"],
        {
            "operating_income": {"type": "number", "description": "Operating income (EBIT)"},
            "revenue": {"type": "number", "description": "Total revenue"}
        },
        ["operating_income", "revenue"],
        {"operating_income": 15000, "revenue": 100000},
        {"result": 0.15, "result_pct": 15.0}
    ),
    (
        "ebitda",
        "EBITDA Calculate",
        "Calculate EBITDA (Earnings Before Interest, Taxes, Depreciation, and Amortization).",
        "finance",
        ["ebitda", "profitability", "finance", "accounting"],
        {
            "net_income": {"type": "number", "description": "Net income"},
            "interest": {"type": "number", "description": "Interest expense"},
            "taxes": {"type": "number", "description": "Tax expense"},
            "depreciation": {"type": "number", "description": "Depreciation"},
            "amortization": {"type": "number", "description": "Amortization"}
        },
        ["net_income", "interest", "taxes", "depreciation", "amortization"],
        {"net_income": 50000, "interest": 5000, "taxes": 15000, "depreciation": 8000, "amortization": 2000},
        {"result": 80000.0}
    ),
    (
        "ebit",
        "EBIT Calculate",
        "Calculate EBIT (Earnings Before Interest and Taxes).",
        "finance",
        ["ebit", "profitability", "finance", "accounting"],
        {
            "net_income": {"type": "number", "description": "Net income"},
            "interest": {"type": "number", "description": "Interest expense"},
            "taxes": {"type": "number", "description": "Tax expense"}
        },
        ["net_income", "interest", "taxes"],
        {"net_income": 50000, "interest": 5000, "taxes": 15000},
        {"result": 70000.0}
    ),
    (
        "debt-ratio",
        "Debt Ratio Calculate",
        "Calculate the debt ratio (total liabilities / total assets).",
        "finance",
        ["debt-ratio", "leverage", "finance", "accounting"],
        {
            "total_liabilities": {"type": "number", "description": "Total liabilities"},
            "total_assets": {"type": "number", "description": "Total assets"}
        },
        ["total_liabilities", "total_assets"],
        {"total_liabilities": 50000, "total_assets": 100000},
        {"result": 0.5}
    ),
    (
        "current-ratio",
        "Current Ratio Calculate",
        "Calculate the current ratio (current assets / current liabilities).",
        "finance",
        ["current-ratio", "liquidity", "finance", "accounting"],
        {
            "current_assets": {"type": "number", "description": "Current assets"},
            "current_liabilities": {"type": "number", "description": "Current liabilities"}
        },
        ["current_assets", "current_liabilities"],
        {"current_assets": 150000, "current_liabilities": 75000},
        {"result": 2.0}
    ),
    (
        "quick-ratio",
        "Quick Ratio Calculate",
        "Calculate the quick ratio (liquid assets / current liabilities), excluding inventory.",
        "finance",
        ["quick-ratio", "acid-test", "liquidity", "finance"],
        {
            "cash": {"type": "number", "description": "Cash and cash equivalents"},
            "short_term_investments": {"type": "number", "description": "Short-term investments", "default": 0},
            "accounts_receivable": {"type": "number", "description": "Accounts receivable"},
            "current_liabilities": {"type": "number", "description": "Current liabilities"}
        },
        ["cash", "accounts_receivable", "current_liabilities"],
        {"cash": 50000, "accounts_receivable": 30000, "current_liabilities": 40000},
        {"result": 2.0}
    ),
    (
        "working-capital",
        "Working Capital Calculate",
        "Calculate working capital (current assets - current liabilities).",
        "finance",
        ["working-capital", "liquidity", "finance", "accounting"],
        {
            "current_assets": {"type": "number", "description": "Current assets"},
            "current_liabilities": {"type": "number", "description": "Current liabilities"}
        },
        ["current_assets", "current_liabilities"],
        {"current_assets": 150000, "current_liabilities": 75000},
        {"result": 75000.0}
    ),
    (
        "inventory-turnover",
        "Inventory Turnover Calculate",
        "Calculate inventory turnover ratio (COGS / average inventory).",
        "finance",
        ["inventory-turnover", "efficiency", "finance", "accounting"],
        {
            "cost_of_goods_sold": {"type": "number", "description": "Cost of goods sold"},
            "average_inventory": {"type": "number", "description": "Average inventory value"}
        },
        ["cost_of_goods_sold", "average_inventory"],
        {"cost_of_goods_sold": 500000, "average_inventory": 100000},
        {"result": 5.0, "days_in_inventory": 73.0}
    ),
    (
        "receivables-turnover",
        "Receivables Turnover Calculate",
        "Calculate accounts receivable turnover ratio.",
        "finance",
        ["receivables-turnover", "efficiency", "finance", "accounting"],
        {
            "net_credit_sales": {"type": "number", "description": "Net credit sales"},
            "average_accounts_receivable": {"type": "number", "description": "Average accounts receivable"}
        },
        ["net_credit_sales", "average_accounts_receivable"],
        {"net_credit_sales": 500000, "average_accounts_receivable": 50000},
        {"result": 10.0, "days_sales_outstanding": 36.5}
    ),
    (
        "payables-turnover",
        "Payables Turnover Calculate",
        "Calculate accounts payable turnover ratio.",
        "finance",
        ["payables-turnover", "efficiency", "finance", "accounting"],
        {
            "cost_of_goods_sold": {"type": "number", "description": "Cost of goods sold"},
            "average_accounts_payable": {"type": "number", "description": "Average accounts payable"}
        },
        ["cost_of_goods_sold", "average_accounts_payable"],
        {"cost_of_goods_sold": 500000, "average_accounts_payable": 50000},
        {"result": 10.0, "days_payable_outstanding": 36.5}
    ),
    (
        "wacc-calculate",
        "WACC Calculate",
        "Calculate the Weighted Average Cost of Capital (WACC).",
        "finance",
        ["wacc", "cost-of-capital", "finance", "valuation"],
        {
            "equity_value": {"type": "number", "description": "Market value of equity"},
            "debt_value": {"type": "number", "description": "Market value of debt"},
            "cost_of_equity": {"type": "number", "description": "Cost of equity (decimal)"},
            "cost_of_debt": {"type": "number", "description": "Cost of debt (decimal)"},
            "tax_rate": {"type": "number", "description": "Corporate tax rate (decimal)"}
        },
        ["equity_value", "debt_value", "cost_of_equity", "cost_of_debt", "tax_rate"],
        {"equity_value": 500000, "debt_value": 200000, "cost_of_equity": 0.12, "cost_of_debt": 0.06, "tax_rate": 0.25},
        {"result": 0.0994}
    ),
    (
        "dcf-valuation",
        "DCF Valuation",
        "Calculate discounted cash flow (DCF) valuation of a business or investment.",
        "finance",
        ["dcf", "discounted-cash-flow", "valuation", "finance"],
        {
            "cash_flows": {"type": "array", "items": {"type": "number"}, "description": "Projected free cash flows"},
            "discount_rate": {"type": "number", "description": "Discount rate (WACC, decimal)"},
            "terminal_growth_rate": {"type": "number", "description": "Terminal growth rate (decimal)", "default": 0.02}
        },
        ["cash_flows", "discount_rate"],
        {"cash_flows": [100000, 110000, 121000, 133100, 146410], "discount_rate": 0.1, "terminal_growth_rate": 0.02},
        {"result": 1234567.0}
    ),
    (
        "discount-factor",
        "Discount Factor Calculate",
        "Calculate the discount factor for a given rate and number of periods.",
        "finance",
        ["discount-factor", "time-value", "finance"],
        {
            "rate": {"type": "number", "description": "Discount rate per period (decimal)"},
            "periods": {"type": "number", "description": "Number of periods"}
        },
        ["rate", "periods"],
        {"rate": 0.1, "periods": 5},
        {"result": 0.6209}
    ),
    (
        "annuity-factor",
        "Annuity Factor Calculate",
        "Calculate the present value annuity factor (PVAF) for a series of equal payments.",
        "finance",
        ["annuity", "annuity-factor", "finance"],
        {
            "rate": {"type": "number", "description": "Interest rate per period (decimal)"},
            "periods": {"type": "number", "description": "Number of periods"}
        },
        ["rate", "periods"],
        {"rate": 0.05, "periods": 10},
        {"result": 7.7217}
    ),
    (
        "capitalize-cost",
        "Capitalize Cost Calculate",
        "Calculate the capitalized cost of an asset including purchase price and additional costs.",
        "finance",
        ["capitalize", "cost", "accounting", "finance"],
        {
            "purchase_price": {"type": "number", "description": "Purchase price of asset"},
            "additional_costs": {"type": "array", "items": {"type": "number"}, "description": "List of additional capitalized costs", "default": []}
        },
        ["purchase_price"],
        {"purchase_price": 50000, "additional_costs": [2000, 1500, 500]},
        {"result": 54000.0}
    ),
    (
        "amortize-cost",
        "Amortize Cost Calculate",
        "Calculate the amortization of an intangible asset cost over its useful life.",
        "finance",
        ["amortize", "amortization", "intangible", "accounting"],
        {
            "cost": {"type": "number", "description": "Total cost to amortize"},
            "useful_life": {"type": "number", "description": "Useful life in years"},
            "period": {"type": "number", "description": "Current period (1-based)", "default": 1}
        },
        ["cost", "useful_life"],
        {"cost": 120000, "useful_life": 10},
        {"result": 12000.0, "annual_amortization": 12000.0}
    ),
    (
        "tax-loss-carryforward",
        "Tax Loss Carryforward Calculate",
        "Calculate the tax benefit from carrying forward net operating losses.",
        "finance",
        ["tax", "loss-carryforward", "nol", "accounting"],
        {
            "net_operating_loss": {"type": "number", "description": "Net operating loss amount"},
            "taxable_income": {"type": "number", "description": "Current year taxable income"},
            "tax_rate": {"type": "number", "description": "Tax rate (decimal)"}
        },
        ["net_operating_loss", "taxable_income", "tax_rate"],
        {"net_operating_loss": 50000, "taxable_income": 80000, "tax_rate": 0.25},
        {"result": 12500.0, "remaining_loss": 0.0, "applied_loss": 50000.0}
    ),
    (
        "deferred-tax",
        "Deferred Tax Calculate",
        "Calculate deferred tax asset or liability from temporary differences.",
        "finance",
        ["deferred-tax", "tax", "accounting", "finance"],
        {
            "temporary_difference": {"type": "number", "description": "Temporary difference amount (positive = deferred tax liability)"},
            "tax_rate": {"type": "number", "description": "Tax rate (decimal)"}
        },
        ["temporary_difference", "tax_rate"],
        {"temporary_difference": 20000, "tax_rate": 0.25},
        {"result": 5000.0, "type": "liability"}
    ),
    (
        "goodwill-calculate",
        "Goodwill Calculate",
        "Calculate goodwill in a business acquisition (purchase price minus fair value of net assets).",
        "finance",
        ["goodwill", "acquisition", "accounting", "finance"],
        {
            "purchase_price": {"type": "number", "description": "Acquisition purchase price"},
            "fair_value_assets": {"type": "number", "description": "Fair value of acquired assets"},
            "fair_value_liabilities": {"type": "number", "description": "Fair value of acquired liabilities"}
        },
        ["purchase_price", "fair_value_assets", "fair_value_liabilities"],
        {"purchase_price": 500000, "fair_value_assets": 400000, "fair_value_liabilities": 100000},
        {"result": 200000.0, "net_assets": 300000.0}
    ),
    (
        "exchange-rate",
        "Exchange Rate Conversion",
        "Convert an amount from one currency to another using a given exchange rate.",
        "finance",
        ["exchange-rate", "currency", "forex", "finance"],
        {
            "amount": {"type": "number", "description": "Amount to convert"},
            "from_currency": {"type": "string", "description": "Source currency code (e.g. USD)"},
            "to_currency": {"type": "string", "description": "Target currency code (e.g. EUR)"},
            "rate": {"type": "number", "description": "Exchange rate (1 from_currency = rate to_currency)"}
        },
        ["amount", "from_currency", "to_currency", "rate"],
        {"amount": 1000, "from_currency": "USD", "to_currency": "EUR", "rate": 0.92},
        {"result": 920.0}
    ),
    (
        "forex-convert",
        "Forex Convert",
        "Convert currency amounts using forex rates with optional fee calculation.",
        "finance",
        ["forex", "currency-conversion", "exchange", "finance"],
        {
            "amount": {"type": "number", "description": "Amount to convert"},
            "rate": {"type": "number", "description": "Exchange rate"},
            "fee_pct": {"type": "number", "description": "Transaction fee percentage (default 0)", "default": 0}
        },
        ["amount", "rate"],
        {"amount": 1000, "rate": 1.25, "fee_pct": 0.5},
        {"result": 1243.75, "converted": 1250.0, "fee": 6.25}
    ),
    (
        "crypto-price",
        "Crypto Price Lookup",
        "Simulate cryptocurrency price lookup with mock data for common cryptocurrencies.",
        "finance",
        ["crypto", "cryptocurrency", "price", "bitcoin"],
        {
            "symbol": {"type": "string", "description": "Cryptocurrency symbol (e.g. BTC, ETH, SOL)"},
            "currency": {"type": "string", "description": "Quote currency (default USD)", "default": "USD"}
        },
        ["symbol"],
        {"symbol": "BTC"},
        {"result": 45000.0, "symbol": "BTC", "currency": "USD", "simulated": True}
    ),
    (
        "stock-price",
        "Stock Price Lookup",
        "Simulate stock price lookup with mock data for common stocks.",
        "finance",
        ["stock", "equity", "price", "market"],
        {
            "ticker": {"type": "string", "description": "Stock ticker symbol (e.g. AAPL, MSFT)"},
            "currency": {"type": "string", "description": "Quote currency (default USD)", "default": "USD"}
        },
        ["ticker"],
        {"ticker": "AAPL"},
        {"result": 175.0, "ticker": "AAPL", "simulated": True}
    ),
    (
        "portfolio-value",
        "Portfolio Value Calculate",
        "Calculate the total value of an investment portfolio.",
        "finance",
        ["portfolio", "value", "investment", "finance"],
        {
            "holdings": {"type": "array", "items": {"type": "object"}, "description": "List of holdings with 'shares' and 'price' fields"}
        },
        ["holdings"],
        {"holdings": [{"shares": 100, "price": 150.0}, {"shares": 50, "price": 200.0}]},
        {"result": 25000.0}
    ),
    (
        "portfolio-return",
        "Portfolio Return Calculate",
        "Calculate the weighted return of an investment portfolio.",
        "finance",
        ["portfolio", "return", "investment", "finance"],
        {
            "holdings": {"type": "array", "items": {"type": "object"}, "description": "List of holdings with 'weight' and 'return' fields"}
        },
        ["holdings"],
        {"holdings": [{"weight": 0.6, "return": 0.12}, {"weight": 0.4, "return": 0.08}]},
        {"result": 0.104}
    ),
    (
        "portfolio-risk",
        "Portfolio Risk Calculate",
        "Calculate portfolio risk (standard deviation) given asset weights, standard deviations, and correlation.",
        "analytics",
        ["portfolio", "risk", "standard-deviation", "finance"],
        {
            "weights": {"type": "array", "items": {"type": "number"}, "description": "Asset weights (must sum to 1)"},
            "std_devs": {"type": "array", "items": {"type": "number"}, "description": "Asset standard deviations"},
            "correlation_matrix": {"type": "array", "items": {"type": "array"}, "description": "Correlation matrix"}
        },
        ["weights", "std_devs", "correlation_matrix"],
        {"weights": [0.6, 0.4], "std_devs": [0.15, 0.10], "correlation_matrix": [[1.0, 0.3], [0.3, 1.0]]},
        {"result": 0.1136}
    ),
    (
        "sharpe-ratio",
        "Sharpe Ratio Calculate",
        "Calculate the Sharpe ratio (risk-adjusted return) of a portfolio or investment.",
        "analytics",
        ["sharpe-ratio", "risk-adjusted-return", "portfolio", "finance"],
        {
            "portfolio_return": {"type": "number", "description": "Portfolio return (decimal)"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "portfolio_std_dev": {"type": "number", "description": "Portfolio standard deviation"}
        },
        ["portfolio_return", "risk_free_rate", "portfolio_std_dev"],
        {"portfolio_return": 0.12, "risk_free_rate": 0.03, "portfolio_std_dev": 0.15},
        {"result": 0.6}
    ),
    (
        "sortino-ratio",
        "Sortino Ratio Calculate",
        "Calculate the Sortino ratio using downside deviation instead of total standard deviation.",
        "analytics",
        ["sortino-ratio", "downside-risk", "portfolio", "finance"],
        {
            "portfolio_return": {"type": "number", "description": "Portfolio return (decimal)"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "downside_deviation": {"type": "number", "description": "Downside deviation"}
        },
        ["portfolio_return", "risk_free_rate", "downside_deviation"],
        {"portfolio_return": 0.12, "risk_free_rate": 0.03, "downside_deviation": 0.08},
        {"result": 1.125}
    ),
    (
        "treynor-ratio",
        "Treynor Ratio Calculate",
        "Calculate the Treynor ratio (excess return per unit of systematic risk).",
        "analytics",
        ["treynor-ratio", "systematic-risk", "beta", "finance"],
        {
            "portfolio_return": {"type": "number", "description": "Portfolio return (decimal)"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "beta": {"type": "number", "description": "Portfolio beta"}
        },
        ["portfolio_return", "risk_free_rate", "beta"],
        {"portfolio_return": 0.12, "risk_free_rate": 0.03, "beta": 1.2},
        {"result": 0.075}
    ),
    (
        "alpha-calculate",
        "Alpha Calculate",
        "Calculate Jensen's alpha (excess return over CAPM expected return).",
        "analytics",
        ["alpha", "capm", "portfolio", "finance"],
        {
            "portfolio_return": {"type": "number", "description": "Portfolio return (decimal)"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "beta": {"type": "number", "description": "Portfolio beta"},
            "market_return": {"type": "number", "description": "Market return (decimal)"}
        },
        ["portfolio_return", "risk_free_rate", "beta", "market_return"],
        {"portfolio_return": 0.15, "risk_free_rate": 0.03, "beta": 1.2, "market_return": 0.10},
        {"result": 0.036}
    ),
    (
        "beta-calculate",
        "Beta Calculate",
        "Calculate the beta of a security relative to a benchmark.",
        "analytics",
        ["beta", "systematic-risk", "capm", "finance"],
        {
            "asset_returns": {"type": "array", "items": {"type": "number"}, "description": "List of asset returns"},
            "market_returns": {"type": "array", "items": {"type": "number"}, "description": "List of market returns"}
        },
        ["asset_returns", "market_returns"],
        {"asset_returns": [0.05, 0.10, -0.03, 0.08, 0.12], "market_returns": [0.04, 0.08, -0.02, 0.06, 0.10]},
        {"result": 1.25}
    ),
    (
        "standard-deviation-portfolio",
        "Portfolio Standard Deviation",
        "Calculate the standard deviation of a portfolio's returns.",
        "analytics",
        ["standard-deviation", "volatility", "portfolio", "statistics"],
        {
            "returns": {"type": "array", "items": {"type": "number"}, "description": "List of portfolio returns"},
            "population": {"type": "boolean", "description": "Use population std dev (default false = sample)", "default": False}
        },
        ["returns"],
        {"returns": [0.05, 0.10, -0.03, 0.08, 0.12, 0.07]},
        {"result": 0.0516}
    ),
    (
        "correlation-coefficient",
        "Correlation Coefficient Calculate",
        "Calculate the Pearson correlation coefficient between two data series.",
        "analytics",
        ["correlation", "statistics", "analytics", "finance"],
        {
            "x": {"type": "array", "items": {"type": "number"}, "description": "First data series"},
            "y": {"type": "array", "items": {"type": "number"}, "description": "Second data series"}
        },
        ["x", "y"],
        {"x": [1, 2, 3, 4, 5], "y": [2, 4, 5, 4, 5]},
        {"result": 0.9129}
    ),
    (
        "covariance",
        "Covariance Calculate",
        "Calculate the covariance between two data series.",
        "analytics",
        ["covariance", "statistics", "analytics", "finance"],
        {
            "x": {"type": "array", "items": {"type": "number"}, "description": "First data series"},
            "y": {"type": "array", "items": {"type": "number"}, "description": "Second data series"},
            "population": {"type": "boolean", "description": "Use population covariance (default false = sample)", "default": False}
        },
        ["x", "y"],
        {"x": [1, 2, 3, 4, 5], "y": [2, 4, 5, 4, 5]},
        {"result": 1.5}
    ),
    (
        "variance-portfolio",
        "Portfolio Variance Calculate",
        "Calculate the variance of a portfolio's returns.",
        "analytics",
        ["variance", "portfolio", "statistics", "finance"],
        {
            "returns": {"type": "array", "items": {"type": "number"}, "description": "List of portfolio returns"},
            "population": {"type": "boolean", "description": "Use population variance (default false = sample)", "default": False}
        },
        ["returns"],
        {"returns": [0.05, 0.10, -0.03, 0.08, 0.12]},
        {"result": 0.0034}
    ),
    (
        "var-calculate",
        "Value at Risk Calculate",
        "Calculate Value at Risk (VaR) using historical simulation or parametric method.",
        "analytics",
        ["var", "value-at-risk", "risk", "finance"],
        {
            "returns": {"type": "array", "items": {"type": "number"}, "description": "Historical returns"},
            "confidence_level": {"type": "number", "description": "Confidence level (e.g. 0.95 for 95%)", "default": 0.95},
            "portfolio_value": {"type": "number", "description": "Portfolio value (default 1)", "default": 1}
        },
        ["returns"],
        {"returns": [0.01, -0.02, 0.03, -0.05, 0.02, -0.01, 0.04, -0.03, 0.01, -0.04], "confidence_level": 0.95},
        {"result": -0.05}
    ),
    (
        "cvar-calculate",
        "Conditional Value at Risk Calculate",
        "Calculate Conditional Value at Risk (CVaR / Expected Shortfall).",
        "analytics",
        ["cvar", "expected-shortfall", "risk", "finance"],
        {
            "returns": {"type": "array", "items": {"type": "number"}, "description": "Historical returns"},
            "confidence_level": {"type": "number", "description": "Confidence level (e.g. 0.95 for 95%)", "default": 0.95},
            "portfolio_value": {"type": "number", "description": "Portfolio value (default 1)", "default": 1}
        },
        ["returns"],
        {"returns": [0.01, -0.02, 0.03, -0.05, 0.02, -0.01, 0.04, -0.03, 0.01, -0.04], "confidence_level": 0.95},
        {"result": -0.05}
    ),
    (
        "beta-distribution",
        "Beta Distribution Calculate",
        "Calculate the probability density function (PDF) and CDF of the beta distribution.",
        "analytics",
        ["beta-distribution", "statistics", "probability", "analytics"],
        {
            "x": {"type": "number", "description": "Value between 0 and 1"},
            "alpha": {"type": "number", "description": "Alpha shape parameter (> 0)"},
            "beta": {"type": "number", "description": "Beta shape parameter (> 0)"}
        },
        ["x", "alpha", "beta"],
        {"x": 0.5, "alpha": 2, "beta": 5},
        {"result": {"pdf": 1.3125, "cdf": 0.8906}}
    ),
    (
        "normal-distribution",
        "Normal Distribution Calculate",
        "Calculate the PDF and CDF of the normal (Gaussian) distribution.",
        "analytics",
        ["normal-distribution", "gaussian", "statistics", "probability"],
        {
            "x": {"type": "number", "description": "Value to evaluate"},
            "mean": {"type": "number", "description": "Mean (mu)", "default": 0},
            "std_dev": {"type": "number", "description": "Standard deviation (sigma)", "default": 1}
        },
        ["x"],
        {"x": 1.96, "mean": 0, "std_dev": 1},
        {"result": {"pdf": 0.0584, "cdf": 0.975}}
    ),
    (
        "log-normal-distribution",
        "Log-Normal Distribution Calculate",
        "Calculate the PDF and CDF of the log-normal distribution.",
        "analytics",
        ["log-normal", "statistics", "probability", "finance"],
        {
            "x": {"type": "number", "description": "Value to evaluate (> 0)"},
            "mu": {"type": "number", "description": "Mean of the underlying normal distribution", "default": 0},
            "sigma": {"type": "number", "description": "Std dev of the underlying normal distribution", "default": 1}
        },
        ["x"],
        {"x": 1.0, "mu": 0, "sigma": 1},
        {"result": {"pdf": 0.3989, "cdf": 0.5}}
    ),
    (
        "monte-carlo-sim",
        "Monte Carlo Simulation",
        "Run a Monte Carlo simulation for portfolio or investment value projection.",
        "analytics",
        ["monte-carlo", "simulation", "risk", "finance"],
        {
            "initial_value": {"type": "number", "description": "Initial portfolio value"},
            "expected_return": {"type": "number", "description": "Expected annual return (decimal)"},
            "volatility": {"type": "number", "description": "Annual volatility (decimal)"},
            "years": {"type": "number", "description": "Simulation horizon in years"},
            "simulations": {"type": "integer", "description": "Number of simulations (default 1000)", "default": 1000},
            "seed": {"type": "integer", "description": "Random seed for reproducibility (optional)"}
        },
        ["initial_value", "expected_return", "volatility", "years"],
        {"initial_value": 10000, "expected_return": 0.08, "volatility": 0.15, "years": 10, "simulations": 100, "seed": 42},
        {"result": {"mean": 21589.0, "median": 20000.0, "p5": 12000.0, "p95": 35000.0}}
    ),
    (
        "historical-volatility",
        "Historical Volatility Calculate",
        "Calculate historical volatility (annualized standard deviation of log returns).",
        "analytics",
        ["volatility", "historical-volatility", "risk", "finance"],
        {
            "prices": {"type": "array", "items": {"type": "number"}, "description": "List of historical prices"},
            "periods_per_year": {"type": "number", "description": "Trading periods per year (default 252)", "default": 252}
        },
        ["prices"],
        {"prices": [100, 102, 99, 103, 101, 105, 103, 107, 104, 108]},
        {"result": 0.1823}
    ),
    (
        "implied-volatility",
        "Implied Volatility Calculate",
        "Calculate implied volatility from option price using Black-Scholes model (Newton-Raphson method).",
        "analytics",
        ["implied-volatility", "options", "black-scholes", "finance"],
        {
            "option_price": {"type": "number", "description": "Market price of the option"},
            "spot_price": {"type": "number", "description": "Current price of underlying asset"},
            "strike_price": {"type": "number", "description": "Option strike price"},
            "time_to_expiry": {"type": "number", "description": "Time to expiry in years"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "option_type": {"type": "string", "description": "Option type: 'call' or 'put'", "default": "call"}
        },
        ["option_price", "spot_price", "strike_price", "time_to_expiry", "risk_free_rate"],
        {"option_price": 10.0, "spot_price": 100, "strike_price": 100, "time_to_expiry": 1.0, "risk_free_rate": 0.05},
        {"result": 0.2}
    ),
    (
        "black-scholes",
        "Black-Scholes Option Pricing",
        "Calculate option price using the Black-Scholes model.",
        "finance",
        ["black-scholes", "options", "derivatives", "finance"],
        {
            "spot_price": {"type": "number", "description": "Current price of underlying asset"},
            "strike_price": {"type": "number", "description": "Option strike price"},
            "time_to_expiry": {"type": "number", "description": "Time to expiry in years"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "volatility": {"type": "number", "description": "Volatility (decimal)"},
            "option_type": {"type": "string", "description": "Option type: 'call' or 'put'", "default": "call"}
        },
        ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate", "volatility"],
        {"spot_price": 100, "strike_price": 100, "time_to_expiry": 1.0, "risk_free_rate": 0.05, "volatility": 0.2},
        {"result": 10.45}
    ),
    (
        "greeks-calculate",
        "Options Greeks Calculate",
        "Calculate all option Greeks (Delta, Gamma, Theta, Vega, Rho) using Black-Scholes.",
        "finance",
        ["greeks", "options", "delta", "gamma", "theta", "vega"],
        {
            "spot_price": {"type": "number", "description": "Current price of underlying asset"},
            "strike_price": {"type": "number", "description": "Option strike price"},
            "time_to_expiry": {"type": "number", "description": "Time to expiry in years"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "volatility": {"type": "number", "description": "Volatility (decimal)"},
            "option_type": {"type": "string", "description": "Option type: 'call' or 'put'", "default": "call"}
        },
        ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate", "volatility"],
        {"spot_price": 100, "strike_price": 100, "time_to_expiry": 1.0, "risk_free_rate": 0.05, "volatility": 0.2},
        {"result": {"delta": 0.6368, "gamma": 0.0188, "theta": -6.414, "vega": 37.524, "rho": 53.232}}
    ),
    (
        "option-price",
        "Option Price Calculate",
        "Calculate option price using binomial tree model.",
        "finance",
        ["option-price", "binomial", "options", "derivatives"],
        {
            "spot_price": {"type": "number", "description": "Current price of underlying asset"},
            "strike_price": {"type": "number", "description": "Option strike price"},
            "time_to_expiry": {"type": "number", "description": "Time to expiry in years"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "volatility": {"type": "number", "description": "Volatility (decimal)"},
            "steps": {"type": "integer", "description": "Number of binomial tree steps (default 100)", "default": 100},
            "option_type": {"type": "string", "description": "Option type: 'call' or 'put'", "default": "call"}
        },
        ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate", "volatility"],
        {"spot_price": 100, "strike_price": 100, "time_to_expiry": 1.0, "risk_free_rate": 0.05, "volatility": 0.2},
        {"result": 10.45}
    ),
    (
        "put-call-parity",
        "Put-Call Parity Check",
        "Verify put-call parity and calculate the theoretical price of a put or call.",
        "finance",
        ["put-call-parity", "options", "derivatives", "finance"],
        {
            "call_price": {"type": "number", "description": "Call option price"},
            "put_price": {"type": "number", "description": "Put option price"},
            "spot_price": {"type": "number", "description": "Current price of underlying asset"},
            "strike_price": {"type": "number", "description": "Option strike price"},
            "time_to_expiry": {"type": "number", "description": "Time to expiry in years"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"}
        },
        ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate"],
        {"call_price": 10.45, "put_price": 5.57, "spot_price": 100, "strike_price": 100, "time_to_expiry": 1.0, "risk_free_rate": 0.05},
        {"result": {"parity_holds": True, "theoretical_call": 10.45, "theoretical_put": 5.57}}
    ),
    (
        "delta-hedge",
        "Delta Hedge Calculate",
        "Calculate the delta hedge ratio and number of shares needed to hedge an options position.",
        "finance",
        ["delta-hedge", "hedging", "options", "derivatives"],
        {
            "spot_price": {"type": "number", "description": "Current price of underlying asset"},
            "strike_price": {"type": "number", "description": "Option strike price"},
            "time_to_expiry": {"type": "number", "description": "Time to expiry in years"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "volatility": {"type": "number", "description": "Volatility (decimal)"},
            "num_contracts": {"type": "number", "description": "Number of option contracts (default 1)", "default": 1},
            "contract_size": {"type": "number", "description": "Shares per contract (default 100)", "default": 100},
            "option_type": {"type": "string", "description": "Option type: 'call' or 'put'", "default": "call"}
        },
        ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate", "volatility"],
        {"spot_price": 100, "strike_price": 100, "time_to_expiry": 1.0, "risk_free_rate": 0.05, "volatility": 0.2},
        {"result": {"delta": 0.6368, "shares_to_short": 63.68}}
    ),
    (
        "yield-to-maturity",
        "Yield to Maturity Calculate",
        "Calculate the yield to maturity (YTM) of a bond.",
        "finance",
        ["yield-to-maturity", "ytm", "bond", "fixed-income"],
        {
            "face_value": {"type": "number", "description": "Bond face value"},
            "coupon_rate": {"type": "number", "description": "Annual coupon rate (decimal)"},
            "years_to_maturity": {"type": "number", "description": "Years to maturity"},
            "market_price": {"type": "number", "description": "Current market price of bond"},
            "periods_per_year": {"type": "number", "description": "Coupon payments per year (default 2)", "default": 2}
        },
        ["face_value", "coupon_rate", "years_to_maturity", "market_price"],
        {"face_value": 1000, "coupon_rate": 0.06, "years_to_maturity": 10, "market_price": 950},
        {"result": 0.0653}
    ),
    (
        "bond-price",
        "Bond Price Calculate",
        "Calculate the price of a bond given yield to maturity.",
        "finance",
        ["bond-price", "bond", "fixed-income", "finance"],
        {
            "face_value": {"type": "number", "description": "Bond face value"},
            "coupon_rate": {"type": "number", "description": "Annual coupon rate (decimal)"},
            "years_to_maturity": {"type": "number", "description": "Years to maturity"},
            "yield_to_maturity": {"type": "number", "description": "Yield to maturity (decimal)"},
            "periods_per_year": {"type": "number", "description": "Coupon payments per year (default 2)", "default": 2}
        },
        ["face_value", "coupon_rate", "years_to_maturity", "yield_to_maturity"],
        {"face_value": 1000, "coupon_rate": 0.06, "years_to_maturity": 10, "yield_to_maturity": 0.065},
        {"result": 964.06}
    ),
    (
        "duration-calculate",
        "Bond Duration Calculate",
        "Calculate Macaulay duration and modified duration of a bond.",
        "finance",
        ["duration", "macaulay-duration", "bond", "fixed-income"],
        {
            "face_value": {"type": "number", "description": "Bond face value"},
            "coupon_rate": {"type": "number", "description": "Annual coupon rate (decimal)"},
            "years_to_maturity": {"type": "number", "description": "Years to maturity"},
            "yield_to_maturity": {"type": "number", "description": "Yield to maturity (decimal)"},
            "periods_per_year": {"type": "number", "description": "Coupon payments per year (default 2)", "default": 2}
        },
        ["face_value", "coupon_rate", "years_to_maturity", "yield_to_maturity"],
        {"face_value": 1000, "coupon_rate": 0.06, "years_to_maturity": 10, "yield_to_maturity": 0.065},
        {"result": {"macaulay_duration": 7.65, "modified_duration": 7.41}}
    ),
    (
        "convexity-calculate",
        "Bond Convexity Calculate",
        "Calculate the convexity of a bond to measure price sensitivity to interest rate changes.",
        "finance",
        ["convexity", "bond", "fixed-income", "interest-rate-risk"],
        {
            "face_value": {"type": "number", "description": "Bond face value"},
            "coupon_rate": {"type": "number", "description": "Annual coupon rate (decimal)"},
            "years_to_maturity": {"type": "number", "description": "Years to maturity"},
            "yield_to_maturity": {"type": "number", "description": "Yield to maturity (decimal)"},
            "periods_per_year": {"type": "number", "description": "Coupon payments per year (default 2)", "default": 2}
        },
        ["face_value", "coupon_rate", "years_to_maturity", "yield_to_maturity"],
        {"face_value": 1000, "coupon_rate": 0.06, "years_to_maturity": 10, "yield_to_maturity": 0.065},
        {"result": 68.34}
    ),
    (
        "credit-spread",
        "Credit Spread Calculate",
        "Calculate the credit spread between a corporate bond yield and a risk-free rate.",
        "finance",
        ["credit-spread", "bond", "credit-risk", "fixed-income"],
        {
            "corporate_yield": {"type": "number", "description": "Corporate bond yield (decimal)"},
            "risk_free_rate": {"type": "number", "description": "Risk-free rate (decimal)"},
            "face_value": {"type": "number", "description": "Bond face value (optional)", "default": 1000}
        },
        ["corporate_yield", "risk_free_rate"],
        {"corporate_yield": 0.065, "risk_free_rate": 0.04},
        {"result": 0.025, "result_bps": 250.0}
    ),
    (
        "default-probability",
        "Default Probability Calculate",
        "Calculate the implied default probability from credit spread using reduced-form model.",
        "finance",
        ["default-probability", "credit-risk", "bond", "finance"],
        {
            "credit_spread": {"type": "number", "description": "Credit spread (decimal)"},
            "recovery_rate": {"type": "number", "description": "Expected recovery rate (decimal, default 0.4)", "default": 0.4},
            "years": {"type": "number", "description": "Time horizon in years (default 1)", "default": 1}
        },
        ["credit_spread"],
        {"credit_spread": 0.025, "recovery_rate": 0.4, "years": 1},
        {"result": 0.0417}
    ),
    (
        "recovery-rate",
        "Recovery Rate Calculate",
        "Calculate the expected recovery rate and loss given default (LGD) for a bond.",
        "finance",
        ["recovery-rate", "lgd", "credit-risk", "bond"],
        {
            "face_value": {"type": "number", "description": "Bond face value"},
            "recovery_amount": {"type": "number", "description": "Expected recovery amount in default"},
            "accrued_interest": {"type": "number", "description": "Accrued interest (default 0)", "default": 0}
        },
        ["face_value", "recovery_amount"],
        {"face_value": 1000, "recovery_amount": 400},
        {"result": 0.4, "lgd": 0.6, "loss_amount": 600.0}
    ),
]


JSONC_TEMPLATE = '''{
  "author": "functionfly",
  "name": "{name}",
  "version": "1.0.0",
  "runtime": "python3.12",
  "title": "{title}",
  "description": "{description}",
  "category": "{category}",
  "tags": {tags},
  "input": {input_schema},
  "output": {
    "type": "object",
    "properties": {
      "ok": { "type": "boolean" },
      "result": {},
      "error": { "type": "string" }
    }
  },
  "deterministic": true,
  "idempotent": true,
  "cache_ttl": 3600,
  "timeout_ms": 1000,
  "memory_mb": 128,
  "example": {
    "input": {example_input},
    "output": {example_output}
  }
}
'''


def make_jsonc(func_def):
    name, title, description, category, tags, input_props, required, example_input, example_output_extra = func_def
    
    input_schema = {
        "type": "object",
        "properties": input_props,
        "required": required
    }
    
    example_output = {"ok": True}
    example_output.update(example_output_extra)
    
    content = {
        "author": "functionfly",
        "name": name,
        "version": "1.0.0",
        "runtime": "python3.12",
        "title": title,
        "description": description,
        "category": category,
        "tags": tags,
        "input": input_schema,
        "output": {
            "type": "object",
            "properties": {
                "ok": {"type": "boolean"},
                "result": {},
                "error": {"type": "string"}
            }
        },
        "deterministic": True,
        "idempotent": True,
        "cache_ttl": 3600,
        "timeout_ms": 1000,
        "memory_mb": 128,
        "example": {
            "input": example_input,
            "output": example_output
        }
    }
    
    return json.dumps(content, indent=2)


# Python implementations for each function
IMPLEMENTATIONS = {
    "pv-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    periods = event.get("periods")
    if rate is None or periods is None:
        return {"ok": False, "error": "rate and periods are required"}
    try:
        rate = float(rate)
        periods = float(periods)
        fv = float(event.get("future_value", 0))
        pmt = float(event.get("payment", 0))
        if rate == 0:
            pv = fv + pmt * periods
        else:
            pv = fv / (1 + rate) ** periods + pmt * (1 - (1 + rate) ** (-periods)) / rate
        return {"ok": True, "result": round(pv, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "fv-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pv = event.get("present_value")
    rate = event.get("rate")
    periods = event.get("periods")
    if pv is None or rate is None or periods is None:
        return {"ok": False, "error": "present_value, rate, and periods are required"}
    try:
        pv = float(pv)
        rate = float(rate)
        periods = float(periods)
        pmt = float(event.get("payment", 0))
        if rate == 0:
            fv = pv + pmt * periods
        else:
            fv = pv * (1 + rate) ** periods + pmt * ((1 + rate) ** periods - 1) / rate
        return {"ok": True, "result": round(fv, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "npv-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    cash_flows = event.get("cash_flows")
    if rate is None or cash_flows is None:
        return {"ok": False, "error": "rate and cash_flows are required"}
    try:
        rate = float(rate)
        cash_flows = [float(cf) for cf in cash_flows]
        npv = sum(cf / (1 + rate) ** t for t, cf in enumerate(cash_flows))
        return {"ok": True, "result": round(npv, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "irr-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cash_flows = event.get("cash_flows")
    if cash_flows is None:
        return {"ok": False, "error": "cash_flows is required"}
    try:
        cash_flows = [float(cf) for cf in cash_flows]
        guess = float(event.get("guess", 0.1))
        # Newton-Raphson method
        rate = guess
        for _ in range(1000):
            npv = sum(cf / (1 + rate) ** t for t, cf in enumerate(cash_flows))
            dnpv = sum(-t * cf / (1 + rate) ** (t + 1) for t, cf in enumerate(cash_flows))
            if abs(dnpv) < 1e-12:
                break
            new_rate = rate - npv / dnpv
            if abs(new_rate - rate) < 1e-10:
                rate = new_rate
                break
            rate = new_rate
        npv_check = sum(cf / (1 + rate) ** t for t, cf in enumerate(cash_flows))
        if abs(npv_check) > 0.01:
            return {"ok": False, "error": "IRR did not converge"}
        return {"ok": True, "result": round(rate, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "pmt-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    periods = event.get("periods")
    pv = event.get("present_value")
    if rate is None or periods is None or pv is None:
        return {"ok": False, "error": "rate, periods, and present_value are required"}
    try:
        rate = float(rate)
        periods = float(periods)
        pv = float(pv)
        fv = float(event.get("future_value", 0))
        if rate == 0:
            pmt = -(pv + fv) / periods
        else:
            pmt = -rate * (pv * (1 + rate) ** periods + fv) / ((1 + rate) ** periods - 1)
        return {"ok": True, "result": round(pmt, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "nper-calculate": '''import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    payment = event.get("payment")
    pv = event.get("present_value")
    if rate is None or payment is None or pv is None:
        return {"ok": False, "error": "rate, payment, and present_value are required"}
    try:
        rate = float(rate)
        payment = float(payment)
        pv = float(pv)
        fv = float(event.get("future_value", 0))
        if rate == 0:
            if payment == 0:
                return {"ok": False, "error": "payment cannot be zero when rate is zero"}
            nper = -(pv + fv) / payment
        else:
            numerator = -fv * rate + payment
            denominator = pv * rate + payment
            if numerator <= 0 or denominator <= 0:
                return {"ok": False, "error": "Cannot compute nper with given inputs"}
            nper = math.log(numerator / denominator) / math.log(1 + rate)
        return {"ok": True, "result": round(nper, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "rate-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    periods = event.get("periods")
    payment = event.get("payment")
    pv = event.get("present_value")
    if periods is None or payment is None or pv is None:
        return {"ok": False, "error": "periods, payment, and present_value are required"}
    try:
        periods = float(periods)
        payment = float(payment)
        pv = float(pv)
        fv = float(event.get("future_value", 0))
        guess = float(event.get("guess", 0.01))
        rate = guess
        for _ in range(1000):
            if rate == 0:
                f = pv + payment * periods + fv
                df = 0
            else:
                f = (pv * (1 + rate) ** periods
                     + payment * ((1 + rate) ** periods - 1) / rate
                     + fv)
                df = (periods * pv * (1 + rate) ** (periods - 1)
                      + payment * (periods * (1 + rate) ** (periods - 1) * rate
                                   - ((1 + rate) ** periods - 1)) / rate ** 2)
            if abs(df) < 1e-12:
                break
            new_rate = rate - f / df
            if abs(new_rate - rate) < 1e-10:
                rate = new_rate
                break
            rate = new_rate
        return {"ok": True, "result": round(rate, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "loan-payment": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    principal = event.get("principal")
    annual_rate = event.get("annual_rate")
    years = event.get("years")
    if principal is None or annual_rate is None or years is None:
        return {"ok": False, "error": "principal, annual_rate, and years are required"}
    try:
        principal = float(principal)
        annual_rate = float(annual_rate)
        years = float(years)
        n = years * 12
        r = annual_rate / 12
        if r == 0:
            monthly = principal / n
        else:
            monthly = principal * r * (1 + r) ** n / ((1 + r) ** n - 1)
        total = monthly * n
        total_interest = total - principal
        return {
            "ok": True,
            "result": round(monthly, 2),
            "monthly_payment": round(monthly, 2),
            "total_payment": round(total, 2),
            "total_interest": round(total_interest, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "amortization-schedule": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    principal = event.get("principal")
    annual_rate = event.get("annual_rate")
    years = event.get("years")
    if principal is None or annual_rate is None or years is None:
        return {"ok": False, "error": "principal, annual_rate, and years are required"}
    try:
        principal = float(principal)
        annual_rate = float(annual_rate)
        years = float(years)
        periods_per_year = int(event.get("periods_per_year", 12))
        n = int(years * periods_per_year)
        r = annual_rate / periods_per_year
        if r == 0:
            payment = principal / n
        else:
            payment = principal * r * (1 + r) ** n / ((1 + r) ** n - 1)
        schedule = []
        balance = principal
        for period in range(1, n + 1):
            interest = balance * r
            principal_paid = payment - interest
            balance -= principal_paid
            if abs(balance) < 0.01:
                balance = 0
            schedule.append({
                "period": period,
                "payment": round(payment, 2),
                "principal": round(principal_paid, 2),
                "interest": round(interest, 2),
                "balance": round(max(balance, 0), 2)
            })
        return {"ok": True, "result": schedule, "payment": round(payment, 2), "total_periods": n}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "mortgage-payment": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    home_price = event.get("home_price")
    down_payment = event.get("down_payment")
    annual_rate = event.get("annual_rate")
    years = event.get("years")
    if home_price is None or down_payment is None or annual_rate is None or years is None:
        return {"ok": False, "error": "home_price, down_payment, annual_rate, and years are required"}
    try:
        home_price = float(home_price)
        down_payment = float(down_payment)
        annual_rate = float(annual_rate)
        years = float(years)
        loan_amount = home_price - down_payment
        if loan_amount <= 0:
            return {"ok": False, "error": "down_payment must be less than home_price"}
        n = years * 12
        r = annual_rate / 12
        if r == 0:
            monthly = loan_amount / n
        else:
            monthly = loan_amount * r * (1 + r) ** n / ((1 + r) ** n - 1)
        total = monthly * n
        return {
            "ok": True,
            "result": round(monthly, 2),
            "monthly_payment": round(monthly, 2),
            "loan_amount": round(loan_amount, 2),
            "total_payment": round(total, 2),
            "total_interest": round(total - loan_amount, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "investment-return": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    initial = event.get("initial_value")
    final = event.get("final_value")
    if initial is None or final is None:
        return {"ok": False, "error": "initial_value and final_value are required"}
    try:
        initial = float(initial)
        final = float(final)
        years = float(event.get("years", 1))
        if initial == 0:
            return {"ok": False, "error": "initial_value cannot be zero"}
        total_return = (final - initial) / initial
        if years > 0:
            annualized = (final / initial) ** (1 / years) - 1
        else:
            annualized = None
        result = {
            "ok": True,
            "result": round(total_return, 6),
            "total_return": round(total_return, 6),
            "total_return_pct": round(total_return * 100, 4)
        }
        if annualized is not None:
            result["annualized_return"] = round(annualized, 6)
        return result
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "roi-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    gain = event.get("gain")
    cost = event.get("cost")
    if gain is None or cost is None:
        return {"ok": False, "error": "gain and cost are required"}
    try:
        gain = float(gain)
        cost = float(cost)
        if cost == 0:
            return {"ok": False, "error": "cost cannot be zero"}
        roi = gain / cost
        return {"ok": True, "result": round(roi, 6), "result_pct": round(roi * 100, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "cagr-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    bv = event.get("beginning_value")
    ev = event.get("ending_value")
    years = event.get("years")
    if bv is None or ev is None or years is None:
        return {"ok": False, "error": "beginning_value, ending_value, and years are required"}
    try:
        bv = float(bv)
        ev = float(ev)
        years = float(years)
        if bv <= 0:
            return {"ok": False, "error": "beginning_value must be positive"}
        if years <= 0:
            return {"ok": False, "error": "years must be positive"}
        cagr = (ev / bv) ** (1 / years) - 1
        return {"ok": True, "result": round(cagr, 6), "result_pct": round(cagr * 100, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "compound-growth": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    initial = event.get("initial_value")
    rate = event.get("growth_rate")
    periods = event.get("periods")
    if initial is None or rate is None or periods is None:
        return {"ok": False, "error": "initial_value, growth_rate, and periods are required"}
    try:
        initial = float(initial)
        rate = float(rate)
        periods = float(periods)
        result = initial * (1 + rate) ** periods
        return {"ok": True, "result": round(result, 6), "growth": round(result - initial, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "depreciation-straight-line": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cost = event.get("cost")
    salvage = event.get("salvage_value")
    life = event.get("useful_life")
    if cost is None or salvage is None or life is None:
        return {"ok": False, "error": "cost, salvage_value, and useful_life are required"}
    try:
        cost = float(cost)
        salvage = float(salvage)
        life = float(life)
        if life <= 0:
            return {"ok": False, "error": "useful_life must be positive"}
        annual_dep = (cost - salvage) / life
        schedule = [{"year": y + 1, "depreciation": round(annual_dep, 2), "book_value": round(cost - annual_dep * (y + 1), 2)} for y in range(int(life))]
        return {"ok": True, "result": round(annual_dep, 6), "annual_depreciation": round(annual_dep, 2), "schedule": schedule}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "depreciation-ddb": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cost = event.get("cost")
    salvage = event.get("salvage_value")
    life = event.get("useful_life")
    period = event.get("period")
    if cost is None or salvage is None or life is None or period is None:
        return {"ok": False, "error": "cost, salvage_value, useful_life, and period are required"}
    try:
        cost = float(cost)
        salvage = float(salvage)
        life = float(life)
        period = int(period)
        if life <= 0:
            return {"ok": False, "error": "useful_life must be positive"}
        if period < 1 or period > life:
            return {"ok": False, "error": f"period must be between 1 and {int(life)}"}
        rate = 2 / life
        book_value = cost
        dep = 0
        for p in range(1, period + 1):
            dep = min(rate * book_value, book_value - salvage)
            dep = max(dep, 0)
            book_value -= dep
        return {"ok": True, "result": round(dep, 6), "book_value_after": round(book_value, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "depreciation-sum-of-years": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cost = event.get("cost")
    salvage = event.get("salvage_value")
    life = event.get("useful_life")
    period = event.get("period")
    if cost is None or salvage is None or life is None or period is None:
        return {"ok": False, "error": "cost, salvage_value, useful_life, and period are required"}
    try:
        cost = float(cost)
        salvage = float(salvage)
        life = int(life)
        period = int(period)
        if life <= 0:
            return {"ok": False, "error": "useful_life must be positive"}
        if period < 1 or period > life:
            return {"ok": False, "error": f"period must be between 1 and {life}"}
        sum_of_years = life * (life + 1) / 2
        fraction = (life - period + 1) / sum_of_years
        dep = (cost - salvage) * fraction
        return {"ok": True, "result": round(dep, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "break-even": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    fixed = event.get("fixed_costs")
    var_cost = event.get("variable_cost_per_unit")
    price = event.get("price_per_unit")
    if fixed is None or var_cost is None or price is None:
        return {"ok": False, "error": "fixed_costs, variable_cost_per_unit, and price_per_unit are required"}
    try:
        fixed = float(fixed)
        var_cost = float(var_cost)
        price = float(price)
        contribution_margin = price - var_cost
        if contribution_margin <= 0:
            return {"ok": False, "error": "price_per_unit must be greater than variable_cost_per_unit"}
        units = fixed / contribution_margin
        revenue = units * price
        return {
            "ok": True,
            "result": round(units, 4),
            "break_even_units": round(units, 4),
            "break_even_revenue": round(revenue, 2),
            "contribution_margin": round(contribution_margin, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "profit-margin": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    net_profit = event.get("net_profit")
    revenue = event.get("revenue")
    if net_profit is None or revenue is None:
        return {"ok": False, "error": "net_profit and revenue are required"}
    try:
        net_profit = float(net_profit)
        revenue = float(revenue)
        if revenue == 0:
            return {"ok": False, "error": "revenue cannot be zero"}
        margin = net_profit / revenue
        return {"ok": True, "result": round(margin, 6), "result_pct": round(margin * 100, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "gross-margin": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    revenue = event.get("revenue")
    cogs = event.get("cost_of_goods_sold")
    if revenue is None or cogs is None:
        return {"ok": False, "error": "revenue and cost_of_goods_sold are required"}
    try:
        revenue = float(revenue)
        cogs = float(cogs)
        if revenue == 0:
            return {"ok": False, "error": "revenue cannot be zero"}
        gross_profit = revenue - cogs
        margin = gross_profit / revenue
        return {"ok": True, "result": round(margin, 6), "result_pct": round(margin * 100, 4), "gross_profit": round(gross_profit, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "operating-margin": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    operating_income = event.get("operating_income")
    revenue = event.get("revenue")
    if operating_income is None or revenue is None:
        return {"ok": False, "error": "operating_income and revenue are required"}
    try:
        operating_income = float(operating_income)
        revenue = float(revenue)
        if revenue == 0:
            return {"ok": False, "error": "revenue cannot be zero"}
        margin = operating_income / revenue
        return {"ok": True, "result": round(margin, 6), "result_pct": round(margin * 100, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "ebitda": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["net_income", "interest", "taxes", "depreciation", "amortization"]
    for field in required:
        if event.get(field) is None:
            return {"ok": False, "error": f"{field} is required"}
    try:
        net_income = float(event["net_income"])
        interest = float(event["interest"])
        taxes = float(event["taxes"])
        depreciation = float(event["depreciation"])
        amortization = float(event["amortization"])
        ebitda = net_income + interest + taxes + depreciation + amortization
        return {"ok": True, "result": round(ebitda, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "ebit": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    net_income = event.get("net_income")
    interest = event.get("interest")
    taxes = event.get("taxes")
    if net_income is None or interest is None or taxes is None:
        return {"ok": False, "error": "net_income, interest, and taxes are required"}
    try:
        ebit = float(net_income) + float(interest) + float(taxes)
        return {"ok": True, "result": round(ebit, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "debt-ratio": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    liabilities = event.get("total_liabilities")
    assets = event.get("total_assets")
    if liabilities is None or assets is None:
        return {"ok": False, "error": "total_liabilities and total_assets are required"}
    try:
        liabilities = float(liabilities)
        assets = float(assets)
        if assets == 0:
            return {"ok": False, "error": "total_assets cannot be zero"}
        ratio = liabilities / assets
        return {"ok": True, "result": round(ratio, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "current-ratio": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    current_assets = event.get("current_assets")
    current_liabilities = event.get("current_liabilities")
    if current_assets is None or current_liabilities is None:
        return {"ok": False, "error": "current_assets and current_liabilities are required"}
    try:
        current_assets = float(current_assets)
        current_liabilities = float(current_liabilities)
        if current_liabilities == 0:
            return {"ok": False, "error": "current_liabilities cannot be zero"}
        ratio = current_assets / current_liabilities
        return {"ok": True, "result": round(ratio, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "quick-ratio": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cash = event.get("cash")
    ar = event.get("accounts_receivable")
    cl = event.get("current_liabilities")
    if cash is None or ar is None or cl is None:
        return {"ok": False, "error": "cash, accounts_receivable, and current_liabilities are required"}
    try:
        cash = float(cash)
        ar = float(ar)
        cl = float(cl)
        sti = float(event.get("short_term_investments", 0))
        if cl == 0:
            return {"ok": False, "error": "current_liabilities cannot be zero"}
        ratio = (cash + sti + ar) / cl
        return {"ok": True, "result": round(ratio, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "working-capital": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    current_assets = event.get("current_assets")
    current_liabilities = event.get("current_liabilities")
    if current_assets is None or current_liabilities is None:
        return {"ok": False, "error": "current_assets and current_liabilities are required"}
    try:
        wc = float(current_assets) - float(current_liabilities)
        return {"ok": True, "result": round(wc, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "inventory-turnover": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cogs = event.get("cost_of_goods_sold")
    avg_inv = event.get("average_inventory")
    if cogs is None or avg_inv is None:
        return {"ok": False, "error": "cost_of_goods_sold and average_inventory are required"}
    try:
        cogs = float(cogs)
        avg_inv = float(avg_inv)
        if avg_inv == 0:
            return {"ok": False, "error": "average_inventory cannot be zero"}
        turnover = cogs / avg_inv
        days = 365 / turnover
        return {"ok": True, "result": round(turnover, 6), "days_in_inventory": round(days, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "receivables-turnover": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    sales = event.get("net_credit_sales")
    avg_ar = event.get("average_accounts_receivable")
    if sales is None or avg_ar is None:
        return {"ok": False, "error": "net_credit_sales and average_accounts_receivable are required"}
    try:
        sales = float(sales)
        avg_ar = float(avg_ar)
        if avg_ar == 0:
            return {"ok": False, "error": "average_accounts_receivable cannot be zero"}
        turnover = sales / avg_ar
        dso = 365 / turnover
        return {"ok": True, "result": round(turnover, 6), "days_sales_outstanding": round(dso, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "payables-turnover": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cogs = event.get("cost_of_goods_sold")
    avg_ap = event.get("average_accounts_payable")
    if cogs is None or avg_ap is None:
        return {"ok": False, "error": "cost_of_goods_sold and average_accounts_payable are required"}
    try:
        cogs = float(cogs)
        avg_ap = float(avg_ap)
        if avg_ap == 0:
            return {"ok": False, "error": "average_accounts_payable cannot be zero"}
        turnover = cogs / avg_ap
        dpo = 365 / turnover
        return {"ok": True, "result": round(turnover, 6), "days_payable_outstanding": round(dpo, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "wacc-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["equity_value", "debt_value", "cost_of_equity", "cost_of_debt", "tax_rate"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        E = float(event["equity_value"])
        D = float(event["debt_value"])
        Re = float(event["cost_of_equity"])
        Rd = float(event["cost_of_debt"])
        T = float(event["tax_rate"])
        V = E + D
        if V == 0:
            return {"ok": False, "error": "Total value (equity + debt) cannot be zero"}
        wacc = (E / V) * Re + (D / V) * Rd * (1 - T)
        return {"ok": True, "result": round(wacc, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "dcf-valuation": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cash_flows = event.get("cash_flows")
    discount_rate = event.get("discount_rate")
    if cash_flows is None or discount_rate is None:
        return {"ok": False, "error": "cash_flows and discount_rate are required"}
    try:
        cash_flows = [float(cf) for cf in cash_flows]
        r = float(discount_rate)
        g = float(event.get("terminal_growth_rate", 0.02))
        if r <= g:
            return {"ok": False, "error": "discount_rate must be greater than terminal_growth_rate"}
        # PV of explicit cash flows
        pv_cfs = sum(cf / (1 + r) ** (t + 1) for t, cf in enumerate(cash_flows))
        # Terminal value (Gordon Growth Model)
        last_cf = cash_flows[-1]
        terminal_value = last_cf * (1 + g) / (r - g)
        pv_terminal = terminal_value / (1 + r) ** len(cash_flows)
        total = pv_cfs + pv_terminal
        return {
            "ok": True,
            "result": round(total, 2),
            "pv_cash_flows": round(pv_cfs, 2),
            "pv_terminal_value": round(pv_terminal, 2),
            "terminal_value": round(terminal_value, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "discount-factor": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    periods = event.get("periods")
    if rate is None or periods is None:
        return {"ok": False, "error": "rate and periods are required"}
    try:
        rate = float(rate)
        periods = float(periods)
        factor = 1 / (1 + rate) ** periods
        return {"ok": True, "result": round(factor, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "annuity-factor": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    periods = event.get("periods")
    if rate is None or periods is None:
        return {"ok": False, "error": "rate and periods are required"}
    try:
        rate = float(rate)
        periods = float(periods)
        if rate == 0:
            factor = periods
        else:
            factor = (1 - (1 + rate) ** (-periods)) / rate
        return {"ok": True, "result": round(factor, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "capitalize-cost": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    purchase_price = event.get("purchase_price")
    if purchase_price is None:
        return {"ok": False, "error": "purchase_price is required"}
    try:
        total = float(purchase_price)
        additional = event.get("additional_costs", [])
        for cost in additional:
            total += float(cost)
        return {"ok": True, "result": round(total, 2), "capitalized_cost": round(total, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "amortize-cost": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cost = event.get("cost")
    useful_life = event.get("useful_life")
    if cost is None or useful_life is None:
        return {"ok": False, "error": "cost and useful_life are required"}
    try:
        cost = float(cost)
        useful_life = float(useful_life)
        period = int(event.get("period", 1))
        if useful_life <= 0:
            return {"ok": False, "error": "useful_life must be positive"}
        annual = cost / useful_life
        accumulated = annual * period
        book_value = max(cost - accumulated, 0)
        return {
            "ok": True,
            "result": round(annual, 2),
            "annual_amortization": round(annual, 2),
            "accumulated_amortization": round(accumulated, 2),
            "book_value": round(book_value, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "tax-loss-carryforward": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    nol = event.get("net_operating_loss")
    taxable_income = event.get("taxable_income")
    tax_rate = event.get("tax_rate")
    if nol is None or taxable_income is None or tax_rate is None:
        return {"ok": False, "error": "net_operating_loss, taxable_income, and tax_rate are required"}
    try:
        nol = float(nol)
        taxable_income = float(taxable_income)
        tax_rate = float(tax_rate)
        applied = min(nol, taxable_income)
        remaining = max(nol - taxable_income, 0)
        tax_benefit = applied * tax_rate
        return {
            "ok": True,
            "result": round(tax_benefit, 2),
            "applied_loss": round(applied, 2),
            "remaining_loss": round(remaining, 2),
            "tax_benefit": round(tax_benefit, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "deferred-tax": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    diff = event.get("temporary_difference")
    tax_rate = event.get("tax_rate")
    if diff is None or tax_rate is None:
        return {"ok": False, "error": "temporary_difference and tax_rate are required"}
    try:
        diff = float(diff)
        tax_rate = float(tax_rate)
        deferred = diff * tax_rate
        dtype = "liability" if diff > 0 else "asset"
        return {"ok": True, "result": round(deferred, 2), "type": dtype}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "goodwill-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    purchase_price = event.get("purchase_price")
    fv_assets = event.get("fair_value_assets")
    fv_liabilities = event.get("fair_value_liabilities")
    if purchase_price is None or fv_assets is None or fv_liabilities is None:
        return {"ok": False, "error": "purchase_price, fair_value_assets, and fair_value_liabilities are required"}
    try:
        purchase_price = float(purchase_price)
        fv_assets = float(fv_assets)
        fv_liabilities = float(fv_liabilities)
        net_assets = fv_assets - fv_liabilities
        goodwill = purchase_price - net_assets
        return {"ok": True, "result": round(goodwill, 2), "net_assets": round(net_assets, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "exchange-rate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    amount = event.get("amount")
    from_currency = event.get("from_currency")
    to_currency = event.get("to_currency")
    rate = event.get("rate")
    if amount is None or from_currency is None or to_currency is None or rate is None:
        return {"ok": False, "error": "amount, from_currency, to_currency, and rate are required"}
    try:
        amount = float(amount)
        rate = float(rate)
        converted = amount * rate
        return {
            "ok": True,
            "result": round(converted, 6),
            "from_currency": str(from_currency).upper(),
            "to_currency": str(to_currency).upper(),
            "rate": rate,
            "original_amount": amount
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "forex-convert": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    amount = event.get("amount")
    rate = event.get("rate")
    if amount is None or rate is None:
        return {"ok": False, "error": "amount and rate are required"}
    try:
        amount = float(amount)
        rate = float(rate)
        fee_pct = float(event.get("fee_pct", 0))
        converted = amount * rate
        fee = converted * fee_pct / 100
        net = converted - fee
        return {"ok": True, "result": round(net, 6), "converted": round(converted, 6), "fee": round(fee, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "crypto-price": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    symbol = event.get("symbol")
    if symbol is None:
        return {"ok": False, "error": "symbol is required"}
    # Simulated prices (mock data)
    MOCK_PRICES = {
        "BTC": 45000.0, "ETH": 2500.0, "SOL": 100.0, "BNB": 300.0,
        "ADA": 0.45, "XRP": 0.55, "DOT": 7.0, "DOGE": 0.08,
        "AVAX": 35.0, "MATIC": 0.85, "LINK": 15.0, "UNI": 6.0,
        "LTC": 70.0, "ATOM": 10.0, "ALGO": 0.15
    }
    symbol = str(symbol).upper()
    currency = str(event.get("currency", "USD")).upper()
    price = MOCK_PRICES.get(symbol)
    if price is None:
        return {"ok": False, "error": f"Symbol {symbol} not found in mock data"}
    return {"ok": True, "result": price, "symbol": symbol, "currency": currency, "simulated": True}
''',

    "stock-price": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    ticker = event.get("ticker")
    if ticker is None:
        return {"ok": False, "error": "ticker is required"}
    # Simulated prices (mock data)
    MOCK_PRICES = {
        "AAPL": 175.0, "MSFT": 380.0, "GOOGL": 140.0, "AMZN": 185.0,
        "NVDA": 500.0, "META": 480.0, "TSLA": 200.0, "BRK.B": 360.0,
        "JPM": 195.0, "V": 270.0, "JNJ": 155.0, "WMT": 60.0,
        "PG": 155.0, "MA": 460.0, "HD": 350.0, "BAC": 35.0,
        "XOM": 105.0, "CVX": 155.0, "ABBV": 165.0, "KO": 60.0
    }
    ticker = str(ticker).upper()
    currency = str(event.get("currency", "USD")).upper()
    price = MOCK_PRICES.get(ticker)
    if price is None:
        return {"ok": False, "error": f"Ticker {ticker} not found in mock data"}
    return {"ok": True, "result": price, "ticker": ticker, "currency": currency, "simulated": True}
''',

    "portfolio-value": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    holdings = event.get("holdings")
    if holdings is None:
        return {"ok": False, "error": "holdings is required"}
    try:
        total = 0.0
        details = []
        for h in holdings:
            shares = float(h.get("shares", 0))
            price = float(h.get("price", 0))
            value = shares * price
            total += value
            details.append({"shares": shares, "price": price, "value": round(value, 2)})
        return {"ok": True, "result": round(total, 2), "holdings": details}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "portfolio-return": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    holdings = event.get("holdings")
    if holdings is None:
        return {"ok": False, "error": "holdings is required"}
    try:
        total_weight = 0.0
        weighted_return = 0.0
        for h in holdings:
            w = float(h.get("weight", 0))
            r = float(h.get("return", 0))
            weighted_return += w * r
            total_weight += w
        if abs(total_weight - 1.0) > 0.01:
            return {"ok": False, "error": f"Weights must sum to 1.0, got {total_weight}"}
        return {"ok": True, "result": round(weighted_return, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "portfolio-risk": '''import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    weights = event.get("weights")
    std_devs = event.get("std_devs")
    corr_matrix = event.get("correlation_matrix")
    if weights is None or std_devs is None or corr_matrix is None:
        return {"ok": False, "error": "weights, std_devs, and correlation_matrix are required"}
    try:
        w = [float(x) for x in weights]
        s = [float(x) for x in std_devs]
        n = len(w)
        if len(s) != n or len(corr_matrix) != n:
            return {"ok": False, "error": "weights, std_devs, and correlation_matrix must have same length"}
        variance = 0.0
        for i in range(n):
            for j in range(n):
                variance += w[i] * w[j] * s[i] * s[j] * float(corr_matrix[i][j])
        std_dev = math.sqrt(variance)
        return {"ok": True, "result": round(std_dev, 8), "variance": round(variance, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "sharpe-ratio": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pr = event.get("portfolio_return")
    rfr = event.get("risk_free_rate")
    std = event.get("portfolio_std_dev")
    if pr is None or rfr is None or std is None:
        return {"ok": False, "error": "portfolio_return, risk_free_rate, and portfolio_std_dev are required"}
    try:
        pr = float(pr)
        rfr = float(rfr)
        std = float(std)
        if std == 0:
            return {"ok": False, "error": "portfolio_std_dev cannot be zero"}
        sharpe = (pr - rfr) / std
        return {"ok": True, "result": round(sharpe, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "sortino-ratio": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pr = event.get("portfolio_return")
    rfr = event.get("risk_free_rate")
    dd = event.get("downside_deviation")
    if pr is None or rfr is None or dd is None:
        return {"ok": False, "error": "portfolio_return, risk_free_rate, and downside_deviation are required"}
    try:
        pr = float(pr)
        rfr = float(rfr)
        dd = float(dd)
        if dd == 0:
            return {"ok": False, "error": "downside_deviation cannot be zero"}
        sortino = (pr - rfr) / dd
        return {"ok": True, "result": round(sortino, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "treynor-ratio": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pr = event.get("portfolio_return")
    rfr = event.get("risk_free_rate")
    beta = event.get("beta")
    if pr is None or rfr is None or beta is None:
        return {"ok": False, "error": "portfolio_return, risk_free_rate, and beta are required"}
    try:
        pr = float(pr)
        rfr = float(rfr)
        beta = float(beta)
        if beta == 0:
            return {"ok": False, "error": "beta cannot be zero"}
        treynor = (pr - rfr) / beta
        return {"ok": True, "result": round(treynor, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "alpha-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    pr = event.get("portfolio_return")
    rfr = event.get("risk_free_rate")
    beta = event.get("beta")
    mr = event.get("market_return")
    if pr is None or rfr is None or beta is None or mr is None:
        return {"ok": False, "error": "portfolio_return, risk_free_rate, beta, and market_return are required"}
    try:
        pr = float(pr)
        rfr = float(rfr)
        beta = float(beta)
        mr = float(mr)
        expected_return = rfr + beta * (mr - rfr)
        alpha = pr - expected_return
        return {"ok": True, "result": round(alpha, 6), "expected_return": round(expected_return, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "beta-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    asset_returns = event.get("asset_returns")
    market_returns = event.get("market_returns")
    if asset_returns is None or market_returns is None:
        return {"ok": False, "error": "asset_returns and market_returns are required"}
    try:
        a = [float(x) for x in asset_returns]
        m = [float(x) for x in market_returns]
        if len(a) != len(m):
            return {"ok": False, "error": "asset_returns and market_returns must have same length"}
        if len(a) < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        n = len(a)
        mean_a = sum(a) / n
        mean_m = sum(m) / n
        cov = sum((a[i] - mean_a) * (m[i] - mean_m) for i in range(n)) / (n - 1)
        var_m = sum((m[i] - mean_m) ** 2 for i in range(n)) / (n - 1)
        if var_m == 0:
            return {"ok": False, "error": "Market returns have zero variance"}
        beta = cov / var_m
        return {"ok": True, "result": round(beta, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "standard-deviation-portfolio": '''import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    returns = event.get("returns")
    if returns is None:
        return {"ok": False, "error": "returns is required"}
    try:
        r = [float(x) for x in returns]
        n = len(r)
        if n < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        population = bool(event.get("population", False))
        mean = sum(r) / n
        divisor = n if population else n - 1
        variance = sum((x - mean) ** 2 for x in r) / divisor
        std_dev = math.sqrt(variance)
        return {"ok": True, "result": round(std_dev, 8), "mean": round(mean, 8), "variance": round(variance, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "correlation-coefficient": '''import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    x = event.get("x")
    y = event.get("y")
    if x is None or y is None:
        return {"ok": False, "error": "x and y are required"}
    try:
        x = [float(v) for v in x]
        y = [float(v) for v in y]
        n = len(x)
        if n != len(y):
            return {"ok": False, "error": "x and y must have same length"}
        if n < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        mean_x = sum(x) / n
        mean_y = sum(y) / n
        cov = sum((x[i] - mean_x) * (y[i] - mean_y) for i in range(n))
        std_x = math.sqrt(sum((v - mean_x) ** 2 for v in x))
        std_y = math.sqrt(sum((v - mean_y) ** 2 for v in y))
        if std_x == 0 or std_y == 0:
            return {"ok": False, "error": "Standard deviation cannot be zero"}
        corr = cov / (std_x * std_y)
        return {"ok": True, "result": round(corr, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "covariance": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    x = event.get("x")
    y = event.get("y")
    if x is None or y is None:
        return {"ok": False, "error": "x and y are required"}
    try:
        x = [float(v) for v in x]
        y = [float(v) for v in y]
        n = len(x)
        if n != len(y):
            return {"ok": False, "error": "x and y must have same length"}
        if n < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        population = bool(event.get("population", False))
        mean_x = sum(x) / n
        mean_y = sum(y) / n
        cov = sum((x[i] - mean_x) * (y[i] - mean_y) for i in range(n))
        divisor = n if population else n - 1
        cov /= divisor
        return {"ok": True, "result": round(cov, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "variance-portfolio": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    returns = event.get("returns")
    if returns is None:
        return {"ok": False, "error": "returns is required"}
    try:
        r = [float(x) for x in returns]
        n = len(r)
        if n < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        population = bool(event.get("population", False))
        mean = sum(r) / n
        divisor = n if population else n - 1
        variance = sum((x - mean) ** 2 for x in r) / divisor
        return {"ok": True, "result": round(variance, 8), "mean": round(mean, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "var-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    returns = event.get("returns")
    if returns is None:
        return {"ok": False, "error": "returns is required"}
    try:
        r = sorted([float(x) for x in returns])
        n = len(r)
        if n < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        confidence = float(event.get("confidence_level", 0.95))
        portfolio_value = float(event.get("portfolio_value", 1))
        # Historical simulation VaR
        idx = int((1 - confidence) * n)
        var = r[idx] * portfolio_value
        return {"ok": True, "result": round(var, 8), "confidence_level": confidence}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "cvar-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    returns = event.get("returns")
    if returns is None:
        return {"ok": False, "error": "returns is required"}
    try:
        r = sorted([float(x) for x in returns])
        n = len(r)
        if n < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        confidence = float(event.get("confidence_level", 0.95))
        portfolio_value = float(event.get("portfolio_value", 1))
        cutoff = int((1 - confidence) * n)
        if cutoff == 0:
            cutoff = 1
        tail = r[:cutoff]
        cvar = (sum(tail) / len(tail)) * portfolio_value
        return {"ok": True, "result": round(cvar, 8), "confidence_level": confidence}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "beta-distribution": '''import math

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
''',

    "normal-distribution": '''import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    x = event.get("x")
    if x is None:
        return {"ok": False, "error": "x is required"}
    try:
        x = float(x)
        mu = float(event.get("mean", 0))
        sigma = float(event.get("std_dev", 1))
        if sigma <= 0:
            return {"ok": False, "error": "std_dev must be positive"}
        z = (x - mu) / sigma
        pdf = math.exp(-0.5 * z ** 2) / (sigma * math.sqrt(2 * math.pi))
        cdf = 0.5 * (1 + math.erf(z / math.sqrt(2)))
        return {"ok": True, "result": {"pdf": round(pdf, 8), "cdf": round(cdf, 8)}}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "log-normal-distribution": '''import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    x = event.get("x")
    if x is None:
        return {"ok": False, "error": "x is required"}
    try:
        x = float(x)
        mu = float(event.get("mu", 0))
        sigma = float(event.get("sigma", 1))
        if x <= 0:
            return {"ok": False, "error": "x must be positive"}
        if sigma <= 0:
            return {"ok": False, "error": "sigma must be positive"}
        pdf = (math.exp(-(math.log(x) - mu) ** 2 / (2 * sigma ** 2))
               / (x * sigma * math.sqrt(2 * math.pi)))
        z = (math.log(x) - mu) / sigma
        cdf = 0.5 * (1 + math.erf(z / math.sqrt(2)))
        return {"ok": True, "result": {"pdf": round(pdf, 8), "cdf": round(cdf, 8)}}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "monte-carlo-sim": '''import math
import random

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    initial = event.get("initial_value")
    mu = event.get("expected_return")
    sigma = event.get("volatility")
    years = event.get("years")
    if initial is None or mu is None or sigma is None or years is None:
        return {"ok": False, "error": "initial_value, expected_return, volatility, and years are required"}
    try:
        initial = float(initial)
        mu = float(mu)
        sigma = float(sigma)
        years = float(years)
        n_sim = int(event.get("simulations", 1000))
        seed = event.get("seed")
        if seed is not None:
            random.seed(int(seed))
        dt = 1 / 252
        steps = int(years * 252)
        final_values = []
        for _ in range(n_sim):
            value = initial
            for _ in range(steps):
                z = random.gauss(0, 1)
                value *= math.exp((mu - 0.5 * sigma ** 2) * dt + sigma * math.sqrt(dt) * z)
            final_values.append(value)
        final_values.sort()
        mean_val = sum(final_values) / n_sim
        median_val = final_values[n_sim // 2]
        p5 = final_values[int(0.05 * n_sim)]
        p95 = final_values[int(0.95 * n_sim)]
        return {
            "ok": True,
            "result": {
                "mean": round(mean_val, 2),
                "median": round(median_val, 2),
                "p5": round(p5, 2),
                "p95": round(p95, 2),
                "simulations": n_sim
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "historical-volatility": '''import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    prices = event.get("prices")
    if prices is None:
        return {"ok": False, "error": "prices is required"}
    try:
        prices = [float(p) for p in prices]
        if len(prices) < 2:
            return {"ok": False, "error": "At least 2 prices required"}
        log_returns = [math.log(prices[i] / prices[i - 1]) for i in range(1, len(prices))]
        n = len(log_returns)
        mean = sum(log_returns) / n
        variance = sum((r - mean) ** 2 for r in log_returns) / (n - 1)
        daily_vol = math.sqrt(variance)
        periods_per_year = float(event.get("periods_per_year", 252))
        annual_vol = daily_vol * math.sqrt(periods_per_year)
        return {
            "ok": True,
            "result": round(annual_vol, 8),
            "daily_volatility": round(daily_vol, 8),
            "annualized_volatility": round(annual_vol, 8)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "implied-volatility": '''import math

def _norm_cdf(x):
    return 0.5 * (1 + math.erf(x / math.sqrt(2)))

def _norm_pdf(x):
    return math.exp(-0.5 * x ** 2) / math.sqrt(2 * math.pi)

def _bs_price(S, K, T, r, sigma, option_type):
    if T <= 0 or sigma <= 0:
        return max(S - K, 0) if option_type == "call" else max(K - S, 0)
    d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
    d2 = d1 - sigma * math.sqrt(T)
    if option_type == "call":
        return S * _norm_cdf(d1) - K * math.exp(-r * T) * _norm_cdf(d2)
    else:
        return K * math.exp(-r * T) * _norm_cdf(-d2) - S * _norm_cdf(-d1)

def _vega(S, K, T, r, sigma):
    if T <= 0 or sigma <= 0:
        return 0
    d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
    return S * _norm_pdf(d1) * math.sqrt(T)

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["option_price", "spot_price", "strike_price", "time_to_expiry", "risk_free_rate"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        market_price = float(event["option_price"])
        S = float(event["spot_price"])
        K = float(event["strike_price"])
        T = float(event["time_to_expiry"])
        r = float(event["risk_free_rate"])
        option_type = str(event.get("option_type", "call")).lower()
        sigma = 0.2  # initial guess
        for _ in range(100):
            price = _bs_price(S, K, T, r, sigma, option_type)
            v = _vega(S, K, T, r, sigma)
            if abs(v) < 1e-10:
                break
            diff = price - market_price
            sigma -= diff / v
            if sigma <= 0:
                sigma = 1e-6
            if abs(diff) < 1e-8:
                break
        return {"ok": True, "result": round(sigma, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "black-scholes": '''import math

def _norm_cdf(x):
    return 0.5 * (1 + math.erf(x / math.sqrt(2)))

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate", "volatility"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        S = float(event["spot_price"])
        K = float(event["strike_price"])
        T = float(event["time_to_expiry"])
        r = float(event["risk_free_rate"])
        sigma = float(event["volatility"])
        option_type = str(event.get("option_type", "call")).lower()
        if T <= 0:
            intrinsic = max(S - K, 0) if option_type == "call" else max(K - S, 0)
            return {"ok": True, "result": round(intrinsic, 6)}
        d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
        d2 = d1 - sigma * math.sqrt(T)
        if option_type == "call":
            price = S * _norm_cdf(d1) - K * math.exp(-r * T) * _norm_cdf(d2)
        else:
            price = K * math.exp(-r * T) * _norm_cdf(-d2) - S * _norm_cdf(-d1)
        return {"ok": True, "result": round(price, 6), "d1": round(d1, 6), "d2": round(d2, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "greeks-calculate": '''import math

def _norm_cdf(x):
    return 0.5 * (1 + math.erf(x / math.sqrt(2)))

def _norm_pdf(x):
    return math.exp(-0.5 * x ** 2) / math.sqrt(2 * math.pi)

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate", "volatility"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        S = float(event["spot_price"])
        K = float(event["strike_price"])
        T = float(event["time_to_expiry"])
        r = float(event["risk_free_rate"])
        sigma = float(event["volatility"])
        option_type = str(event.get("option_type", "call")).lower()
        if T <= 0:
            return {"ok": False, "error": "time_to_expiry must be positive"}
        d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
        d2 = d1 - sigma * math.sqrt(T)
        if option_type == "call":
            delta = _norm_cdf(d1)
            rho = K * T * math.exp(-r * T) * _norm_cdf(d2)
        else:
            delta = _norm_cdf(d1) - 1
            rho = -K * T * math.exp(-r * T) * _norm_cdf(-d2)
        gamma = _norm_pdf(d1) / (S * sigma * math.sqrt(T))
        theta = ((-S * _norm_pdf(d1) * sigma / (2 * math.sqrt(T))
                  - r * K * math.exp(-r * T) * (_norm_cdf(d2) if option_type == "call" else _norm_cdf(-d2)))
                 / 365)
        vega = S * _norm_pdf(d1) * math.sqrt(T) / 100
        return {
            "ok": True,
            "result": {
                "delta": round(delta, 6),
                "gamma": round(gamma, 6),
                "theta": round(theta, 6),
                "vega": round(vega, 6),
                "rho": round(rho, 6)
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "option-price": '''import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate", "volatility"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        S = float(event["spot_price"])
        K = float(event["strike_price"])
        T = float(event["time_to_expiry"])
        r = float(event["risk_free_rate"])
        sigma = float(event["volatility"])
        N = int(event.get("steps", 100))
        option_type = str(event.get("option_type", "call")).lower()
        # Binomial tree (CRR model)
        dt = T / N
        u = math.exp(sigma * math.sqrt(dt))
        d = 1 / u
        p = (math.exp(r * dt) - d) / (u - d)
        # Terminal payoffs
        prices = [S * u ** (N - 2 * j) for j in range(N + 1)]
        if option_type == "call":
            values = [max(price - K, 0) for price in prices]
        else:
            values = [max(K - price, 0) for price in prices]
        # Backward induction
        discount = math.exp(-r * dt)
        for i in range(N - 1, -1, -1):
            values = [discount * (p * values[j] + (1 - p) * values[j + 1]) for j in range(i + 1)]
        return {"ok": True, "result": round(values[0], 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "put-call-parity": '''import math

def _norm_cdf(x):
    return 0.5 * (1 + math.erf(x / math.sqrt(2)))

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    S = event.get("spot_price")
    K = event.get("strike_price")
    T = event.get("time_to_expiry")
    r = event.get("risk_free_rate")
    if S is None or K is None or T is None or r is None:
        return {"ok": False, "error": "spot_price, strike_price, time_to_expiry, and risk_free_rate are required"}
    try:
        S = float(S)
        K = float(K)
        T = float(T)
        r = float(r)
        call_price = event.get("call_price")
        put_price = event.get("put_price")
        pv_strike = K * math.exp(-r * T)
        # C - P = S - PV(K)
        if call_price is not None and put_price is not None:
            call_price = float(call_price)
            put_price = float(put_price)
            lhs = call_price - put_price
            rhs = S - pv_strike
            parity_holds = abs(lhs - rhs) < 0.01
            theoretical_call = put_price + S - pv_strike
            theoretical_put = call_price - S + pv_strike
        elif call_price is not None:
            call_price = float(call_price)
            theoretical_put = call_price - S + pv_strike
            theoretical_call = call_price
            parity_holds = True
        elif put_price is not None:
            put_price = float(put_price)
            theoretical_call = put_price + S - pv_strike
            theoretical_put = put_price
            parity_holds = True
        else:
            return {"ok": False, "error": "At least one of call_price or put_price is required"}
        return {
            "ok": True,
            "result": {
                "parity_holds": parity_holds,
                "theoretical_call": round(theoretical_call, 6),
                "theoretical_put": round(theoretical_put, 6),
                "pv_strike": round(pv_strike, 6)
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "delta-hedge": '''import math

def _norm_cdf(x):
    return 0.5 * (1 + math.erf(x / math.sqrt(2)))

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate", "volatility"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        S = float(event["spot_price"])
        K = float(event["strike_price"])
        T = float(event["time_to_expiry"])
        r = float(event["risk_free_rate"])
        sigma = float(event["volatility"])
        option_type = str(event.get("option_type", "call")).lower()
        num_contracts = float(event.get("num_contracts", 1))
        contract_size = float(event.get("contract_size", 100))
        if T <= 0:
            return {"ok": False, "error": "time_to_expiry must be positive"}
        d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
        if option_type == "call":
            delta = _norm_cdf(d1)
        else:
            delta = _norm_cdf(d1) - 1
        shares_to_hedge = delta * num_contracts * contract_size
        return {
            "ok": True,
            "result": {
                "delta": round(delta, 6),
                "shares_to_short": round(shares_to_hedge, 4),
                "hedge_cost": round(shares_to_hedge * S, 2)
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "yield-to-maturity": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["face_value", "coupon_rate", "years_to_maturity", "market_price"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        F = float(event["face_value"])
        c_rate = float(event["coupon_rate"])
        years = float(event["years_to_maturity"])
        P = float(event["market_price"])
        freq = int(event.get("periods_per_year", 2))
        n = int(years * freq)
        coupon = F * c_rate / freq
        # Newton-Raphson to find YTM
        ytm = c_rate  # initial guess
        for _ in range(1000):
            r = ytm / freq
            price = sum(coupon / (1 + r) ** t for t in range(1, n + 1)) + F / (1 + r) ** n
            dprice = sum(-t * coupon / (1 + r) ** (t + 1) for t in range(1, n + 1)) - n * F / (1 + r) ** (n + 1)
            dprice /= freq
            diff = price - P
            if abs(dprice) < 1e-12:
                break
            new_ytm = ytm - diff / dprice
            if abs(new_ytm - ytm) < 1e-10:
                ytm = new_ytm
                break
            ytm = new_ytm
        return {"ok": True, "result": round(ytm, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "bond-price": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["face_value", "coupon_rate", "years_to_maturity", "yield_to_maturity"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        F = float(event["face_value"])
        c_rate = float(event["coupon_rate"])
        years = float(event["years_to_maturity"])
        ytm = float(event["yield_to_maturity"])
        freq = int(event.get("periods_per_year", 2))
        n = int(years * freq)
        coupon = F * c_rate / freq
        r = ytm / freq
        if r == 0:
            price = coupon * n + F
        else:
            price = sum(coupon / (1 + r) ** t for t in range(1, n + 1)) + F / (1 + r) ** n
        return {"ok": True, "result": round(price, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "duration-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["face_value", "coupon_rate", "years_to_maturity", "yield_to_maturity"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        F = float(event["face_value"])
        c_rate = float(event["coupon_rate"])
        years = float(event["years_to_maturity"])
        ytm = float(event["yield_to_maturity"])
        freq = int(event.get("periods_per_year", 2))
        n = int(years * freq)
        coupon = F * c_rate / freq
        r = ytm / freq
        if r == 0:
            price = coupon * n + F
            weighted_time = sum(t * coupon for t in range(1, n + 1)) + n * F
            macaulay = weighted_time / price / freq
        else:
            price = sum(coupon / (1 + r) ** t for t in range(1, n + 1)) + F / (1 + r) ** n
            weighted_time = sum(t * coupon / (1 + r) ** t for t in range(1, n + 1)) + n * F / (1 + r) ** n
            macaulay = weighted_time / price / freq
        modified = macaulay / (1 + ytm / freq)
        return {
            "ok": True,
            "result": {"macaulay_duration": round(macaulay, 6), "modified_duration": round(modified, 6)}
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "convexity-calculate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["face_value", "coupon_rate", "years_to_maturity", "yield_to_maturity"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        F = float(event["face_value"])
        c_rate = float(event["coupon_rate"])
        years = float(event["years_to_maturity"])
        ytm = float(event["yield_to_maturity"])
        freq = int(event.get("periods_per_year", 2))
        n = int(years * freq)
        coupon = F * c_rate / freq
        r = ytm / freq
        if r == 0:
            price = coupon * n + F
            convexity_sum = sum(t * (t + 1) * coupon for t in range(1, n + 1)) + n * (n + 1) * F
        else:
            price = sum(coupon / (1 + r) ** t for t in range(1, n + 1)) + F / (1 + r) ** n
            convexity_sum = (sum(t * (t + 1) * coupon / (1 + r) ** (t + 2) for t in range(1, n + 1))
                             + n * (n + 1) * F / (1 + r) ** (n + 2))
        convexity = convexity_sum / (price * freq ** 2)
        return {"ok": True, "result": round(convexity, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "credit-spread": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    corp_yield = event.get("corporate_yield")
    rfr = event.get("risk_free_rate")
    if corp_yield is None or rfr is None:
        return {"ok": False, "error": "corporate_yield and risk_free_rate are required"}
    try:
        corp_yield = float(corp_yield)
        rfr = float(rfr)
        spread = corp_yield - rfr
        spread_bps = spread * 10000
        return {"ok": True, "result": round(spread, 8), "result_bps": round(spread_bps, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "default-probability": '''import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    credit_spread = event.get("credit_spread")
    if credit_spread is None:
        return {"ok": False, "error": "credit_spread is required"}
    try:
        spread = float(credit_spread)
        recovery = float(event.get("recovery_rate", 0.4))
        years = float(event.get("years", 1))
        if recovery >= 1:
            return {"ok": False, "error": "recovery_rate must be less than 1"}
        # Hazard rate approximation: lambda = spread / (1 - recovery)
        hazard_rate = spread / (1 - recovery)
        # Probability of default over horizon
        prob_default = 1 - math.exp(-hazard_rate * years)
        return {
            "ok": True,
            "result": round(prob_default, 8),
            "hazard_rate": round(hazard_rate, 8),
            "survival_probability": round(1 - prob_default, 8)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',

    "recovery-rate": '''def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    face_value = event.get("face_value")
    recovery_amount = event.get("recovery_amount")
    if face_value is None or recovery_amount is None:
        return {"ok": False, "error": "face_value and recovery_amount are required"}
    try:
        face_value = float(face_value)
        recovery_amount = float(recovery_amount)
        accrued = float(event.get("accrued_interest", 0))
        total_claim = face_value + accrued
        if total_claim == 0:
            return {"ok": False, "error": "face_value cannot be zero"}
        recovery_rate = recovery_amount / total_claim
        lgd = 1 - recovery_rate
        loss_amount = total_claim - recovery_amount
        return {
            "ok": True,
            "result": round(recovery_rate, 6),
            "lgd": round(lgd, 6),
            "loss_amount": round(loss_amount, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
''',
}


def create_function(func_def):
    name = func_def[0]
    func_dir = os.path.join(BASE_DIR, name)
    os.makedirs(func_dir, exist_ok=True)
    
    # Write JSONC
    jsonc_path = os.path.join(func_dir, "functionfly.jsonc")
    jsonc_content = make_jsonc(func_def)
    with open(jsonc_path, "w") as f:
        f.write(jsonc_content)
    
    # Write Python
    py_path = os.path.join(func_dir, "main.py")
    py_content = IMPLEMENTATIONS.get(name, f'def handler(event):\n    return {{"ok": False, "error": "not implemented"}}\n')
    with open(py_path, "w") as f:
        f.write(py_content)
    
    print(f"Created: {name}")


if __name__ == "__main__":
    for func_def in FUNCTIONS:
        create_function(func_def)
    print(f"\nTotal functions created: {len(FUNCTIONS)}")
