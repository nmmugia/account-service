# Dockerfile
FROM golang:1.22 AS builder

WORKDIR /app

COPY . . 

RUN go mod download

COPY src/ ./

RUN go test -v -coverprofile=coverage.out -coverpkg=./... ./test/unit/...

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/account-service .

FROM alpine:latest

RUN apk add --no-cache curl

WORKDIR /root

COPY --from=builder /go/bin/account-service .
COPY --from=builder /app/migration .
COPY --from=builder /app/.env .
COPY --from=builder /app/src/database/migrations /migrations

COPY ./entrypoint.sh .
RUN chmod +x ./entrypoint.sh

EXPOSE 3000
ENTRYPOINT ["./entrypoint.sh"]