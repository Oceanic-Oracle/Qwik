FROM golang:alpine AS build

RUN mkdir /app

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal ./internal
COPY cmd/mailer ./cmd/mailer

RUN go build -o /app/bin ./cmd/mailer/app.go

FROM alpine:3.21

COPY --from=build /app/bin /app/bin

COPY internal/config/.env /internal/config/.env

EXPOSE 8081

ENTRYPOINT ["/app/bin"]