# build stage
FROM golang:1.9-alpine AS build-env
RUN apk --no-cache add git
WORKDIR /go/src/github.com/mcadam/kube-plex/
COPY . .
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kube-plex .

# final stage
FROM alpine
COPY --from=build-env /go/src/github.com/mcadam/kube-plex/kube-plex /
ENTRYPOINT /kube-plex
