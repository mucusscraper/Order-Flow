FROM golang:1.25-alpine AS builder
WORKDIR /app
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/api cmd/api/main.go
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /bin/api /app/api
COPY --from=builder /app/migrations /app/migrations
EXPOSE 8080
CMD ["./api"]