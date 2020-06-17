#Build Stage
FROM golang:1.14.2-alpine as builder
ENV GOPROXY=https://gocenter.io 

#Add packages for dependency management
RUN apk add dep
RUN apk add git

#Copy project files and get dependencies
COPY . /go/src/github.com/munnerz/kube-plex
WORKDIR /go/src/github.com/munnerz/kube-plex
RUN set -x && \
	go get . && \
	dep ensure -v

#Run Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o kube-plex_linux_amd64 /go/src/github.com/munnerz/kube-plex
RUN ls /go/src/github.com/munnerz/kube-plex

#Copy into clean container
FROM alpine:3.6

COPY --from=builder /go/src/github.com/munnerz/kube-plex/kube-plex_linux_amd64 /kube-plex
