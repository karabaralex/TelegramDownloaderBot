# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app

COPY . .
RUN go mod download

RUN go build -o /bot

EXPOSE 8080

CMD [ "/bot" ]
