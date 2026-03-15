# wx-purchase-api

Simple Go API for currency conversion.

## Endpoints

### GET /health
Returns service status.

Example response:

```json
{
  "status": "ok",
  "service": "wx-purchase-api",
  "timestamp": "2026-03-15T12:00:00Z"
}
```

### POST /transactions
Store a new purchase transaction.

Expected body request:
```json
{
 "description": "test-transaction",
 "transactionDate": "12/01/2019",
 "amount": 2323.12
}
```

Example response:

```json
{
  "message": "Transaction created successfully"
}
```

## Run

```bash
go run .
```

The API starts on port `8080` by default.
To override, set:

```bash
PORT=3000 go run .
```
