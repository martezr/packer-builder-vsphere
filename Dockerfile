FROM golang:1.9.3

RUN go get github.com/vmware/govmomi/govc

ENV GOPATH=$HOME/src/go PATH=$PATH:$GOPATH/bin
