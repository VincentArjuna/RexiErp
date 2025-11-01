# Data Models

## Users

**Purpose:** Core user management with role-based access control for multi-tenant Indonesian MSME operations

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- email: string - User login and notifications (Indonesian email validation)
- password_hash: string - Secure password storage
- full_name: string - User's full name (Indonesian naming conventions)
- phone_number: string - Indonesian mobile number format (+62)
- role: enum - Role-based permissions (super_admin, tenant_admin, staff, viewer)
- is_active: boolean - Account status
- last_login: timestamp - Security tracking
- created_at: timestamp - Audit trail
- updated_at: timestamp - Audit trail

**Relationships:**
-belongsTo Tenant (tenant_id)
-hasMany UserSessions
-hasMany ActivityLogs

## UserSessions

**Purpose:** JWT session management with Redis-based token blacklisting for multi-tenant authentication security

**Key Attributes:**
- id: UUID - Primary identifier
- user_id: UUID - Foreign key to Users table
- tenant_id: UUID - Multi-tenant isolation
- session_id: string - Unique session identifier
- token_hash: string - Hashed JWT token for blacklisting
- refresh_token_hash: string - Hashed refresh token
- device_info: json - Device fingerprinting data
- ip_address: string - Client IP address for security tracking
- user_agent: string - Browser/client identification
- expires_at: timestamp - Token expiration time
- last_activity: timestamp - Session activity tracking
- is_active: boolean - Session status (true=active, false=revoked)
- created_at: timestamp - Session creation
- updated_at: timestamp - Last modification

**Relationships:**
- belongsTo User (user_id)
- belongsTo Tenant (tenant_id)

**Indexes:**
- Unique index on session_id
- Index on user_id for user session lookups
- Index on token_hash for blacklisting checks
- Index on expires_at for cleanup operations

## ActivityLogs

**Purpose:** Security audit trail for all user actions with multi-tenant context and compliance tracking

**Key Attributes:**
- id: UUID - Primary identifier
- user_id: UUID - Foreign key to Users table (nullable for system actions)
- tenant_id: UUID - Multi-tenant isolation
- action: string - Action performed (login, logout, create, update, delete, etc.)
- resource_type: string - Type of resource affected (user, product, customer, etc.)
- resource_id: UUID - ID of affected resource (nullable)
- old_values: json - Previous state for audit (nullable)
- new_values: json - New state for audit (nullable)
- ip_address: string - Client IP address for security tracking
- user_agent: string - Browser/client identification
- session_id: string - Related session identifier
- success: boolean - Action success status
- error_message: string - Error details if action failed (nullable)
- context: json - Additional context data (tenant context, request details, etc.)
- created_at: timestamp - Activity timestamp

**Relationships:**
- belongsTo User (user_id, nullable)
- belongsTo Tenant (tenant_id)

**Indexes:**
- Index on user_id for user activity lookups
- Index on tenant_id for multi-tenant queries
- Index on action for activity type filtering
- Index on resource_type + resource_id for resource tracking
- Index on created_at for time-based queries
- Composite index on (tenant_id, created_at) for tenant audit reports

## Tenants

**Purpose:** Multi-tenant management for Indonesian MSME customers with company-specific data isolation

**Key Attributes:**
- id: UUID - Primary identifier
- company_name: string - Legal company name (Indonesian PT/CV/Firma formats)
- company_type: enum - Business entity type (PT, CV, Firma, Udaha)
- tax_id: string - NPWP (Indonesian Taxpayer Identification Number)
- business_license: string - SIUP/NIB (Indonesian Business License)
- address: text - Full Indonesian address format
- province: string - Indonesian province code
- city: string - Indonesian city code
- postal_code: string - Indonesian postal code
- phone: string - Company phone (Indonesian format)
- email: string - Company email
- is_active: boolean - Tenant status
- subscription_plan: enum - Subscription tier (basic, professional, enterprise)
- max_users: integer - User limits per plan
- created_at: timestamp - Account creation
- updated_at: timestamp - Last modifications

**Relationships:**
-hasMany Users
-hasMany Companies
-hasMany Warehouses
-hasMany Products
-hasMany Customers
-hasMany Suppliers
-hasMany Transactions

## Products

**Purpose:** Product catalog management with Indonesian tax compliance and inventory tracking

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- sku: string - Stock Keeping Unit (unique per tenant)
- barcode: string - Product barcode (Indonesian format)
- name: string - Product name (Indonesian/English)
- description: text - Product description
- category_id: UUID - Product categorization
- unit: string - Measurement unit (pcs, kg, liter, etc.)
- purchase_price: decimal - Cost price (IDR)
- selling_price: decimal - Retail price (IDR)
- tax_rate: decimal - PPN rate (11% standard, other rates)
- is_taxable: boolean - Tax eligibility
- stock_quantity: integer - Current stock
- min_stock: integer - Reorder point
- max_stock: integer - Maximum stock
- weight: decimal - Shipping weight (kg)
- dimensions: json - Product dimensions (LxWxH cm)
- is_active: boolean - Product status
- created_at: timestamp - Product creation
- updated_at: timestamp - Last modification

**Relationships:**
-belongsTo Tenant
-belongsTo Category
-hasMany InventoryTransactions
-hasMany OrderItems
-hasMany PurchaseOrderItems

## Customers

**Purpose:** Indonesian customer management with tax compliance and billing information

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- customer_code: string - Customer identifier (auto-generated)
- name: string - Customer name (Indonesian naming conventions)
- email: string - Customer email
- phone: string - Indonesian phone format
- address: text - Full Indonesian address
- province: string - Indonesian province code
- city: string - Indonesian city code
- postal_code: string - Indonesian postal code
- tax_id: string - NPWP (for tax invoices)
- customer_type: enum - Individual/Company/B2B
- credit_limit: decimal - Credit limit (IDR)
- is_active: boolean - Customer status
- created_at: timestamp - Customer creation
- updated_at: timestamp - Last modification

**Relationships:**
-belongsTo Tenant
-hasMany SalesOrders
-hasMany Invoices
-hasMany Payments

## Suppliers

**Purpose:** Indonesian supplier management with procurement and tax compliance

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- supplier_code: string - Supplier identifier
- name: string - Supplier company name
- contact_person: string - Contact person
- email: string - Supplier email
- phone: string - Indonesian phone format
- address: text - Full Indonesian address
- tax_id: string - Supplier NPWP
- payment_terms: integer - Payment terms in days
- is_active: boolean - Supplier status
- created_at: timestamp - Supplier creation
- updated_at: timestamp - Last modification

**Relationships:**
-belongsTo Tenant
-hasMany PurchaseOrders
-hasMany Bills

## Transactions

**Purpose:** Financial transaction tracking with Indonesian tax compliance and audit trails

**Key Attributes:**
- id: UUID - Primary identifier
- tenant_id: UUID - Multi-tenant isolation
- transaction_number: string - Auto-generated transaction number
- transaction_type: enum - Sale/Purchase/Return/Adjustment
- reference_id: UUID - Related document ID (order/invoice/etc.)
- customer_id: UUID - Customer reference (nullable)
- supplier_id: UUID - Supplier reference (nullable)
- total_amount: decimal - Total amount (IDR)
- tax_amount: decimal - PPN tax amount
- discount_amount: decimal - Discount amount
- final_amount: decimal - Final amount paid
- payment_method: enum - Cash/Bank Transfer/E-wallet
- payment_status: enum - Pending/Paid/Overdue/Cancelled
- transaction_date: timestamp - Transaction date
- due_date: timestamp - Payment due date
- notes: text - Transaction notes
- created_at: timestamp - Transaction creation
- updated_at: timestamp - Last modification

**Relationships:**
-belongsTo Tenant
-belongsTo Customer
-belongsTo Supplier
-hasMany TransactionItems
-hasMany Payments
