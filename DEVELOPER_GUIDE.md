# Developer Reference Guide

This document provides technical depth on the architecture, patterns, and design decisions used in the Order Processing System.

## 1. Architectural Patterns

### N-Tier / Clean Architecture
The project is strictly divided into layers to ensure separation of concerns:
- **cmd/src/server:** Orchestration and application lifecycle.
- **Controller:** HTTP entry point, request validation, and response formatting.
- **Service:** Core business logic, identity resolution, and cross-layer coordination.
- **Repository:** Data persistence and GORM abstraction.
- **Models/DTOs:** Pure data structures.

### Interface-Driven Development
Every layer communicates through interfaces. This allows for:
- **Easy Mocking:** (Future-proofing for tests).
- **Loose Coupling:** You can swap PostgreSQL for another DB by simply implementing the Repository interface.
- **Manual Dependency Injection:** Handled in `main.go` to avoid complex DI frameworks.

## 2. Technical Decisions

### Identity Management (UUID vs PK)
- **External:** All API inputs/outputs use `uuid.UUID`. This prevents "ID scraping" and keeps internal DB details private.
- **Internal:** Database relations use `uint` Primary Keys. These are more efficient for indexing and foreign key constraints in PostgreSQL.
- **The Bridge:** The Service layer is responsible for mapping UUIDs to PKs on every request.

### Database Indexing
Optimized for production-scale queries:
- **Unique Indexes:** Email, SKU, and UUID.
- **Standard Indexes:** Order Status (for CRON scanning), UserID (for user history), and Product Name (for search).

### Asynchronous Flow & Reliability
- **RabbitMQ:** Offloads post-order activities.
- **Retries:** Implemented with linear backoff. If a task fails, it's retried up to 3 times.
- **DLX (Dead Letter Exchange):** Messages that fail all retries are moved to a specific queue for manual inspection, ensuring data safety.

## 3. Security & Observability

### Security Standards
- **JWT Best Practices:** 
    - Stateless authentication.
    - Token holds only the `user_uuid`.
    - Every request is verified against the DB record.
    - Refresh Token rotation logic implemented.
- **Middlewares:**
    - **Rate Limiting:** Prevents DDoS by limiting requests per IP.
    - **Security Headers:** Standard protection against XSS and Clickjacking.
- **Encryption:** Passwords stored using `bcrypt` with a cost factor of 10.

### Observability (Zap + RequestID)
- Every request is assigned a unique `X-Request-ID`.
- This ID is injected into the Go `context.Context`.
- The `Zap` logger is configured to pull this ID for every log line, enabling effortless distributed tracing across layers.

## 4. Time Management
- All time-related logic is routed through `internal/utils/time.go`.
- This prevents "time drift" issues and makes the system easier to test across timezones.
- Database uses `timestamptz` for PostgreSQL-native timezone awareness.
