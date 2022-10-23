# syntax=docker/dockerfile:1

FROM golang:1.18-alpine

WORKDIR /app

COPY . /app

RUN go mod tidy
RUN go mod download
RUN go build -o /docker-gs-ping

CMD [ "/docker-gs-ping" ]