FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/gateway
RUN go build -v -o gateway_service

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/cmd/gateway/gateway_service .
COPY --from=builder /app/cmd/docs/open-api.yaml .
COPY --from=builder /app/cmd/docs/redoc-static.html .

EXPOSE 8080

CMD ["./gateway_service"]
