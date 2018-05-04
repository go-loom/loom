FROM golang:1.10.1
COPY . /go/src/github.com/go-loom/loom
WORKDIR /go/src/github.com/go-loom/loom/cmd/loom

RUN go install 

ENTRYPOINT ["loom"]
CMD ["-h"]
