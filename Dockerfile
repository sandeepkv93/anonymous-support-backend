FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Install buf for proto generation
RUN go install github.com/bufbuild/buf/cmd/buf@latest

COPY . .

# Generate proto files
RUN buf generate && mv gen/proto/* gen/ && rmdir gen/proto

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/.env.example .env

EXPOSE 8080

CMD ["./server"]
