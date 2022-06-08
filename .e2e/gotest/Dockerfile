FROM golang:1.16-alpine as build-env

ADD .e2e/gotest/go.* /go/src/

WORKDIR /go/src/

RUN go mod download

COPY . /go/src/

ARG BASE_URL=http://geolocation-go
ENV BASE_URL ${BASE_URL}
ENV CGO_ENABLED 0

WORKDIR /go/src/.e2e/gotest

CMD [ "sh", "-c", "go test -v ./... -baseurl ${BASE_URL}" ]