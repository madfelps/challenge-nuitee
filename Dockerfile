FROM golang:1.24.3 AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

ENV CGO_ENABLED=0
RUN go build -o main -ldflags "-s -w" ./cmd/api

FROM golang:alpine3.22

WORKDIR /app

COPY --from=builder /build/main /app/main

ENTRYPOINT ["/app/main"]