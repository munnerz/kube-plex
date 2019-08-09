# build stage
FROM golang:alpine AS build-env
RUN apk --no-cache add git
ADD . /src
RUN cd /src && go get && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o kube-plex .

# final stage
FROM scratch
COPY --from=build-env /src/kube-plex /
ENTRYPOINT /kube-plex
