# Build
FROM golang:1.25.0-alpine3.22 AS build
WORKDIR /app
COPY . .
RUN go build -o /app/bot cmd/bot.go

# Run
FROM alpine:3.22
WORKDIR /app
COPY --from=build /app/bot /app/bot
ENTRYPOINT ["./bot"]
