FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /bin/server ./cmd/main.go

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /bin/server /app/server
EXPOSE 8080
CMD ["/app/server"]
