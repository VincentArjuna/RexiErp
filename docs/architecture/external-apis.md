# External APIs

## e-Faktur API

**Purpose:** Indonesian tax compliance for electronic tax invoices (e-Faktur) required by DJP (Direktorat Jenderal Pajak)

- **Documentation:** https://efaktur.pajak.go.id/
- **Base URL(s):** https://api.efaktur.pajak.go.id/v1/
- **Authentication:** OAuth 2.0 with client credentials + digital certificate
- **Rate Limits:** 100 requests per minute, 5000 requests per day

**Key Endpoints Used:**
- `POST /invoices` - Create e-Faktur invoice
- `GET /invoices/{id}` - Retrieve e-Faktur status
- `POST /invoices/{id}/cancel` - Cancel e-Faktur
- `GET /tax-rates` - Get current PPN tax rates
- `POST /validation` - Validate taxpayer data

**Integration Notes:**
- Requires digital certificate for authentication
- Indonesian timezone (WIB/UTC+7) for all timestamps
- Retry logic essential due to government API reliability issues
- 30-day retention period for invoice data

## BPJS API

**Purpose:** Indonesian social security program (BPJS Ketenagakerjaan & BPJS Kesehatan) integration for employee contributions

- **Documentation:** https://api.bpjsketenagakerjaan.go.id/ and https://api.bpjs-kesehatan.go.id/
- **Base URL(s):**
  - Employment: https://api.bpjsketenagakerjaan.go.id/v1/
  - Health: https://api.bpjs-kesehatan.go.id/v1/
- **Authentication:** Consumer Key + Consumer Secret + HMAC signature
- **Rate Limits:** 150 requests per minute per service

**Key Endpoints Used:**
- `POST /participant/verification` - Verify employee BPJS status
- `GET /contributions/{period}` - Get contribution rates
- `POST /contributions/calculate` - Calculate employee contributions
- `GET /participants/{nik}` - Get participant details
- `POST /payment/report` - Report contribution payments

**Integration Notes:**
- Separate APIs for employment and health BPJS
- Indonesian ID number (NIK) validation required
- Monthly contribution periods (1st-21st of each month)
- Complex calculation rules for different employee categories

## e-Invoice API

**Purpose:** Indonesian electronic invoice system for B2G (Business-to-Government) transactions

- **Documentation:** https://einvoice.pajak.go.id/
- **Base URL(s):** https://api.einvoice.pajak.go.id/v1/
- **Authentication:** Digital certificate + OAuth 2.0
- **Rate Limits:** 200 requests per minute

**Key Endpoints Used:**
- `POST /invoices/create` - Create electronic invoice
- `GET /invoices/status/{id}` - Check invoice status
- `POST /invoices/cancel` - Cancel electronic invoice
- `GET /partners` - Get government partner information

**Integration Notes:**
- Mandatory for B2G transactions above certain thresholds
- Requires pre-approval from tax authorities
- Integration with existing accounting system critical

## Bank API Integrations

**Purpose:** Payment processing and bank account verification for Indonesian banks

- **Documentation:** Varies by bank (BCA, Mandiri, BNI, BRI, etc.)
- **Base URL(s):** Bank-specific APIs
- **Authentication:** OAuth 2.0 + API keys
- **Rate Limits:** Varies by bank (typically 100-500 requests/minute)

**Key Endpoints Used:**
- `POST /payments/verify` - Verify payment status
- `GET /accounts/{number}/validate` - Validate bank accounts
- `POST /payments/disbursement` - Process payments
- `GET /transactions/history` - Get transaction history

**Integration Notes:**
- Multiple bank integrations required for Indonesian MSMEs
- BI (Bank Indonesia) compliance for payment processing
- Real-time verification vs batch processing options

## SMS Gateway API

**Purpose:** SMS notifications for Indonesian mobile numbers (transaction alerts, OTP, notifications)

- **Documentation:** Provider-specific (e.g., Telkom, Vonage, Twilio Indonesia)
- **Base URL(s):** Provider-specific
- **Authentication:** API key + HMAC signature
- **Rate Limits:** Provider-specific (typically 100 SMS/minute)

**Key Endpoints Used:**
- `POST /sms/send` - Send SMS notification
- `GET /sms/delivery/{id}` - Check delivery status
- `GET /sms/balance` - Check SMS balance

**Integration Notes:**
- Indonesian mobile number format validation (+62)
- Provider selection for cost optimization
- Delivery reliability considerations
