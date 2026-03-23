# Design Principles & DB Management Checklist

This checklist ensures that the system maintains high-level architectural integrity and database efficiency.

## 1. Design Principles (SOLID & Clean Code)
- [ ] **Single Responsibility (S):** Does each package or struct have only one reason to change? (e.g., `UserRepository` only knows about DB, not business rules).
- [ ] **Open/Closed (O):** Can we add new features (like a new notification type) without modifying existing service code? (Use of Interfaces).
- [ ] **Interface Segregation (I):** Are interfaces lean and focused? (Avoid large, "god" interfaces).
- [ ] **Dependency Inversion (D):** Do high-level services depend on abstractions (interfaces) rather than concrete implementations (rabbitmq/postgres)?
- [ ] **KISS (Keep It Simple, Stupid):** Is the logic straightforward? Avoid over-engineering patterns where a simple function would suffice.
- [ ] **DRY (Don't Repeat Yourself):** Are common logic patterns (like error formatting or date parsing) extracted into `internal/utils`?
- [ ] **YAGNI (You Ain't Gonna Need It):** Have we avoided adding features or "just-in-case" fields not defined in the original requirements?

## 2. Database Management & Integrity
- [ ] **Atomic Transactions:** Are all multi-step writes (e.g., creating an order and deducting stock) wrapped in a GORM transaction?
- [ ] **Data Integrity:** Are we using database-level constraints (`not null`, `uniqueIndex`) instead of relying solely on application-level checks?
- [ ] **Indexing Strategy:** 
    - [ ] Are all foreign keys indexed?
    - [ ] Are columns used in `WHERE` clauses (e.g., `status`, `sku`, `email`) indexed?
    - [ ] Are compound indexes used for multi-column filters (e.g., `user_id` + `status`)?
- [ ] **Soft Deletes:** Are we utilizing GORM’s `DeletedAt` via the `Base` model to prevent accidental data loss?
- [ ] **Query Performance:** 
    - [ ] Have we avoided "N+1" problems by using `.Preload()` for associations?
    - [ ] Are we avoiding `SELECT *` where possible (though GORM defaults to it, be mindful in high-volume queries)?
- [ ] **Audit Trail:** Does every mutation trigger an entry in the `OrderEventLog` or update the `ModifiedBy` / `ModifiedOn` fields?
- [ ] **Migration Safety:** 
    - [ ] Is the `ENABLE_MIGRATION` flag checked before running schema updates?
    - [ ] Are migrations idempotent (safe to run multiple times)?

## 3. Stock & Inventory Management
- [ ] **Atomic Decrement:** Is stock deduction using a raw SQL expression (`stock_quantity - ?`) to prevent race conditions?
- [ ] **Rollback Logic:** Does the system correctly increment stock back if a transaction fails after deduction but before final commit?
- [ ] **Status Validation:** Does every status update that results in `CANCELLED` verify the previous state before replenishing stock?
