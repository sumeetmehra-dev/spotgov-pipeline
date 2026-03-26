FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/spotgov-pipeline ./cmd/server/main.go

# Runtime
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/bin/spotgov-pipeline .

USER appuser

EXPOSE 8080

CMD ["./spotgov-pipeline"]
