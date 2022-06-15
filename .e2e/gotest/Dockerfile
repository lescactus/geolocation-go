FROM golang:1.16-alpine as build-env

ADD .e2e/gotest/go.* /go/src/

WORKDIR /go/src/

RUN go mod download

COPY . /go/src/

ARG BASE_URL=http://geolocation-go
ARG REDIS_CONNECTION_STRING=redis://redis:6379
ENV CGO_ENABLED 0

WORKDIR /go/src/.e2e/gotest

RUN go test -c -o e2e

FROM alpine:3

ARG BASE_URL=http://geolocation-go
ARG REDIS_CONNECTION_STRING=redis://redis:6379
ENV BASE_URL ${BASE_URL}
ENV REDIS_CONNECTION_STRING ${REDIS_CONNECTION_STRING}

COPY --from=build-env /go/src/.e2e/gotest/e2e /e2e

CMD [ "sh", "-c", "/e2e -test.v -test.parallel 1 -baseurl ${BASE_URL} -metricsurl ${BASE_URL}/metrics -redisconnstr ${REDIS_CONNECTION_STRING} ./..." ]