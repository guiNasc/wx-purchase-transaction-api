FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/wx-purchase-api ./cmd

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/wx-purchase-api /usr/local/bin/wx-purchase-api

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/wx-purchase-api"]