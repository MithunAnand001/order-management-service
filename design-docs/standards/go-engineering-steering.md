# Go Engineering Steering Document

This document defines the strict engineering standards and idiomatic patterns for the Order Processing System.

## 1. Naming Conventions
- **Packages:** Short, lowercase, single word. Avoid underscores or camelCase (e.g., `internal/repository`).
- **Variables:** `camelCase`. Use short, descriptive names. For short-lived scopes, use single letters (e.g., `r` for repository, `ctx` for context).
- **Functions:** `PascalCase` for exported, `camelCase` for private.
- **JSON Keys:** Strictly `snake_case` (e.g., `order_uuid`, `created_on`).
- **Receivers:** Use 1-3 letter abbreviations of the struct name (e.g., `func (s *orderSer) ...`).

## 2. API & Routing Standards
- **Paths:** Always lowercase. Use plural nouns for resources (e.g., `/users`, `/orders`).
- **Methods:**
  - `GET`: Retrieve data. No side effects.
  - `POST`: Create new resources.
  - `PUT`: Update existing resources or perform semantic actions (e.g., `/cancel`).
- **Versioning:** All routes must be prefixed with `/api/v1`.
- **Responses:** All responses must wrap data/errors in the standardized `ApiResponse[T]` struct.

## 3. Architectural Rules
- **Interface-First:** All layers (Repo, Service, Controller) must have an interface defined right above the concrete struct in the same file.
- **Constructors:** `NewX(...)` functions must return the **Interface**, never the concrete pointer.
- **Decoupling:** Layers must only communicate via interfaces. No circular dependencies (use interface segregation to break cycles).
- **Identity:** External layers (API/DTO) use `uuid.UUID`. Internal layers (Repository/DB) use `uint` PKs. Resolution happens in the Service layer.

## 4. Error Handling Logic
- **Domain Errors:** Use the custom `dto.AppError` for all cross-layer communication.
- **Mapping:** Repositories must map GORM/DB errors to `dto.ErrCode...`.
- **Early Return:** Handle errors immediately. Return `nil, appErr` to bubble up issues to the controller for formatting.

## 5. Observability & Context
- **Context:** `context.Context` must be the first parameter of every IO-bound or business logic function.
- **Tracing:** Every log entry **MUST** include the `request_id` retrieved via `utils.GetRequestID(ctx)`.
- **Logging:** Use `zap.Logger`. Format: `"Action Package.Method", zap.String("request_id", reqID), ...`

## 6. Infrastructure & Safety
- **Time:** Strictly use `internal/utils/time.go` helpers. No direct `time.Now()` calls.
- **Transactions:** Any operation affecting multiple tables (e.g., Order + Stock) **MUST** be wrapped in a GORM Transaction at the Repository level.
- **RabbitMQ:** All asynchronous side effects must implement a **Retry + DLX** pattern.
- **SMTP:** Use the **Connection Pool** pattern for parallel email delivery.
