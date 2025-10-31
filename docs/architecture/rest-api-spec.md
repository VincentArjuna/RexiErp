# REST API Spec

```yaml
openapi: 3.0.0
info:
  title: RexiERP Backend API
  version: v1.0.0
  description: Comprehensive ERP system for Indonesian MSMEs with tax compliance and government integrations
servers:
  - url: https://api.rexi-erp.id/v1
    description: Production server (Indonesia)
  - url: https://staging-api.rexi-erp.id/v1
    description: Staging server
  - url: http://localhost:8080/v1
    description: Local development server

paths:
  /auth/login:
    post:
      tags:
        - Authentication
      summary: User login
      description: Authenticate user and return JWT token
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - email
                - password
              properties:
                email:
                  type: string
                  format: email
                  example: "user@company.co.id"
                password:
                  type: string
                  format: password
                  example: "SecurePass123!"
      responses:
        '200':
          description: Login successful
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    type: object
                    properties:
                      token:
                        type: string
                        example: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
                      user:
                        $ref: '#/components/schemas/User'
                      expires_in:
                        type: integer
                        example: 3600
        '401':
          description: Invalid credentials
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /products:
    get:
      tags:
        - Inventory
      summary: List products
      description: Get paginated list of products with filters
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
        - name: category
          in: query
          schema:
            type: string
        - name: search
          in: query
          schema:
            type: string
      responses:
        '200':
          description: Products retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    type: object
                    properties:
                      products:
                        type: array
                        items:
                          $ref: '#/components/schemas/Product'
                      pagination:
                        $ref: '#/components/schemas/Pagination'

    post:
      tags:
        - Inventory
      summary: Create product
      description: Create new product with automatic SKU generation
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateProductRequest'
      responses:
        '201':
          description: Product created successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    $ref: '#/components/schemas/Product'

  /products/{id}:
    get:
      tags:
        - Inventory
      summary: Get product by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Product retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    $ref: '#/components/schemas/Product'
        '404':
          description: Product not found

  /invoices:
    post:
      tags:
        - Accounting
      summary: Create invoice
      description: Create sales invoice with Indonesian tax compliance
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateInvoiceRequest'
      responses:
        '201':
          description: Invoice created successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    $ref: '#/components/schemas/Invoice'

  /invoices/{id}/e-faktur:
    post:
      tags:
        - Accounting
      summary: Submit to e-Faktur
      description: Submit invoice to Indonesian e-Faktur system
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: e-Faktur submitted successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    example: true
                  data:
                    type: object
                    properties:
                      efaktur_number:
                        type: string
                        example: "010.000-25.12345678"
                      status:
                        type: string
                        example: "submitted"

components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        full_name:
          type: string
        role:
          type: string
          enum: [super_admin, tenant_admin, staff, viewer]
        tenant_id:
          type: string
          format: uuid
        is_active:
          type: boolean
        created_at:
          type: string
          format: date-time

    Product:
      type: object
      properties:
        id:
          type: string
          format: uuid
        tenant_id:
          type: string
          format: uuid
        sku:
          type: string
        name:
          type: string
        description:
          type: string
        category_id:
          type: string
          format: uuid
        unit:
          type: string
        purchase_price:
          type: number
          format: decimal
        selling_price:
          type: number
          format: decimal
        tax_rate:
          type: number
          format: decimal
          example: 0.11
        is_taxable:
          type: boolean
        stock_quantity:
          type: integer
        min_stock:
          type: integer
        created_at:
          type: string
          format: date-time

    CreateProductRequest:
      type: object
      required:
        - name
        - category_id
        - unit
        - purchase_price
        - selling_price
      properties:
        name:
          type: string
        description:
          type: string
        category_id:
          type: string
          format: uuid
        unit:
          type: string
        purchase_price:
          type: number
          format: decimal
        selling_price:
          type: number
          format: decimal
        is_taxable:
          type: boolean
          default: true
        min_stock:
          type: integer
          default: 0

    Invoice:
      type: object
      properties:
        id:
          type: string
          format: uuid
        tenant_id:
          type: string
          format: uuid
        invoice_number:
          type: string
        customer_id:
          type: string
          format: uuid
        invoice_date:
          type: string
          format: date
        due_date:
          type: string
          format: date
        status:
          type: string
          enum: [draft, sent, paid, overdue, cancelled]
        subtotal:
          type: number
          format: decimal
        tax_amount:
          type: number
          format: decimal
        discount_amount:
          type: number
          format: decimal
        total_amount:
          type: number
          format: decimal
        paid_amount:
          type: number
          format: decimal
        efaktur_number:
          type: string
        efaktur_status:
          type: string
          enum: [pending, submitted, approved, rejected]
        efaktur_url:
          type: string
        notes:
          type: string
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    CreateInvoiceRequest:
      type: object
      required:
        - customer_id
        - items
      properties:
        customer_id:
          type: string
          format: uuid
        items:
          type: array
          items:
            type: object
            required:
              - product_id
              - quantity
              - unit_price
            properties:
              product_id:
                type: string
                format: uuid
              quantity:
                type: integer
                minimum: 1
              unit_price:
                type: number
                format: decimal
                minimum: 0

    Pagination:
      type: object
      properties:
        current_page:
          type: integer
        total_pages:
          type: integer
        total_items:
          type: integer
        items_per_page:
          type: integer

    Error:
      type: object
      properties:
        success:
          type: boolean
          example: false
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
            details:
              type: object

  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []
```
