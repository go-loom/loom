FROM golang:1.4.2
COPY . /go/src/gopkg.in/loom.v1
WORKDIR /go/src/gopkg.in/loom.v1

ENV GOPATH /go/src/gopkg.in/loom.v1/Godeps/_workspace:$GOPATH
RUN go install -v 

ENV LOGXI "*=INF"

ENTRYPOINT ["loom.v1"]
CMD ["loom.v1"]
