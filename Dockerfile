FROM golang:1.10.1
COPY . /go/src/gopkg.in/loom.v1
WORKDIR /go/src/gopkg.in/loom.v1

ENV GOPATH /go/src/gopkg.in/loom.v1/Godeps/_workspace:$GOPATH
RUN go install 

ENV LOGXI "*=INF"

ENTRYPOINT ["loom.v1"]
CMD ["-h"]
