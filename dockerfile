
FROM golang:1.26.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o fynt-api ./cmd/api

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add tzdata ca-certificates

ENV TZ=Asia/Jakarta

COPY --from=builder /app/fynt-api ./fynt-api

EXPOSE 8080

CMD ["./fynt-api"]
