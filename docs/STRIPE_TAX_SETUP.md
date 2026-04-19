# Tax/VAT Compliance Implementation

This document describes the Stripe Tax integration for FunctionFly's billing system, enabling automatic tax calculation for EU VAT, US sales tax, and global tax compliance.

## Overview

The tax compliance implementation consists of:

1. **Database Schema** - Stores customer tax settings (country, state, postal code, tax ID)
2. **Stripe Tax Integration** - Automatic tax calculation on all transactions
3. **Tax ID Validation** - Client-side validation for EU VAT, US EIN, and other formats
4. **API Endpoints** - Customer-facing endpoints for managing tax settings
5. **Billing Portal** - Customer self-service for tax ID management

## Supported Tax Types

### European Union (EU VAT)

- Automatic VAT calculation for all EU countries
- VAT ID validation (format validation for all 27 EU member states)
- Reverse charge handling for B2B transactions with valid VAT ID
- MOSS (Mini One Stop Shop) compliance

### United States (Sales Tax)

- State-level sales tax calculation
- Local jurisdiction taxes (county, city, special districts)
- Economic nexus compliance (all states with sales tax)
- Tax-exempt handling for non-profit/resale certificates

### Other Regions

- **UK** - VAT (post-Brexit rules)
- **Canada** - GST/HST (federal and provincial)
- **Australia** - GST
- **New Zealand** - GST
- **Singapore** - GST
- **Switzerland** - VAT
- **Norway** - VAT
- **Japan** - Consumption Tax
- **India** - GST

## Database Schema

### New Fields on `tenants` table

| Field | Type | Description |
|-------|------|-------------|
| `billing_country` | VARCHAR(2) | ISO 3166-1 alpha-2 country code |
| `billing_state` | VARCHAR(50) | State/Province (required for US, Canada) |
| `billing_postal_code` | VARCHAR(20) | Postal/ZIP code for jurisdiction |
| `tax_id` | VARCHAR(50) | Tax ID (VAT, EIN, GST, etc.) |
| `tax_id_type` | VARCHAR(20) | Type: eu_vat, us_ein, ca_gst, etc. |
| `tax_status` | VARCHAR(20) | pending, valid, invalid, exempt |
| `tax_exempt` | BOOLEAN | Tax exempt flag |
| `stripe_tax_location_id` | VARCHAR(255) | Stripe Tax Location ID |
| `stripe_customer_tax_id` | VARCHAR(255) | Stripe Customer Tax ID object ID |

### New Tables

#### `tax_rates`

Cached tax rates from Stripe for reporting purposes.

#### `invoice_tax_details`

Tax breakdown per invoice (tax amount, rate, jurisdiction).

#### `tax_id_validation_logs`

Audit log for tax ID validation attempts.

## API Endpoints

### Get Tax Settings

```
GET /v1/billing/tax/settings
```

Returns current tenant's tax settings and applicable tax type.

### Update Tax Settings

```
POST /v1/billing/tax/settings
{
  "billing_country": "DE",
  "billing_state": "",
  "billing_postal_code": "10115",
  "tax_id": "DE123456789",
  "tax_id_type": "eu_vat",
  "tax_exempt": false
}
```

Updates tax settings with validation.

### Get Tax Types for Country

```
GET /v1/billing/tax/types?country=DE
```

Returns applicable tax types and requirements for a country.

### Calculate Tax

```
POST /v1/billing/tax/calculate
{
  "amount_cents": 10000,
  "currency": "USD",
  "transaction_type": "subscription"
}
```

Calculates tax for a hypothetical transaction.

### Validate Tax ID

```
POST /v1/billing/tax/validate
{
  "tax_id": "DE123456789",
  "tax_id_type": "eu_vat"
}
```

Validates a tax ID format without saving it.

## Stripe Configuration

### Required Stripe Dashboard Configuration

1. **Enable Stripe Tax**
   - Go to: <https://dashboard.stripe.com/settings/tax>
   - Enable Stripe Tax for your account
   - Configure tax jurisdictions (EU, US, etc.)

2. **Configure Customer Portal**
   - Go to: <https://dashboard.stripe.com/settings/billing/portal>
   - Enable:
     - "Tax ID collection"
     - "Customer address updates"
     - "Invoice history with tax breakdowns"
   - This allows customers to manage their tax settings

3. **Set Tax Behavior on Products/Prices**
   - In Stripe Dashboard, set tax behavior to "exclusive" (tax added to price)
   - For reverse charge eligible B2B: set to "exclusive" with customer tax exemption

### Environment Variables

No new environment variables are required. The existing `STRIPE_SECRET_KEY` is used.

## Automatic Tax Calculation

All checkout sessions now include `automatic_tax: { enabled: true }`, which enables:

- Real-time tax calculation based on customer location
- VAT ID validation and reverse charge handling
- US sales tax calculation with full jurisdiction breakdown
- Tax-inclusive pricing where required by law

### Checkout Session Updates

All checkout types now include automatic tax:

- Subscription checkouts
- Bundle subscriptions
- State Fabric add-on subscriptions
- One-time payments (credits, verification fees, username changes)

## Tax ID Formats

### EU VAT Numbers

Format: `CCXXXXXXXXX` (2-letter country code + 2-12 alphanumeric)
Examples:

- Germany: `DE123456789`
- France: `FR12345678901`
- UK: `GB123456789` (Northern Ireland: `XI123456789`)

### US EIN

Format: `XXXXXXXXX` (9 digits)
Example: `12-3456787` (dashes are stripped)

### Canada GST/HST

Format: `XXXXXXXXXRT0001`
Example: `123456789RT0001`

### Australia ABN

Format: `XXXXXXXXXXX` (11 digits)
Example: `51824753556`

### UK VAT (Post-Brexit)

Format: `GBXXXXXXXXX` (GB + 9 digits)
Example: `GB123456789`

### Switzerland VAT (UID/MWST)

Format: `CHE-XXX.XXX.XXX` or `CHEXXXXXXXXX`
Example: `CHE-123.456.789`

## Implementation Notes

### Reverse Charge (B2B)

When a valid EU VAT ID is provided:

- Stripe automatically applies reverse charge rules
- No VAT is charged on the invoice
- Customer is responsible for remitting VAT

### US Sales Tax

- Stripe monitors economic nexus thresholds
- Automatically registers in new jurisdictions when thresholds are met
- Handles filing and remittance (with Stripe Tax add-on)

### Tax Exemptions

- `tax_exempt: true` prevents tax calculation
- Exemption certificates can be stored in Stripe Customer

## Testing

### Test Mode

In Stripe test mode:

- Use test tax ID formats
- Tax calculations return realistic mock data
- No actual tax liability is created

### Test Tax IDs

- EU VAT: Use any valid format (validation is format-only, not VIES lookup)
- US: Any 9 digits
- For VIES lookup testing: Use real VAT numbers in live mode only

## Compliance Notes

### EU VAT

- VAT is charged on B2C transactions
- Reverse charge applied on B2B with valid VAT ID
- Invoices include VAT number and rate

### US Sales Tax

- Nexus monitoring for all states
- Real-time tax rate lookup
- Jurisdiction-level reporting

### Record Keeping

- All tax calculations are stored in `invoice_tax_details`
- Stripe Tax provides official tax reports
- Transaction-level audit trail

## Migration

To apply the tax compliance schema:

```bash
# Run the migration
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migrations/20260418000001_add_tax_compliance.up.sql
```

For existing customers:

- Tax status defaults to "pending"
- Billing country should be collected during next checkout
- Retroactive tax calculation not applied

## Future Enhancements

- VIES (VAT Information Exchange System) real-time validation
- Stripe Tax reporting API integration
- Automated tax exemption certificate management
- Multi-jurisdiction tax filing automation
