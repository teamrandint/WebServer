FROM golang:latest
WORKDIR /go/src/app
ADD . .
RUN go build WebServer.go
CMD ./WebServer '' 5000