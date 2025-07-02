FROM golang:1.24.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o rest-server ./cmd/rest 
RUN CGO_ENABLED=0 GOOS=linux go build -o grpc-server ./cmd/grpc
RUN CGO_ENABLED=0 GOOS=linux go build -o kafka-consumer ./cmd/kafka

FROM alpine:latest

RUN apk --no-cache add tini ca-certificates

WORKDIR /app

COPY --from=builder /app/rest-server .
COPY --from=builder /app/grpc-server .
COPY --from=builder /app/kafka-consumer .

COPY config/ ./config/

RUN mkdir -p /app/logs

EXPOSE 8001
EXPOSE 50051

ENTRYPOINT ["/sbin/tini", "--"]

CMD ["sh", "-c", "./rest-server & ./grpc-server & ./kafka-consumer & wait -n"]