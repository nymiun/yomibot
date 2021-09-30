FROM golang:latest AS builder

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOPROXY=https://proxy.golang.org go build -o app

FROM alpine:latest

RUN apk --no-cache add ca-certificates mailcap && addgroup -S app && adduser -S app -G app

ENV DISCORD_BOT_TOKEN=""
ENV DISCORD_BOT_PREFIX="\$"
ENV SPOTIFY_CLIENT_ID=""
ENV SPOTIFY_CLIENT_SECRET=""
ENV DB_DSN=""
ENV LAVA_HOSTNAME=""
ENV LAVA_PORT=""
ENV LAVA_PASSWORD=""
ENV LAVA_RESUME_KEY=""

USER app
WORKDIR /app
COPY --from=builder /app/app .
ENTRYPOINT [ "./app" ]