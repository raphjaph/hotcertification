FROM golang:alpine AS builder

WORKDIR /go/src/github.com/raphasch/hotcertification
COPY . .
RUN go mod download
RUN go install -ldflags='-s -w' ./...

FROM alpine

RUN apk add iproute2

COPY --from=builder /go/bin/* /usr/bin/