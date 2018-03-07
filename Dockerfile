# build stage
FROM golang:alpine AS build-env
COPY . /go/src/seng468/WebServer
RUN apk add --no-cache git \
    && go get github.com/garyburd/redigo/redis \
    && go get github.com/shopspring/decimal \
    && cd /go/src/seng468/WebServer \
    && go build -o webserve

# final stage
FROM alpine

ARG webaddr
ENV webaddr=$webaddr
ARG webport
ENV webport=$webport
ARG auditaddr
ENV auditaddr=$auditaddr
ARG auditport
ENV auditport=$auditport

WORKDIR /app
COPY --from=build-env /go/src/seng468/WebServer/webserve /app/
EXPOSE 44455-44459
ENTRYPOINT ./webserve