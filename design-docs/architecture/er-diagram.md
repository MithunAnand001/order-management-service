# Finalized Entity Relationship Diagram (ERD) - v3 (Enum RBAC & Addresses)

This diagram represents the production-ready schema including Enum-based Role-Based Access Control (RBAC) and User Address management.

```mermaid
erDiagram
    USER ||--o{ USER_ADDRESS : "maintains"
    USER ||--o{ ORDER : "places"
    USER_ADDRESS ||--o{ ORDER : "shipping destination"
    PRODUCT ||--o{ ORDER_ITEM : "defined in"
    ORDER ||--|{ ORDER_ITEM : "contains"
    ORDER ||--o{ ORDER_EVENT_LOG : "tracks"

    USER {
        uint id PK
        uuid uuid UK
        string role "Enum: ADMIN, USER, DELIVERY"
        string name
        string email UK
        string password
    }

    USER_ADDRESS {
        uint id PK
        uuid uuid UK
        uint user_id FK "References USER.id"
        string address_line1
        string address_line2
        string city
        string state
        string postal_code
        string country
        bool is_current "Determines default address"
        bool is_active
    }

    PRODUCT {
        uint id PK
        uuid uuid UK
        string sku UK
        string name
        decimal current_price
        int stock_quantity
    }

    ORDER {
        uint id PK
        uuid uuid UK
        uint user_id FK "References USER.id"
        uint address_id FK "References USER_ADDRESS.id"
        string status "Enum: PENDING, PROCESSING, OUT_FOR_DELIVERY, SHIPPED, DELIVERED, CANCELLED"
        decimal total_amount
    }

    ORDER_ITEM {
        uint id PK
        uint order_id FK
        uint product_id FK
        int quantity
        decimal unit_price_snapshot
        decimal subtotal
    }

    ORDER_EVENT_LOG {
        uint id PK
        uint order_id FK
        string from_status
        string to_status
        string reason
        string triggered_by "User UUID"
    }
```

## Entity Details

### 1. User (Updated)
- **role**: Implemented as an enum (text in DB) for simplicity and performance. Valid values: `ADMIN`, `USER`, `DELIVERY`.

### 2. User Address
- **is_current**: Only one address per user can be marked as current. This is managed atomically in the Repository layer.

### 3. Order
- **address_id**: Atomic link to the specific address at the time of order.
- **status**: Managed via role-based transition rules in the Service layer.

### 4. Atomic Integrity (Architectural Input)
- **Stock Movements**: Handled within DB transactions during `CreateOrder` (Deduction) and `CancelOrder`/`Status Update to CANCELLED` (Replenishment).
