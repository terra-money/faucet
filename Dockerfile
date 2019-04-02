FROM golang:latest

RUN mkdir -p "$GOPATH/src/github.com/terra-project/faucet"
WORKDIR $GOPATH/src/github.com/terra-project/faucet
COPY faucet.go .

RUN go install -v ./...

CMD ["faucet"]
