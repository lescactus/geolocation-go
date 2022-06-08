FROM golang:1.16 as build-env

ADD go.* /go/src/

WORKDIR /go/src/

RUN go mod download

COPY . /go/src/

ARG BASE_URL=http://geolocation-go
ENV BASE_URL ${BASE_URL}

CMD [ "sh", "-c", "go test -v ./.e2e/gotest -baseurl ${BASE_URL}" ]