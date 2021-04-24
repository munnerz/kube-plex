FROM golang:1.16 AS builder

WORKDIR /go/src/app
COPY . .

ENV CGO_ENABLED=0

RUN go build -o kube-plex main.go
RUN go build ./cmd/...

FROM alpine:3.13

COPY --from=builder /go/src/app/kube-plex /go/src/app/transcode-launcher /