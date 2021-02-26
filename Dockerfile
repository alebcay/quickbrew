FROM golang:1.15-buster AS build

WORKDIR /tmp/golang

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -trimpath

FROM homebrew/brew:latest
COPY --from=build /tmp/golang/quickbrew /usr/bin/quickbrew
