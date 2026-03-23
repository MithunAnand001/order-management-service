# Go Code Review Checklist

Use this checklist during Pull Requests to ensure code quality, architectural consistency, and adherence to the project's steering principles.

## 1. General Go Idioms
- [ ] **Naming:** Does it follow the Steering Doc? (camelCase variables, PascalCase exported functions, short receivers).
- [ ] **Formatting:** Is the code formatted with `gofmt` or `goimports`?
- [ ] **Explicit over Implicit:** Are there any "magic" values? Use constants instead.
- [ ] **Package Design:** Does the package name "stutter" with its exports? (e.g., `repository.UserRepository` is better than `repository.Repository`).

## 2. Architectural Integrity (N-Tier)
- [ ] **Interfaces:** Is the interface defined in the implementation file? Does the constructor return the Interface?
- [ ] **Dependency Injection:** Are dependencies passed through constructors? (No global state access).
- [ ] **Layer Boundaries:** 
    - Does the Controller handle HTTP only?
    - Does the Service contain all business logic and identity mapping?
    - Does the Repository handle GORM/DB specific logic only?
- [ ] **DTO Usage:** Are database models (`internal/models`) leaking into the API response? (Always map to DTOs).

## 3. Identity & Type Safety
- [ ] **UUID vs PK:** Are external parameters strictly `uuid.UUID`? Is the resolution to `uint` PK happening only in the Service layer?
- [ ] **Type Consistency:** Are we avoiding `interface{}` where a concrete type or generic `[T any]` could be used?

## 4. Error Handling
- [ ] **Custom Errors:** Are all returned errors of type `*dto.AppError`?
- [ ] **Error Wrapping:** Are internal errors wrapped with enough context?
- [ ] **Early Return:** Is the "Happy Path" on the left? (Avoid deep nesting of if-statements).
- [ ] **No Silent Failures:** Are all returned errors checked and logged?

## 5. Observability & Context
- [ ] **Context Propagation:** Is `context.Context` the first argument of all IO functions?
- [ ] **Tracing:** Does every log entry include the `request_id` from the context?
- [ ] **Log Levels:** Is `Info` used for lifecycles, `Warn` for business rejections (e.g., Auth), and `Error` for system failures?

## 6. Security & Performance
- [ ] **SQL Injection:** Are all queries using GORM's parameterized methods? (No string concatenation in queries).
- [ ] **Sensitive Data:** Is there any logging of passwords, tokens, or PII?
- [ ] **Transactions:** Are multi-table writes (e.g., Order + Stock) wrapped in `db.Transaction`?
- [ ] **Concurrency:** Are shared resources (like the SMTP Pool) protected by Mutexes or Channels?
- [ ] **Time Management:** Are we using `internal/utils/time.go` instead of direct `time.Now()`?

## 7. Infrastructure (RabbitMQ & SMTP)
- [ ] **Acks/Nacks:** Does the RabbitMQ consumer correctly Ack on success and Nack on failure?
- [ ] **Retry Logic:** Is the retry counter correctly incremented in headers?
- [ ] **Connection Safety:** Are we using the Connection Pool for SMTP tasks?
