FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o fynt-api ./cmd/api

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Jakarta

COPY --from=builder /app/fynt-api .
COPY --from=builder /app/infrastructure/database/migrations ./infrastructure/database/migrations

EXPOSE 8080

CMD ["./fynt-api"]