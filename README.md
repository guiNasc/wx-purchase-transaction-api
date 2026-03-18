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
- Docker and Docker Compose (for local PostgreSQL)

## Environment Variables

Application:

- `PORT` (default: `8080`)
- `LOG_LEVEL` (`debug`, `info`, `warn`, `error`; default `info`)
- `LOG_FORMAT` (`text`, `json`; default `text`)

## Local Setup

1. Start PostgreSQL

```bash
docker compose up -d
```

2. Run the API

```bash
go run ./cmd
```

3. Optional: custom port

```bash
PORT=3000 go run ./cmd
```

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
