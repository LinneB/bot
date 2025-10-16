# Build
FROM golang:1.22.6-alpine3.20 AS build
WORKDIR /app
COPY . .
RUN go build -o /app/bot cmd/bot.go

# Run
FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/bot /app/bot
ENTRYPOINT ["./bot"]
