FROM golang:latest

ADD . /go/src/github.com/gtalent/dospin/

WORKDIR /go/src/github.com/gtalent/dospin

RUN go get
RUN go build -o dospin

RUN mv dospin /usr/local/bin/dospin


ENTRYPOINT dospin
CMD -log /var/log/dospin.log
