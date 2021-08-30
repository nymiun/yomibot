FROM golang:latest AS builder

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOPROXY=https://proxy.golang.org go build -o app

FROM alpine:latest

RUN apk --no-cache add ca-certificates mailcap && addgroup -S app && adduser -S app -G app

ARG DISCORD_TOKEN
ARG DISCORD_PREFIX="!"
ARG SPOTIFY_CLIENT_ID
ARG SPOTIFY_CLIENT_SECRET
ARG DB_DSN

ENV DISCORD_BOT_TOKEN=${DISCORD_TOKEN}
ENV DISCORD_BOT_PREFIX=${DISCORD_PREFIX}
ENV SPOTIFY_CLIENT_ID=${SPOTIFY_CLIENT_ID}
ENV SPOTIFY_CLIENT_SECRET=${SPOTIFY_CLIENT_SECRET}
ENV DB_DSN=${DB_DSN}

USER app
WORKDIR /app
COPY --from=builder /app/app .
ENTRYPOINT [ "./app" ]