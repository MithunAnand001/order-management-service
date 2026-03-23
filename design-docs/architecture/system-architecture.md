# System Architecture

The Order Processing System follows a strict N-tier, interface-driven architecture in Go, emphasizing separation of concerns, observability, and asynchronous reliability.

## 1. High-Level Component Diagram

```mermaid
graph TD
    subgraph "Client Layer"
        Client[Mobile/Web/Postman]
    end

    subgraph "Application Layer (Go Service)"
        subgraph "HTTP/Middleware"
            Router[gorilla/mux]
            MW_ID[RequestID MW]
            MW_Log[Logger MW - Zap]
            MW_Auth[Auth MW - JWT]
            MW_Sec[Security MW - Headers/Limit]
        end

        subgraph "Logic Layer"
            Ctrl[Controllers]
            Svc[Services]
            Cron[robfig/cron Job]
        end

        subgraph "Data & Infra Access"
            Repo[Repositories - GORM]
            Broker[Message Broker - RabbitMQ]
        end
    end

    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        RMQ[(RabbitMQ)]
    end

    %% Flows
    Client --> Router
    Router --> MW_ID --> MW_Log --> MW_Sec --> MW_Auth
    MW_Auth --> Ctrl
    Ctrl --> Svc
    Svc --> Repo
    Svc --> Broker
    Cron --> Svc
    Repo --> DB
    Broker --> RMQ
```

## 2. Layer Responsibilities

### Controller Layer
- Handles HTTP request parsing using `DecodeAndValidate`.
- Maps incoming data to DTOs.
- Calls the appropriate Service method.
- Formats standardized Spenza-style responses (`ApiResponse`).

### Service Layer (Business Logic)
- **Identity Resolution:** Maps external UUIDs to internal PKs.
- **Validation:** Enforces business rules (e.g., "Only PENDING orders can be cancelled").
- **Orchestration:** Coordinates between Repositories and the Message Broker.
- **Asynchronous Flow:** Publishes events to RabbitMQ after critical state changes.

### Repository Layer (Data Access)
- Encapsulates GORM operations.
- Performs database lookups and persistence.
- Maps database-specific errors to custom domain `AppError` types.

### Infrastructure Layer
- **PostgreSQL:** Primary relational storage with optimized indexing.
- **RabbitMQ:** Handles post-creation activities with a robust **Retry & DLX** mechanism.
- **Cron Worker:** Operates as a background goroutine to transition order statuses automatically.

## 3. Asynchronous Event & Retry Flow

```mermaid
sequenceDiagram
    participant S as OrderService
    participant E as Order Exchange (Topic)
    participant Q as Order Queue
    participant C as Order Consumer
    participant DLX as Retry/DLX Queue

    S->>E: Publish 'order.created'
    E->>Q: Route to Queue
    Q->>C: Deliver Message
    Note over C: processMessage()
    alt Processing Success
        C->>Q: Ack
    else Processing Failure (Attempt < 3)
        C->>E: Republish with x-retry-count++
        Note right of C: Exponential Backoff Delay
    else Max Retries Reached
        C->>Q: Nack(requeue=false)
        Q->>DLX: Move to DLX
    end
```

## 4. Security & Observability
- **Tracing:** Every request is assigned a `RequestID`, which is propagated through the `context.Context` and included in every log line.
- **Logging:** Structured logging via `uber-go/zap` ensures all Start/End/Error events are searchable and actionable.
- **Authentication:** Stateless JWT flow where the token holds the UUID, and the middleware verifies the user against the database.
