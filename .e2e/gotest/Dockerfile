FROM golang:1.16-alpine as build-env

ADD .e2e/gotest/go.* /go/src/

WORKDIR /go/src/

RUN go mod download

COPY . /go/src/

ARG BASE_URL=http://geolocation-go
ARG REDIS_CONNECTION_STRING=redis://redis:6379

ENV BASE_URL ${BASE_URL}
ENV REDIS_CONNECTION_STRING ${REDIS_CONNECTION_STRING}
ENV CGO_ENABLED 0

WORKDIR /go/src/.e2e/gotest

CMD [ "sh", "-c", "go test -v ./... -baseurl ${BASE_URL} -redisconnstr ${REDIS_CONNECTION_STRING}" ]