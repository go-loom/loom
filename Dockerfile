FROM golang:1.10.1
COPY . /go/src/github.com/go-loom/loom
WORKDIR /go/src/github.com/go-loom/loom

RUN go install 

ENV LOGXI "*=INF"

ENTRYPOINT ["loom"]
CMD ["-h"]
