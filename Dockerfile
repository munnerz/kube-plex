FROM golang:latest

WORKDIR /go/src/github.com/munnerz/kube-plex

COPY . .

# RUN go test -cpu=2 -race -v -covermode=atomic $(go list ./... | grep -v /vendor/)
RUN CGO_ENABLED=0 GOOS=linux go build github.com/munnerz/kube-plex/cmd/kubeplex

FROM alpine:3.6

COPY --from=0 /go/src/github.com/munnerz/kube-plex/kubeplex /kubeplex
