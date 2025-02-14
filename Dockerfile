FROM golang:1.23.6 AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOSE=linux go build -o /app/main .

FROM alpine:3.19

RUN apk add --no-cache bash

WORKDIR /app

COPY --from=builder /app/main /app/main

EXPOSE 8080

CMD ["/app/main"]