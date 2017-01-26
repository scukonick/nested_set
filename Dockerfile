from golang:latest

RUN go get bitbucket.org/liamstask/goose/cmd/goose
RUN go get github.com/lib/pq
WORKDIR /go/src/app
COPY . /go/src/app
RUN go get -d -v 

ENTRYPOINT ["/bin/bash", "tests.sh"]
