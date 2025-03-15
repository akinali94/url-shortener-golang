FROM golang:1.24
#FROM golang:1.23-alpine as builder


WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /url-shortener
#RUN CGO_ENABLED=0 GOOS=linux go build -o ./cmd/main.go


CMD ["/url-shortener"]

