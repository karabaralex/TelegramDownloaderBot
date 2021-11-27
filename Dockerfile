# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

RUN apk add --no-cache git
RUN git clone https://github.com/karabaralex/TelegramDownloaderBot.git
WORKDIR TelegramDownloaderBot
# RUN go mod download
RUN go build -o /bot
CMD [ "/bot" ]
