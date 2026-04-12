# wx-purchase-api

REST API for purchase transactions and exchange conversion using Go, Gin, PostgreSQL, and Treasury exchange rates.

## Overview

This service allows you to:

- Create and list purchase transactions.
- Retrieve a transaction by id.
- Convert a transaction amount to another currency using Treasury exchange data for a 6-month lookback window.

## Tech Stack

- Go 1.25+
- Gin (HTTP server)
- PostgreSQL 16
- Wire (dependency injection)
- slog (structured logging)

## Architecture

Project layers:

- controller: HTTP handlers and request/response mapping.
- usecase: business rules and orchestration.
- repository: database persistence.
- infra: outbound exchange rate gateway.
- database: PostgreSQL connection setup.
- cmd: application bootstrap and DI wiring.

Dependency injection is generated with Google Wire from `cmd/wire.go` into `cmd/wire_gen.go`.

## Requirements

- Go installed
- Docker and Docker Compose

## Environment Variables

Application:

- `PORT` (default: `8080`)
- `LOG_LEVEL` (`debug`, `info`, `warn`, `error`; default `info`)
- `LOG_FORMAT` (`text`, `json`; default `text`)

Database:

- `DB_HOST` (default: `localhost`)
- `DB_PORT` (default: `5432`)
- `DB_USER` (default: `postgres`)
- `DB_PASSWORD` (default: `postgres`)
- `DB_NAME` (default: `wx_purchase_transactions`)
- `DB_SSLMODE` (default: `disable`)
- `DB_MAX_OPEN_CONNS` (default: `25`)
- `DB_MAX_IDLE_CONNS` (default: `25`, capped to max open conns)
- `DB_CONN_MAX_LIFETIME_SECONDS` (default: `300`)
- `DB_CONN_MAX_IDLE_TIME_SECONDS` (default: `120`)
- `DB_CONNECT_TIMEOUT_SECONDS` (default: `5`)
- `DB_CONNECT_MAX_RETRIES` (default: `8`)
- `DB_CONNECT_BACKOFF_INITIAL_MS` (default: `500`)
- `DB_CONNECT_BACKOFF_MAX_MS` (default: `5000`)

## Run With Docker

This repository includes a multi-stage `Dockerfile` and a `docker-compose.yml` that run both PostgreSQL and the API in containers.

1. Create your environment file

```bash
cp .env.example .env
```

2. Start the full stack

```bash
docker compose up --build -d
```

3. Call the API

```bash
curl http://localhost:8080/health
```

Only the API port is published to the host. PostgreSQL stays on the internal Docker network and is reachable by the API as `postgres:5432`.

To stop everything:

```bash
docker compose down
```

To remove the database volume too:

```bash
docker compose down -v
```

## Local Setup Without Docker For The API

1. Start PostgreSQL

```bash
docker compose up -d postgres
```

2. Run the API

```bash
go run ./cmd
```

3. Optional: custom port

```bash
PORT=3000 go run ./cmd
```

When running the API outside Docker, keep `DB_HOST=localhost`. When running the full stack in Docker Compose, the API service automatically uses `DB_HOST=postgres`.

## Dependency Injection (Wire)

Regenerate wiring code when providers change:

```bash
go generate ./cmd
```

Or run Wire directly in the cmd package.

## API Endpoints

Base URL:

```text
http://localhost:8080
```

### Health Check

`GET /health`

Response `200`:

```json
{
  "status": "ok",
  "service": "wx-purchase-api",
  "timestamp": "2026-03-18T12:00:00Z"
}
```

### Create Transaction

`POST /transactions`

Request body:

```json
{
  "description": "Laptop purchase",
  "transactionDate": "2019-07-25",
  "amount": 2323.12
}
```

Validation rules:

- `description` max length: 50
- `amount` must be positive
- `transactionDate` format: `YYYY-MM-DD`

Response `201`:

```json
{
  "message": "Transaction created successfully",
  "transaction": {
    "id": "1",
    "description": "Laptop purchase",
    "transactionDate": "2019-07-25",
    "amount": 2323.12
  }
}
```

### List Transactions

`GET /transactions`

Response `200`:

```json
[
  {
    "id": "1",
    "description": "Laptop purchase",
    "transactionDate": "2019-07-25",
    "amount": 2323.12
  },
  {
    "id": "2",
    "description": "Phone purchase",
    "transactionDate": "2019-07-20",
    "amount": 1200.5
  }
]
```

### Get Transaction By Id

`GET /transactions/:id`

Example:

```text
GET /transactions/1
```

Response `200`:

```json
{
  "id": "1",
  "description": "Laptop purchase",
  "transactionDate": "2019-07-25",
  "amount": 2323.12
}
```

### Get Transaction Exchange

`GET /transactions/:id/exchange/:currency`

Example:

```text
GET /transactions/1/exchange/Brazil-Real
```

Behavior:

- Loads transaction by id.
- Uses transaction date as `toDate`.
- Uses `toDate - 6 months` as `fromDate`.
- Queries Treasury rates and applies the first record returned.

Response `200`:

```json
{
  "id": "1",
  "description": "Laptop purchase",
  "transactionDate": "2019-07-25",
  "amount": 2323.12,
  "currency": "Brazil-Real",
  "exchangeRate": 3.812,
  "convertedAmount": 8855.33
}
```

## Error Response Format

Endpoints return errors in this shape:

```json
{
  "error": "Failed to retrieve transaction"
}
```

## Run Tests

```bash
go test ./...
```
