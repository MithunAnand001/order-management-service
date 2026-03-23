# Order Processing System

A robust backend service built with Go for managing e-commerce orders.

## Features
- **Create Order:** Support for multiple items with automatic total calculation.
- **Order Retrieval:** Fetch order details by UUID.
- **List Orders:** Retrieve all orders with optional status filtering.
- **Background Job:** Automatically updates `PENDING` orders to `PROCESSING` every 5 minutes.
- **Cancel Order:** Allows cancellation ONLY if the order is in `PENDING` status.
- **Authentication:** JWT-based authentication with Access and Refresh tokens.
- **Security:** Rate limiting, security headers, and structured logging.

## Tech Stack
- **Language:** Go (Golang)
- **Framework:** [gorilla/mux](https://github.com/gorilla/mux)
- **ORM:** [GORM](https://gorm.io/)
- **Database:** PostgreSQL
- **Background Tasks:** [robfig/cron](https://github.com/robfig/cron)
- **Validation:** [go-playground/validator](https://github.com/go-playground/validator)
- **Message Broker:** RabbitMQ (for asynchronous activities)
- **Logging:** [uber-go/zap](https://github.com/uber-go/zap)

## Project Structure
```
order-management-service/
├── cmd/
│   └── src/
│       └── server/         # main.go (Entry point)
├── internal/
│   ├── config/             # Configuration loading
│   ├── controller/         # HTTP handlers
│   ├── dto/                # Data Transfer Objects & Custom Errors
│   ├── jobs/               # Background cron jobs
│   ├── middleware/         # HTTP middlewares (Auth, Logging, Security)
│   ├── models/             # Database models (GORM)
│   ├── rabbitmq/           # RabbitMQ publisher/subscriber
│   ├── repository/         # Database access layer (DAO)
│   ├── routes/             # Route definitions
│   ├── service/            # Business logic
│   └── utils/              # General utilities (Time, Validator, Response)
├── .env                    # Environment configuration
└── README.md
```

## How to Run

### 1. Install Dependencies
```bash
go mod tidy
```

### 2. Configure Environment
Create or update the `.env` file with your database and RabbitMQ credentials.

### 3. Run the Server
```bash
go run cmd/src/server/main.go
```
The server starts on `http://localhost:8080` by default.

## API Endpoints (v1)

### Authentication
- `POST /api/v1/users/register`: Create a new user account.
- `POST /api/v1/users/login`: Authenticate and receive tokens.
- `POST /api/v1/users/refresh-token`: Refresh your access token.

### Products
- `GET /api/v1/products`: List all active products (supports `search`, `limit`, `offset`).

### Orders
- `POST /api/v1/orders`: Place a new order.
- `GET /api/v1/orders`: List user's orders (Filter: `?status=PENDING`).
- `GET /api/v1/orders/{uuid}`: Get order details.
- `PUT /api/v1/orders/{uuid}/cancel`: Cancel an order (if PENDING).

## Technical Notes
1. **Security:** JWTs only store the user's UUID. A database lookup is performed on every authenticated request to verify the user's status and retrieve their internal ID.
2. **Observability:** Structured logging with `zap` provides detailed insights into every request and function execution.
