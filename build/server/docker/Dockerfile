FROM golang:1.23.6-bookworm AS build

WORKDIR /build

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN go build -o /build/server /build/cmd/server/main.go

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=build /build/server /app/server

ENTRYPOINT ["/app/server"]