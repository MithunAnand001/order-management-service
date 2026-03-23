# Project Folder Structure

This project follows a strictly organized, N-tier architecture designed for scalability, maintainability, and clear separation of concerns in Go.

## 1. Root Directory
- **`cmd/src/server/`**: Contains the application entry point (`main.go`). This is where the application lifecycle is managed, including initialization, dependency injection, and graceful shutdown.
- **`design-docs/`**: Centralized location for all architectural blueprints, API specifications, and database diagrams.
- **`internal/`**: Private application code. Following Go best practices, code in this directory cannot be imported by external projects, ensuring the integrity of the service boundary.
- **`.env`**: Environment configuration (secrets, database URLs, infrastructure ports).
- **`docker-compose.yml`**: Orchestration for local infrastructure (RabbitMQ, MailHog).
- **`Makefile`**: Development shortcuts for common tasks (build, run, seed).

## 2. Internal Directory (`internal/`)

### `config/`
- **Purpose:** Handles configuration loading and infrastructure initialization.
- **Key Files:** 
  - `config.go`: Loads environment variables into a typed struct.
  - `db.go`: Encapsulates PostgreSQL connection logic.

### `controller/`
- **Purpose:** The HTTP transport layer.
- **Responsibility:** Parses request parameters, validates input using DTOs, calls the service layer, and formats the standardized `ApiResponse`.

### `dto/` (Data Transfer Objects)
- **Purpose:** Defines the "Contract" between the API and the outside world.
- **Key Types:**
  - Request/Response structs with `json` and `validate` tags.
  - `AppError`: The custom domain error type used across all layers.
  - `ApiResponse`: The standardized Spenza-style success/error wrapper.

### `middleware/`
- **Purpose:** Cross-cutting concerns applied to HTTP requests.
- **Features:** Tracing (RequestID), Structured Logging (Zap), Security Headers, Rate Limiting, and JWT Authentication.

### `models/`
- **Purpose:** Domain entities and database schema definitions.
- **Responsibility:** Pure data structures with GORM tags. Includes the `Base` model for universal audit tracking.

### `repository/`
- **Purpose:** Data Access Object (DAO) layer.
- **Responsibility:** Performs all database operations (CRUD). It is the only layer that interacts with the GORM `Instance`. Maps database errors to domain `AppError` types.

### `routes/`
- **Purpose:** Centralized API routing.
- **Responsibility:** Defines the URL paths and attaches the correct controllers and middlewares to them.

### `service/`
- **Purpose:** The "Brain" of the application (Business Logic).
- **Responsibility:** Orchestrates complex workflows, handles role-based access control (RBAC), manages stock deduction/replenishment, and triggers asynchronous events.
- **`service/rabbitmq/`**: Contains the implementations for message publishing and background consumption.

### `utils/`
- **Purpose:** Reusable helper functions.
- **Key Files:**
  - `validator.go`: Centralized JSON decoding and validation logic.
  - `time.go`: Standardized time management.
  - `response.go`: Helpers for constructing API responses.

## 3. Communication Flow
`Client -> Middleware -> Controller -> Service -> Repository -> Database/Infrastructure`
